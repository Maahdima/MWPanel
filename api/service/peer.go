package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/shopclient"
	"github.com/maahdima/mwp/api/utils"
	"github.com/maahdima/mwp/api/utils/timehelper"
	"github.com/maahdima/mwp/api/utils/wireguard"
)

type WgPeer struct {
	tenantID        *uint
	serverID        uint
	db              *gorm.DB
	mikrotikAdaptor *mikrotik.Adaptor
	scheduler       *Scheduler
	queue           *Queue
	configGenerator *ConfigGenerator
	qrCodeGenerator *QRCodeGenerator
	shopLink        *shopclient.Client
	logger          *zap.Logger
}

func NewWGPeer(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor, scheduler *Scheduler, queue *Queue, configGenerator *ConfigGenerator, qrCodeGenerator *QRCodeGenerator) *WgPeer {
	return &WgPeer{
		db:              db,
		mikrotikAdaptor: mikrotikAdaptor,
		scheduler:       scheduler,
		queue:           queue,
		configGenerator: configGenerator,
		qrCodeGenerator: qrCodeGenerator,
		logger:          zap.L().Named("WgPeerService"),
	}
}

func (w *WgPeer) TogglePeerStatus(id uint) error {
	var initial model.Peer
	if err := w.db.First(&initial, id).Error; err != nil {
		return err
	}
	if initial.ResellerID != nil {
		if _, err := NewBilling(w.db).ReconcileAccount(*initial.ResellerID); err != nil {
			return err
		}
	}
	update := func(tx *gorm.DB) error {
		lock := tx.Model(&model.Peer{}).Where("id = ?", id).UpdateColumn("restriction_revision", gorm.Expr("restriction_revision"))
		if lock.Error != nil {
			return lock.Error
		}
		if lock.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		var peer model.Peer
		if err := tx.First(&peer, id).Error; err != nil {
			return err
		}
		if peer.PurchaseID != nil || !uintPointersEqual(peer.ResellerID, initial.ResellerID) {
			return ErrPendingPeerOperation
		}
		disabled := !peer.Disabled
		if !disabled {
			blocked, err := peerOtherRestriction(tx, peer, time.Now())
			if err != nil {
				return err
			}
			if blocked {
				return ErrPeerRestricted
			}
			if peer.ResellerID != nil {
				candidate := peer
				candidate.Disabled = false
				if err := ValidatePeerResource(tx, *peer.ResellerID, candidate, peer.ID); err != nil {
					return err
				}
			}
		}
		return tx.Model(&peer).Updates(map[string]any{"manual_disabled": disabled, "disabled": disabled, "restrictions_initialized": true, "restriction_revision": gorm.Expr("restriction_revision + 1"), "sync_next_attempt_at": 0}).Error
	}
	var err error
	if initial.ResellerID != nil {
		err = NewBilling(w.db).accountTransaction(*initial.ResellerID, func(tx *gorm.DB, _ *model.BillingAccount) error { return update(tx) })
	} else {
		err = w.db.Transaction(update)
	}
	if err != nil {
		return err
	}
	if err := NewPeerReconciler(w.db, w.mikrotikAdaptor, NewBilling(w.db)).ReconcilePeer(id); err != nil {
		return fmt.Errorf("%w: %v", ErrRouterSyncPending, err)
	}
	return nil
}

func (w *WgPeer) GetPeerCredentials() (*schema.PeerCredentialsResponse, error) {
	privKey, privateKey, err := wireguard.GeneratePrivateKey()
	if err != nil {
		w.logger.Error("failed to generate private key", zap.Error(err))
		return nil, err
	}

	publicKey, err := wireguard.GeneratePublicKey(privKey)
	if err != nil {
		w.logger.Error("failed to generate public key from private key", zap.Error(err))
		return nil, err
	}

	return &schema.PeerCredentialsResponse{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

func (w *WgPeer) GetNewPeerAllowedAddress(interfaceId uint) (*schema.NewPeerAllowedAddressResponse, error) {
	var iface model.Interface
	if err := w.db.Preload("IPPool").First(&iface, "id = ?", interfaceId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.logger.Error("interface not found in database", zap.Uint("interfaceId", interfaceId))
			return nil, fmt.Errorf("interface %d not found", interfaceId)
		}
		w.logger.Error("failed to query interface from database", zap.Error(err))
		return nil, err
	}
	bound, err := w.ForServer(iface.ServerID)
	if err != nil {
		return nil, err
	}
	w = bound

	if iface.IPPool == nil {
		w.logger.Warn("no IP pool associated with interface", zap.Uint("interfaceId", interfaceId))
		return &schema.NewPeerAllowedAddressResponse{AllowedAddress: ""}, nil
	}

	startIP := net.ParseIP(strings.TrimSpace(strings.Split(iface.IPPool.StartIP, "/")[0])).To4()
	endIP := net.ParseIP(strings.TrimSpace(strings.Split(iface.IPPool.EndIP, "/")[0])).To4()

	if startIP == nil || endIP == nil {
		w.logger.Error("invalid IP pool format",
			zap.String("start_ip", iface.IPPool.StartIP),
			zap.String("end_ip", iface.IPPool.EndIP))
		return nil, fmt.Errorf("invalid IP pool format")
	}

	var peers []model.Peer
	if err := w.db.Where("server_id = ?", iface.ServerID).Find(&peers, "interface = ?", iface.Name).Error; err != nil {
		w.logger.Error("failed to query peers from database", zap.Error(err))
		return nil, fmt.Errorf("failed to find peers: %w", err)
	}

	usedIPs := make([]net.IP, 0, len(peers))
	for _, peer := range peers {
		ipStr := strings.TrimSpace(strings.Split(peer.AllowedAddress, "/")[0])
		currentIP := net.ParseIP(ipStr).To4()
		if currentIP == nil {
			w.logger.Warn("skipping peer with invalid allowed_address",
				zap.String("allowed_address", peer.AllowedAddress))
			continue
		}
		usedIPs = append(usedIPs, currentIP)
	}
	usedIPs = append(usedIPs, w.interfaceReservedIPs(iface.Name)...)

	nextIP, err := utils.NextAvailableIPv4(startIP, endIP, usedIPs)
	if err != nil {
		w.logger.Error("IP pool exhausted or next IP out of range",
			zap.String("start_ip", startIP.String()),
			zap.String("end_ip", endIP.String()),
			zap.Error(err))
		return nil, fmt.Errorf("IP pool exhausted for interface %d", interfaceId)
	}

	return &schema.NewPeerAllowedAddressResponse{
		AllowedAddress: fmt.Sprintf("%s/32", nextIP.String()),
	}, nil
}

func (w *WgPeer) interfaceReservedIPs(ifaceName string) []net.IP {
	ifaceName = strings.TrimSpace(ifaceName)
	if w.mikrotikAdaptor == nil || ifaceName == "" {
		return nil
	}
	addrs, err := w.mikrotikAdaptor.FetchIPv4Addresses(context.Background())
	if err != nil {
		w.logger.Warn("failed to read interface addresses; the wireguard address may be offered to a peer",
			zap.String("interface", ifaceName),
			zap.Error(err))
		return nil
	}
	var reserved []net.IP
	for _, addr := range addrs {
		if !strings.EqualFold(addr.Interface, ifaceName) && !strings.EqualFold(addr.ActualInterface, ifaceName) {
			continue
		}
		reserved = append(reserved, utils.ReservedIPv4s(addr.Address)...)
	}
	return reserved
}

func (w *WgPeer) GetPeerShareStatus(id uint) (*schema.PeerShareStatusResponse, error) {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.logger.Error("peer not found in database", zap.Uint("id", id))
			return nil, gorm.ErrRecordNotFound
		}
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return nil, err
	}

	if !peer.IsShared {
		return &schema.PeerShareStatusResponse{
			IsShared:   false,
			UUID:       nil,
			ExpireTime: nil,
		}, nil
	}

	return &schema.PeerShareStatusResponse{
		IsShared:   peer.IsShared,
		UUID:       &peer.UUID,
		ExpireTime: peer.ShareExpireTime,
	}, nil
}

func (w *WgPeer) TogglePeerShareStatus(id uint) error {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return fmt.Errorf("peer not found: %w", err)
	}

	isShared := !peer.IsShared

	if err := w.db.Model(&peer).Update("is_shared", isShared).Error; err != nil {
		w.logger.Error("failed to update peer share status in database", zap.Error(err))
		return fmt.Errorf("failed to update peer share status: %w", err)
	}

	return nil
}

func (w *WgPeer) UpdatePeerShareExpireTime(id uint, expireTime *string) error {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return fmt.Errorf("peer not found: %w", err)
	}

	if !peer.IsShared {
		w.logger.Error("peer is not shared, cannot set expire time", zap.Uint("id", id))
		return fmt.Errorf("peer is not shared, cannot set expire time")
	}

	if err := w.db.Model(&peer).Update("share_expire_time", expireTime).Error; err != nil {
		w.logger.Error("failed to update peer share expire time in database", zap.Error(err))
		return fmt.Errorf("failed to update peer share expire time: %w", err)
	}

	return nil
}

func (w *WgPeer) GetPeerDetails(uuid string) (*schema.PeerDetailsResponse, error) {
	var peer model.Peer
	if err := w.db.First(&peer, "uuid = ?", uuid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.logger.Error("peer not found in database", zap.String("uuid", uuid))
			return nil, err
		}
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return nil, err
	}

	isSharable := utils.IsPeerSharable(peer.IsShared, peer.ShareExpireTime)
	if !isSharable {
		return nil, common.ErrPeerNotShared
	}

	totalUsage := peer.DownloadUsage + peer.UploadUsage

	var usagePercent, trafficLimit *string
	if peer.TrafficLimit != nil {
		trafficLimit = utils.Ptr(utils.BytesToGB(*peer.TrafficLimit))
		percent := float64(totalUsage) / float64(*peer.TrafficLimit) * 100
		usagePercent = utils.Ptr(fmt.Sprintf("%.1f", percent))
	}

	isOnline := w.PeerOnline(peer)

	return &schema.PeerDetailsResponse{
		Name:          peer.Name,
		UUID:          peer.UUID,
		Disabled:      peer.Disabled,
		TrafficLimit:  trafficLimit,
		ExpireTime:    peer.ExpireTime,
		DownloadUsage: utils.BytesToGB(peer.DownloadUsage),
		UploadUsage:   utils.BytesToGB(peer.UploadUsage),
		TotalUsage:    utils.BytesToGB(totalUsage),
		UsagePercent:  usagePercent,
		IsOnline:      isOnline,
	}, nil
}

func (w *WgPeer) GetPeerTelegramStatus(uuid string) (*schema.PeerTelegramStatusResponse, error) {
	var peer model.Peer
	if err := w.db.First(&peer, "uuid = ?", uuid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return nil, err
	}

	if !utils.IsPeerSharable(peer.IsShared, peer.ShareExpireTime) {
		return nil, common.ErrPeerNotShared
	}

	status := schema.PeerTelegramStatusResponse{}
	if w.shopLink != nil && w.shopLink.Enabled() {
		links, err := w.shopLink.Links(context.Background(), []string{peer.UUID})
		if err != nil {
			w.logger.Error("failed to lookup telegram chat links", zap.Error(err))
			return nil, err
		}
		link, ok := links[peer.UUID]
		if !ok || !link.Linked {
			return &status, nil
		}
		status.Linked = true
		status.NotifyEnabled = link.NotifyEnabled
		status.Username = link.Username
		return &status, nil
	}

	var chats []model.TelegramChat
	if err := w.db.Where("peer_uuid = ?", peer.UUID).Find(&chats).Error; err != nil {
		w.logger.Error("failed to lookup telegram chat links", zap.Error(err))
		return nil, err
	}
	if len(chats) == 0 {
		return &status, nil
	}

	chat := chats[0]
	for _, item := range chats {
		if item.Username != nil && *item.Username != "" {
			chat = item
			break
		}
	}

	status.Linked = true
	status.NotifyEnabled = chat.NotifyEnabled
	if chat.Username != nil && *chat.Username != "" {
		username := "@" + *chat.Username
		status.Username = &username
	}
	return &status, nil
}

func (w *WgPeer) GetPeers() (*[]schema.PeerResponse, error) {
	snapshots, err := w.peerSnapshots()
	if err != nil {
		return nil, err
	}
	peers := make([]schema.PeerResponse, 0, len(snapshots))
	for _, snapshot := range snapshots {
		peer := w.transformPeerToResponse(snapshot.peer)
		peer.IsOnline = snapshot.online
		peers = append(peers, peer)
	}
	w.attachTelegramLinkStatus(peers)
	return &peers, nil
}

func (w *WgPeer) CreatePeer(req *schema.CreatePeerRequest) (*schema.PeerResponse, error) {
	iface, err := w.getInterface(req.InterfaceId)
	if err != nil {
		return nil, err
	}
	bound, err := w.ForServer(iface.ServerID)
	if err != nil {
		return nil, err
	}
	w = bound

	if err := w.ensureAllowedAddressIsUnique(req.AllowedAddress, 0); err != nil {
		return nil, err
	}

	mtPeer, err := w.createMikrotikPeer(req, iface.Name)
	if err != nil {
		return nil, err
	}

	var (
		schedulerId *string
		queueId     *string
		dbPeer      model.Peer
		dbCreated   bool
	)

	cleanup := true
	defer func() {
		if !cleanup {
			return
		}
		w.rollbackPartialPeerCreate(mtPeer, schedulerId, queueId, dbPeer, dbCreated)
	}()

	schedulerId, err = w.scheduler.createScheduler(mtPeer.ID, mtPeer.Name, req.ExpireTime)
	if err != nil {
		return nil, err
	}

	queueId, err = w.queue.createQueue(mtPeer.Name, mtPeer.AllowedAddress, req.DownloadBandwidth, req.UploadBandwidth)
	if err != nil {
		return nil, err
	}

	dbPeer, err = w.buildAndStoreDbPeer(req, iface, mtPeer, schedulerId, queueId)
	if err != nil {
		return nil, err
	}
	dbCreated = true

	if err := w.generatePeerAssets(req.PrivateKey, dbPeer, iface.PublicKey); err != nil {
		return nil, err
	}

	if err := w.syncExpiredPeer(&dbPeer); err != nil {
		return nil, err
	}

	cleanup = false
	resp := w.peerResponse(dbPeer)
	return &resp, nil
}

func (w *WgPeer) UpdatePeer(id uint, req *schema.UpdatePeerRequest) (*schema.PeerResponse, error) {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to get peer from database", zap.Error(err))
		return nil, err
	}
	if peer.ResellerID != nil && !req.BillingAuthorized {
		return nil, ErrBilledOperationRequired
	}
	if peer.PurchaseID != nil {
		return nil, ErrPendingPeerOperation
	}
	bound, err := w.ForServer(peer.ServerID)
	if err != nil {
		return nil, err
	}
	w = bound

	manual := peer.ManualDisabled
	if !peer.RestrictionsInitialized {
		manual = peer.Disabled
	}
	if req.Disabled != nil && *req.Disabled != peer.Disabled {
		manual = *req.Disabled
	}
	if peer.ResellerID != nil && !req.BillingAuthorized {
		if req.TrafficLimit == nil || req.ExpireTime == nil {
			return nil, ErrInvalidBillingQuantity
		}
		quota, err := ParseBillingGiB(*req.TrafficLimit)
		if err != nil || quota <= 0 {
			return nil, ErrInvalidBillingQuantity
		}
		if peer.TrafficLimit == nil || quota > *peer.TrafficLimit || (peer.ExpireTime != nil && *req.ExpireTime > *peer.ExpireTime) {
			return nil, ErrBilledOperationRequired
		}
	}
	proposed := peer
	if req.ExpireTime != nil {
		proposed.ExpireTime = req.ExpireTime
	}
	if req.TrafficLimit != nil {
		quota, err := ParseBillingGiB(*req.TrafficLimit)
		if err == nil && quota > 0 {
			proposed.TrafficLimit = &quota
		}
	}
	blocked, err := peerOtherRestriction(w.db, proposed, time.Now())
	if err != nil {
		return nil, err
	}
	effective := manual || blocked
	copied := *req
	copied.Disabled = &effective
	req = &copied

	if err := w.ensureAllowedAddressIsUnique(req.AllowedAddress, peer.ID); err != nil {
		return nil, err
	}

	if err := w.updateMikrotikPeer(peer.PeerID, req); err != nil {
		return nil, err
	}

	schedulerID, err := w.handleScheduler(&peer, req)
	if err != nil {
		return nil, err
	}

	queueID, err := w.handleQueue(&peer, req)
	if err != nil {
		return nil, err
	}

	updateData := w.preparePeerUpdate(&peer, req, schedulerID, queueID)
	updateData["manual_disabled"] = manual
	updateData["restrictions_initialized"] = true
	updateData["restriction_revision"] = gorm.Expr("restriction_revision + 1")
	updateData["sync_next_attempt_at"] = 0
	if err := w.db.Model(&peer).Updates(updateData).Error; err != nil {
		return nil, err
	}

	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		return nil, err
	}

	if err := w.syncExpiredPeer(&peer); err != nil {
		return nil, err
	}

	iface, err := w.getInterfaceByName(peer.ServerID, peer.Interface)
	if err != nil {
		return nil, err
	}
	if err := w.generatePeerAssets(peer.PrivateKey, peer, iface.PublicKey); err != nil {
		return nil, err
	}
	transformed := w.peerResponse(peer)
	return &transformed, nil
}

func (w *WgPeer) DeletePeer(id uint) error {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return fmt.Errorf("peer not found: %w", err)
	}

	return w.deletePeerRecord(&peer)
}

func (w *WgPeer) DeletePeersByInterfaceName(interfaceName string) error {
	if w.serverID == 0 {
		return ErrServerSelectionRequired
	}
	var peers []model.Peer
	if err := w.db.Where("server_id = ? AND interface = ?", w.serverID, interfaceName).Find(&peers).Error; err != nil {
		w.logger.Error("failed to fetch peers for interface", zap.String("interface", interfaceName), zap.Error(err))
		return err
	}

	for _, peer := range peers {
		if peer.ResellerID != nil || peer.PurchaseID != nil {
			return ErrPendingPeerOperation
		}
	}
	var firstErr error
	for i := range peers {
		if err := w.deletePeerRecord(&peers[i]); err != nil {
			w.logger.Error("failed to delete peer during interface cascade",
				zap.String("peerID", peers[i].PeerID),
				zap.String("name", peers[i].Name),
				zap.Error(err),
			)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (w *WgPeer) rollbackPartialPeerCreate(mtPeer *mikrotik.WireGuardPeer, schedulerId, queueId *string, dbPeer model.Peer, dbCreated bool) {
	if queueId != nil {
		if err := w.queue.deleteQueue(queueId); err != nil {
			w.logger.Warn("failed to rollback queue after create failure", zap.Error(err))
		}
	}
	if schedulerId != nil {
		if err := w.scheduler.deleteScheduler(schedulerId); err != nil {
			w.logger.Warn("failed to rollback scheduler after create failure", zap.Error(err))
		}
	}
	if mtPeer != nil && mtPeer.ID != "" {
		if err := w.mikrotikAdaptor.DeleteWgPeer(context.Background(), mtPeer.ID); err != nil {
			w.logger.Warn("failed to rollback mikrotik peer after create failure", zap.Error(err))
		}
	}
	if !dbCreated || dbPeer.ID == 0 {
		return
	}

	_ = w.qrCodeGenerator.RemovePeerQRCodeByUUID(dbPeer.UUID)
	_ = w.configGenerator.RemovePeerConfigByUUID(dbPeer.UUID)
	_ = w.deletePeerTelegramChats(dbPeer.UUID)
	_ = w.deletePeerSessions(w.db, dbPeer.ID)
	if err := w.db.Unscoped().Delete(&dbPeer).Error; err != nil {
		w.logger.Warn("failed to rollback peer db row after create failure", zap.Error(err))
	}
}

func (w *WgPeer) deletePeerRecord(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is required")
	}
	if peer.ResellerID != nil && peer.PurchaseID == nil {
		return ErrBilledOperationRequired
	}
	if peer.PurchaseID != nil {
		return ErrPendingPeerOperation
	}
	bound, err := w.ForServer(peer.ServerID)
	if err != nil {
		return err
	}
	w = bound
	if peer == nil {
		return errors.New("peer is required")
	}

	if err := w.scheduler.deleteScheduler(peer.SchedulerID); err != nil {
		return fmt.Errorf("failed to delete scheduler: %w", err)
	}

	if err := w.queue.deleteQueue(peer.QueueID); err != nil {
		return fmt.Errorf("failed to delete simple queue: %w", err)
	}

	if err := w.mikrotikAdaptor.DeleteWgPeer(context.Background(), peer.PeerID); err != nil {
		w.logger.Error("failed to delete wireguard peer from Mikrotik", zap.Error(err))
		return fmt.Errorf("failed to delete wireguard peer: %w", err)
	}

	if err := w.qrCodeGenerator.RemovePeerQRCodeByUUID(peer.UUID); err != nil {
		w.logger.Error("failed to remove QR Code file", zap.Error(err))
		return err
	}

	if err := w.configGenerator.RemovePeerConfigByUUID(peer.UUID); err != nil {
		w.logger.Error("failed to remove peer config", zap.Error(err))
		return err
	}

	if err := w.deletePeerTelegramChats(peer.UUID); err != nil {
		w.logger.Error("failed to delete peer telegram chats", zap.Error(err))
		return fmt.Errorf("failed to delete peer telegram chats: %w", err)
	}

	if err := w.deletePeerSessions(w.db, peer.ID); err != nil {
		w.logger.Error("failed to delete peer sessions", zap.Error(err))
		return fmt.Errorf("failed to delete peer sessions: %w", err)
	}

	if err := w.db.Unscoped().Delete(peer).Error; err != nil {
		w.logger.Error("failed to delete peer from database", zap.Error(err))
		return fmt.Errorf("failed to delete peer from database: %w", err)
	}

	return nil
}

func (w *WgPeer) deletePeerTelegramChats(peerUUID string) error {
	if strings.TrimSpace(peerUUID) == "" {
		return nil
	}
	if w.shopLink != nil && w.shopLink.Enabled() {
		if err := w.shopLink.Unlink(context.Background(), peerUUID); err != nil {
			return err
		}
	}
	return w.db.Unscoped().Where("peer_uuid = ?", peerUUID).Delete(&model.TelegramChat{}).Error
}

type peerSnapshot struct {
	peer      model.Peer
	online    bool
	handshake time.Duration
}

func (w *WgPeer) peerSnapshots() ([]peerSnapshot, error) {
	var peers []model.Peer
	if err := w.peerQuery().Order("id ASC").Find(&peers).Error; err != nil {
		return nil, err
	}
	live := make(map[uint]map[string]mikrotik.WireGuardPeer)
	for _, peer := range peers {
		if _, seen := live[peer.ServerID]; seen {
			continue
		}
		live[peer.ServerID] = map[string]mikrotik.WireGuardPeer{}
		adaptor, err := w.mikrotikAdaptor.ForServer(peer.ServerID)
		if err != nil {
			continue
		}
		observations, err := adaptor.FetchWgPeers(context.Background())
		if err != nil {
			w.logger.Warn("router unavailable; showing stored peers", zap.Uint("server_id", peer.ServerID), zap.Error(err))
			continue
		}
		for _, observation := range observations {
			live[peer.ServerID][observation.ID] = observation
		}
	}
	snapshots := make([]peerSnapshot, 0, len(peers))
	for _, peer := range peers {
		snapshot := peerSnapshot{peer: peer}
		if observed, ok := live[peer.ServerID][peer.PeerID]; ok {
			duration, online, err := w.handshakeData(&observed)
			if err == nil {
				snapshot.online = online
				snapshot.handshake = duration
			}
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

func (w *WgPeer) GetPeersData() (*schema.PeerStatsResponse, error) {
	snapshots, err := w.peerSnapshots()
	if err != nil {
		return nil, err
	}
	var online []peerSnapshot
	disabled := 0
	for _, snapshot := range snapshots {
		if snapshot.peer.Disabled {
			disabled++
			continue
		}
		if snapshot.online {
			online = append(online, snapshot)
		}
	}
	sort.Slice(online, func(i, j int) bool { return online[i].handshake < online[j].handshake })
	recent := make([]schema.RecentOnlinePeers, 0, min(5, len(online)))
	for _, snapshot := range online[:min(5, len(online))] {
		recent = append(recent, schema.RecentOnlinePeers{Name: snapshot.peer.Name, LastSeen: utils.FormatDuration(snapshot.handshake)})
	}
	return &schema.PeerStatsResponse{RecentOnlinePeers: &recent, TotalPeers: len(snapshots), OnlinePeers: len(online), DisabledPeers: disabled, OfflinePeers: len(snapshots) - len(online) - disabled}, nil
}

func (w *WgPeer) getInterface(id uint) (model.Interface, error) {
	var iface model.Interface
	if err := w.db.First(&iface, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.logger.Error("interface not found", zap.Uint("interfaceId", id))
			return iface, fmt.Errorf("interface %d not found", id)
		}
		w.logger.Error("db error while fetching interface", zap.Error(err))
		return iface, err
	}
	return iface, nil
}

func (w *WgPeer) ensureAllowedAddressIsUnique(address string, excludePeerID uint) error {
	var existing model.Peer
	query := w.db.Where("server_id = ? AND allowed_address = ?", w.serverID, address)
	if excludePeerID > 0 {
		query = query.Where("id <> ?", excludePeerID)
	}

	if err := query.First(&existing).Error; err == nil {
		return fmt.Errorf("allowed address %s is already in use by peer %s", address, existing.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		w.logger.Error("allowed address lookup failed", zap.Error(err))
		return err
	}
	return nil
}

func (w *WgPeer) createMikrotikPeer(req *schema.CreatePeerRequest, ifaceName string) (*mikrotik.WireGuardPeer, error) {
	peer := &mikrotik.WireGuardPeer{
		Comment:        req.Comment,
		Name:           req.Name,
		AllowedAddress: req.AllowedAddress,
		Interface:      ifaceName,
		PresharedKey:   req.PresharedKey,
		PrivateKey:     &req.PrivateKey,
		PublicKey:      req.PublicKey,
	}
	return w.mikrotikAdaptor.CreateWgPeer(context.Background(), *peer)
}

func (w *WgPeer) buildAndStoreDbPeer(req *schema.CreatePeerRequest, iface model.Interface, mtPeer *mikrotik.WireGuardPeer, schedulerId, queueId *string) (model.Peer, error) {
	keepalive := common.DefaultKeepalive
	if req.PersistentKeepAlive != nil {
		parsed, err := timehelper.ParseTime(*req.PersistentKeepAlive)
		if err != nil {
			w.logger.Error("invalid keepalive", zap.Error(err))
			return model.Peer{}, err
		}
		keepalive = strconv.Itoa(parsed)
	}

	disabled, err := strconv.ParseBool(mtPeer.Disabled)
	if err != nil {
		w.logger.Error("invalid mikrotik disabled field", zap.Error(err))
		return model.Peer{}, err
	}

	var trafficLimit *int64
	if req.TrafficLimit != nil {
		trafficBytes := utils.GBToBytes(utils.DerefString(req.TrafficLimit))
		trafficLimit = &trafficBytes
	}

	dbPeer := model.Peer{
		RestrictionsInitialized: true,
		ManualDisabled:          disabled,
		ServerID:                iface.ServerID,
		UUID:                    uuid.New().String(),
		PeerID:                  mtPeer.ID,
		Disabled:                disabled,
		Comment:                 mtPeer.Comment,
		Name:                    mtPeer.Name,
		PrivateKey:              *mtPeer.PrivateKey,
		PublicKey:               mtPeer.PublicKey,
		Interface:               mtPeer.Interface,
		AllowedAddress:          mtPeer.AllowedAddress,
		Endpoint:                req.Endpoint,
		EndpointPort:            iface.ListenPort,
		PersistentKeepalive:     keepalive,
		SchedulerID:             schedulerId,
		QueueID:                 queueId,
		ExpireTime:              req.ExpireTime,
		TrafficLimit:            trafficLimit,
		DownloadBandwidth:       req.DownloadBandwidth,
		UploadBandwidth:         req.UploadBandwidth,
	}

	if err := w.db.Create(&dbPeer).Error; err != nil {
		w.logger.Error("failed to persist peer", zap.Error(err))
		return model.Peer{}, err
	}

	return dbPeer, nil
}

func (w *WgPeer) generatePeerAssets(privateKey string, peer model.Peer, ifacePubKey string) error {
	peerConfig := fmt.Sprintf(wireguard.Template, privateKey, peer.AllowedAddress, common.DefaultDns, ifacePubKey, peer.Endpoint, peer.EndpointPort, common.AllowedIpsIncludeLocal, peer.PersistentKeepalive)

	if err := w.configGenerator.BuildPeerConfig(peerConfig, peer.UUID); err != nil {
		return err
	}
	return w.qrCodeGenerator.BuildPeerQRCode(peerConfig, peer.UUID)
}

func (w *WgPeer) updateMikrotikPeer(peerID string, req *schema.UpdatePeerRequest) error {
	wgPeer := mikrotik.WireGuardPeer{}

	if req.Disabled != nil {
		disabledStr := strconv.FormatBool(*req.Disabled)
		wgPeer.Disabled = disabledStr
	}
	if req.Comment != nil {
		wgPeer.Comment = req.Comment
	}

	wgPeer.Name = req.Name
	wgPeer.AllowedAddress = req.AllowedAddress

	if req.PersistentKeepAlive != nil {
		wgPeer.PersistentKeepAlive = req.PersistentKeepAlive
	}
	if req.PresharedKey != nil {
		wgPeer.PresharedKey = req.PresharedKey
	}

	_, err := w.mikrotikAdaptor.UpdateWgPeer(context.Background(), peerID, wgPeer)
	if err != nil {
		w.logger.Error("failed to update wireguard peer in Mikrotik", zap.Error(err))
	}

	return err
}

func (w *WgPeer) handleScheduler(peer *model.Peer, req *schema.UpdatePeerRequest) (*string, error) {
	if req.ExpireTime == nil && peer.SchedulerID != nil {
		err := w.scheduler.deleteScheduler(peer.SchedulerID)
		if err != nil {
			w.logger.Error("failed to delete scheduler for wireguard peer", zap.Error(err))
			return peer.SchedulerID, err
		}
		return nil, nil
	}

	if req.ExpireTime != nil && peer.SchedulerID == nil {
		return w.scheduler.createScheduler(peer.PeerID, peer.Name, req.ExpireTime)
	}

	if req.ExpireTime != nil && peer.SchedulerID != nil {
		err := w.scheduler.updateScheduler(peer.SchedulerID, peer.PeerID, req.ExpireTime)
		if err != nil {
			w.logger.Error("failed to update scheduler for wireguard peer", zap.Error(err))
			return peer.SchedulerID, err
		}
	}

	return peer.SchedulerID, nil
}

func (w *WgPeer) handleQueue(peer *model.Peer, req *schema.UpdatePeerRequest) (*string, error) {
	download := req.DownloadBandwidth
	upload := req.UploadBandwidth
	queueID := peer.QueueID

	if download == nil && upload == nil {
		if queueID != nil {
			err := w.queue.deleteQueue(queueID)
			if err != nil {
				w.logger.Error("failed to delete queue for wireguard peer", zap.Error(err))
				return queueID, err
			}
		}
		return nil, nil
	}

	if queueID == nil {
		newQueueID, err := w.queue.createQueue(peer.Name, peer.AllowedAddress, download, upload)
		if err != nil {
			w.logger.Error("failed to create queue for wireguard peer", zap.Error(err))
			return nil, err
		}
		return newQueueID, nil
	}

	if !w.bandwidthsEqual(peer.DownloadBandwidth, download) || !w.bandwidthsEqual(peer.UploadBandwidth, upload) {
		err := w.queue.updateQueue(queueID, download, upload)
		if err != nil {
			w.logger.Error("failed to update queue for wireguard peer", zap.Error(err))
			return queueID, err
		}
	}

	return queueID, nil
}

func (w *WgPeer) preparePeerUpdate(peer *model.Peer, req *schema.UpdatePeerRequest, schedulerID, queueID *string) map[string]interface{} {
	updateData := map[string]interface{}{}

	trafficBytes := utils.GBToBytes(utils.DerefString(req.TrafficLimit))
	var trafficLimit *int64
	if trafficBytes > 0 {
		trafficLimit = &trafficBytes
		updateData["traffic_limit"] = trafficBytes
	} else {
		updateData["traffic_limit"] = nil
	}

	if !int64PtrEqual(peer.TrafficLimit, trafficLimit) {
		updateData["traffic_first_notify"] = false
		updateData["traffic_second_notify"] = false
		updateData["traffic_third_notify"] = false
	}

	if !w.bandwidthsEqual(peer.ExpireTime, req.ExpireTime) {
		updateData["expire_first_notify"] = false
		updateData["expire_second_notify"] = false
		updateData["expire_third_notify"] = false
	}

	updateData["disabled"] = req.Disabled
	updateData["comment"] = req.Comment
	updateData["name"] = req.Name
	updateData["allowed_address"] = req.AllowedAddress
	updateData["persistent_keepalive"] = req.PersistentKeepAlive
	updateData["expire_time"] = req.ExpireTime
	updateData["download_bandwidth"] = req.DownloadBandwidth
	updateData["upload_bandwidth"] = req.UploadBandwidth
	updateData["scheduler_id"] = schedulerID
	updateData["queue_id"] = queueID

	return updateData
}

func (w *WgPeer) transformPeerToResponse(peer model.Peer) schema.PeerResponse {
	statuses := w.transformPeerStatus(peer)

	var trafficLimit *string

	if peer.TrafficLimit == nil {
		trafficLimit = nil
	} else {
		trafficLimit = utils.Ptr(utils.BytesToGB(*peer.TrafficLimit))
	}

	reasons := []string{}
	if peer.ManualDisabled {
		reasons = append(reasons, "manual")
	}
	if utils.IsPeerExpired(peer.ExpireTime, time.Now()) {
		reasons = append(reasons, "expired")
	}
	if peer.TrafficLimit != nil && peer.DownloadUsage+peer.UploadUsage >= *peer.TrafficLimit {
		reasons = append(reasons, "quota")
	}
	if peer.BillingBlocked {
		reasons = append(reasons, "billing")
	}
	if peer.AdminBlocked {
		reasons = append(reasons, "administrative")
	}
	if peer.ProvisioningBlocked {
		reasons = append(reasons, "provisioning")
	}
	syncStatus := "synced"
	if peer.SyncedDisabled == nil || peer.AppliedRevision != peer.RestrictionRevision {
		syncStatus = "pending"
	}
	if peer.SyncError != "" {
		syncStatus = "error"
	}
	return schema.PeerResponse{
		TrafficLimitBytes: int64Text(peer.TrafficLimit), TotalUsageBytes: strconv.FormatInt(peer.DownloadUsage+peer.UploadUsage, 10), ConfiguredDurationDays: strconv.FormatInt(peer.ConfiguredDurationDays, 10), PurchasePending: peer.PurchaseID != nil,
		ManualDisabled: peer.ManualDisabled, RestrictionReasons: reasons, SyncStatus: syncStatus,
		Id:                peer.ID,
		ServerID:          peer.ServerID,
		ResellerID:        peer.ResellerID,
		UUID:              peer.UUID,
		Disabled:          peer.Disabled,
		Comment:           peer.Comment,
		TelegramUsername:  nil,
		TelegramLinked:    false,
		Name:              peer.Name,
		Interface:         peer.Interface,
		AllowedAddress:    peer.AllowedAddress,
		TrafficLimit:      trafficLimit,
		ExpireTime:        peer.ExpireTime,
		CreatedAt:         utils.FormatUnixTime(int64(peer.CreatedAt)),
		DownloadBandwidth: peer.DownloadBandwidth,
		UploadBandwidth:   peer.UploadBandwidth,
		TotalUsage:        utils.BytesToGB(peer.DownloadUsage + peer.UploadUsage),
		Status:            statuses,
		IsShared:          peer.IsShared,
	}
}

func (w *WgPeer) peerResponse(peer model.Peer) schema.PeerResponse {
	peers := []schema.PeerResponse{w.transformPeerToResponse(peer)}
	w.attachTelegramLinkStatus(peers)
	return peers[0]
}

func (w *WgPeer) attachTelegramLinkStatus(peers []schema.PeerResponse) {
	uuids := make([]string, 0, len(peers))
	for _, peer := range peers {
		if peer.UUID != "" {
			uuids = append(uuids, peer.UUID)
		}
	}
	if len(uuids) == 0 {
		return
	}

	if w.shopLink != nil && w.shopLink.Enabled() {
		links, err := w.shopLink.Links(context.Background(), uuids)
		if err != nil {
			w.logger.Error("failed to lookup telegram chat links", zap.Error(err))
			return
		}
		for i := range peers {
			link, ok := links[peers[i].UUID]
			if !ok || !link.Linked {
				continue
			}
			peers[i].TelegramLinked = true
			peers[i].TelegramUsername = link.Username
		}
		return
	}

	var chats []model.TelegramChat
	if err := w.db.Where("peer_uuid IN ?", uuids).Find(&chats).Error; err != nil {
		w.logger.Error("failed to lookup telegram chat links", zap.Error(err))
		return
	}

	linked := make(map[string]model.TelegramChat, len(chats))
	for _, chat := range chats {
		existing, ok := linked[chat.PeerUUID]
		if !ok || (existing.Username == nil && chat.Username != nil) {
			linked[chat.PeerUUID] = chat
		}
	}

	for i := range peers {
		chat, ok := linked[peers[i].UUID]
		if !ok {
			continue
		}
		peers[i].TelegramLinked = true
		if chat.Username != nil && *chat.Username != "" {
			username := "@" + *chat.Username
			peers[i].TelegramUsername = &username
		}
	}
}

func (w *WgPeer) transformPeerStatus(peer model.Peer) []schema.PeerStatus {
	var peerStatus []schema.PeerStatus

	if peer.Disabled {
		peerStatus = append(peerStatus, schema.InactivePeer)
	} else {
		peerStatus = append(peerStatus, schema.ActivePeer)
	}

	if utils.IsPeerExpired(peer.ExpireTime, time.Now()) {
		peerStatus = append(peerStatus, schema.ExpiredPeer)
	}

	if peer.TrafficLimit != nil {
		totalUsedTraffic := peer.DownloadUsage + peer.UploadUsage
		if totalUsedTraffic >= *peer.TrafficLimit {
			peerStatus = append(peerStatus, schema.SuspendedPeer)
		}
	}

	return peerStatus
}

func (w *WgPeer) syncExpiredPeer(peer *model.Peer) error {
	if peer.PurchaseID != nil || peer.Disabled || !utils.IsPeerExpired(peer.ExpireTime, time.Now()) {
		return nil
	}

	disabled := strconv.FormatBool(true)
	if _, err := w.mikrotikAdaptor.UpdateWgPeer(context.Background(), peer.PeerID, mikrotik.WireGuardPeer{
		Disabled: disabled,
	}); err != nil {
		w.logger.Error("failed to disable expired wireguard peer", zap.Error(err))
		return fmt.Errorf("failed to disable expired peer: %w", err)
	}

	if peer.QueueID != nil {
		if _, err := w.mikrotikAdaptor.UpdateSimpleQueue(context.Background(), *peer.QueueID, mikrotik.Queue{
			Disabled: disabled,
		}); err != nil {
			w.logger.Error("failed to disable queue for expired peer", zap.Error(err))
		}
	}

	if err := w.db.Model(peer).Update("disabled", true).Error; err != nil {
		w.logger.Error("failed to mark expired peer as disabled", zap.Error(err))
		return fmt.Errorf("failed to mark expired peer as disabled: %w", err)
	}

	peer.Disabled = true
	return nil
}

func (w *WgPeer) handshakeData(peer *mikrotik.WireGuardPeer) (duration time.Duration, isOnline bool, err error) {
	duration, isOnline, err = utils.HandshakeStatus(peer.Disabled, peer.LastHandshake, common.PeerOnlineHandshakeTimeout)
	if err != nil {
		w.logger.Error("failed to parse last handshake duration", zap.Error(err))
	}
	return
}

func (w *WgPeer) bandwidthsEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a != nil && b != nil {
		return *a == *b
	}
	return false
}

func int64PtrEqual(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func (w *WgPeer) getInterfaceByName(serverID uint, name string) (model.Interface, error) {
	var iface model.Interface
	err := w.db.Where("server_id = ? AND name = ?", serverID, name).First(&iface).Error
	return iface, err
}
