package service

import (
	"bytes"
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
	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/utils"
	"github.com/maahdima/mwp/api/utils/timehelper"
	"github.com/maahdima/mwp/api/utils/wireguard"
)

type WgPeer struct {
	db              *gorm.DB
	mikrotikAdaptor *mikrotik.Adaptor
	scheduler       *Scheduler
	queue           *Queue
	configGenerator *ConfigGenerator
	qrCodeGenerator *QRCodeGenerator
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
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return fmt.Errorf("peer not found: %w", err)
	}

	disabled := strconv.FormatBool(!peer.Disabled)

	wgPeer := mikrotik.WireGuardPeer{
		Disabled: disabled,
	}
	wgScheduler := mikrotik.Scheduler{
		Disabled: disabled,
	}
	wgQueue := mikrotik.Queue{
		Disabled: disabled,
	}

	if _, err := w.mikrotikAdaptor.UpdateWgPeer(context.Background(), peer.PeerID, wgPeer); err != nil {
		w.logger.Error("failed to update wireguard peer in Mikrotik", zap.Error(err))
		return fmt.Errorf("failed to update wireguard peer: %w", err)
	}

	if peer.SchedulerID != nil {
		if _, err := w.mikrotikAdaptor.UpdateScheduler(context.Background(), *peer.SchedulerID, wgScheduler); err != nil {
			w.logger.Error("failed to update scheduler for wireguard peer", zap.Error(err))
			return fmt.Errorf("failed to update scheduler: %w", err)
		}
	}
	if peer.QueueID != nil {
		if _, err := w.mikrotikAdaptor.UpdateSimpleQueue(context.Background(), *peer.QueueID, wgQueue); err != nil {
			w.logger.Error("failed to update queue for wireguard peer", zap.Error(err))
			return fmt.Errorf("failed to update queue: %w", err)
		}
	}

	if err := w.db.Model(&peer).Update("disabled", disabled).Error; err != nil {
		w.logger.Error("failed to update peer status in database", zap.Error(err))
		return fmt.Errorf("failed to update peer status in database: %w", err)
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

	if iface.IPPool == nil {
		w.logger.Warn("no IP pool associated with interface", zap.Uint("interfaceId", interfaceId))
		return &schema.NewPeerAllowedAddressResponse{
			AllowedAddress: "",
		}, nil
	}

	startIP := net.ParseIP(strings.TrimSuffix(iface.IPPool.StartIP, "/32")).To4()
	endIP := net.ParseIP(strings.TrimSuffix(iface.IPPool.EndIP, "/32")).To4()

	if startIP == nil || endIP == nil {
		w.logger.Error("invalid IP pool format",
			zap.String("start_ip", iface.IPPool.StartIP),
			zap.String("end_ip", iface.IPPool.EndIP))
		return nil, fmt.Errorf("invalid IP pool format")
	}

	var lastPeer model.Peer
	err := w.db.Order("allowed_address DESC").First(&lastPeer, "interface = ?", iface.Name).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			w.logger.Error("failed to find last peer", zap.Error(err))
			return nil, fmt.Errorf("failed to find last peer: %w", err)
		}

		w.logger.Info("no peers found for interface, assigning start IP", zap.Uint("interfaceId", interfaceId))
		return &schema.NewPeerAllowedAddressResponse{
			AllowedAddress: fmt.Sprintf("%s/32", startIP.String()),
		}, nil
	}

	lastIP, _, err := net.ParseCIDR(lastPeer.AllowedAddress)
	if err != nil {
		w.logger.Error("invalid CIDR format in allowed address", zap.String("allowed_address", lastPeer.AllowedAddress), zap.Error(err))
		return nil, fmt.Errorf("failed to parse allowed address: %w", err)
	}

	lastIP = lastIP.To4()
	if lastIP == nil {
		w.logger.Error("allowed address is not a valid IPv4", zap.String("allowed_address", lastPeer.AllowedAddress))
		return nil, fmt.Errorf("invalid IPv4 address: %s", lastPeer.AllowedAddress)
	}

	nextIP := make(net.IP, len(lastIP))
	copy(nextIP, lastIP)
	for i := len(nextIP) - 1; i >= 0; i-- {
		nextIP[i]++
		if nextIP[i] != 0 {
			break
		}
	}

	if bytes.Compare(nextIP, endIP) > 0 {
		w.logger.Error("IP pool exhausted or next IP out of range",
			zap.String("next_ip", nextIP.String()),
			zap.String("end_ip", endIP.String()))
		return nil, fmt.Errorf("IP pool exhausted for interface %d", interfaceId)
	}

	return &schema.NewPeerAllowedAddressResponse{
		AllowedAddress: fmt.Sprintf("%s/32", nextIP.String()),
	}, nil
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
			ShareLink:  nil,
			ExpireTime: nil,
		}, nil
	}

	appCfg := config.GetAppConfig()

	// TODO: https
	shareLink := fmt.Sprintf("http://%s:%s/share?shareId=%s", appCfg.Host, appCfg.Port, peer.UUID)

	return &schema.PeerShareStatusResponse{
		IsShared:   peer.IsShared,
		ShareLink:  &shareLink,
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

	return &schema.PeerDetailsResponse{
		Name:          peer.Name,
		TrafficLimit:  trafficLimit,
		ExpireTime:    peer.ExpireTime,
		DownloadUsage: utils.BytesToGB(peer.DownloadUsage),
		UploadUsage:   utils.BytesToGB(peer.UploadUsage),
		TotalUsage:    utils.BytesToGB(totalUsage),
		UsagePercent:  usagePercent,
	}, nil
}

func (w *WgPeer) GetPeers() (*[]schema.PeerResponse, error) {
	var dbPeers []model.Peer
	if err := w.db.Order("created_at ASC").Find(&dbPeers).Error; err != nil {
		w.logger.Error("failed to get peers from database", zap.Error(err))
		return nil, err
	}

	var wgPeers []schema.PeerResponse
	for _, dbPeer := range dbPeers {
		peer, err := w.mikrotikAdaptor.FetchWgPeer(context.Background(), dbPeer.PeerID)
		if err != nil {
			w.logger.Error("failed to fetch wireguard peer from Mikrotik", zap.String("peerID", dbPeer.PeerID), zap.Error(err))
			continue
		}

		wgPeer := w.transformPeerToResponse(dbPeer)

		var duration time.Duration
		if peer.LastHandshake != nil {
			duration, err = time.ParseDuration(*peer.LastHandshake)
			if err != nil {
				w.logger.Error("failed to parse last handshake duration", zap.Error(err))
				return nil, fmt.Errorf("failed to parse last handshake duration: %w", err)
			}

			if duration < 150*time.Second {
				wgPeer.IsOnline = true
			}
		}

		wgPeers = append(wgPeers, wgPeer)
	}

	return &wgPeers, nil
}

func (w *WgPeer) CreatePeer(req *schema.CreatePeerRequest) (*schema.PeerResponse, error) {
	iface, err := w.getInterface(req.InterfaceId)
	if err != nil {
		return nil, err
	}

	if err := w.ensureAllowedAddressIsUnique(req.AllowedAddress); err != nil {
		return nil, err
	}

	mtPeer, err := w.createMikrotikPeer(req, iface.Name)
	if err != nil {
		return nil, err
	}

	schedulerId, err := w.scheduler.createScheduler(mtPeer.ID, mtPeer.Name, req.ExpireTime)
	if err != nil {
		return nil, err
	}

	queueId, err := w.queue.createQueue(mtPeer.Name, mtPeer.AllowedAddress, req.DownloadBandwidth, req.UploadBandwidth)
	if err != nil {
		return nil, err
	}

	dbPeer, err := w.buildAndStoreDbPeer(req, iface, mtPeer, schedulerId, queueId)
	if err != nil {
		return nil, err
	}

	if err := w.generatePeerAssets(req.PrivateKey, dbPeer, iface.PublicKey); err != nil {
		return nil, err
	}

	resp := w.transformPeerToResponse(dbPeer)
	return &resp, nil
}

func (w *WgPeer) UpdatePeer(id uint, req *schema.UpdatePeerRequest) (*schema.PeerResponse, error) {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to get peer from database", zap.Error(err))
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

	updateData := w.preparePeerUpdate(req, schedulerID, queueID)
	if err := w.db.Model(&peer).Updates(updateData).Error; err != nil {
		return nil, err
	}

	transformed := w.transformPeerToResponse(peer)
	return &transformed, nil
}

func (w *WgPeer) DeletePeer(id uint) error {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return fmt.Errorf("peer not found: %w", err)
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

	err := w.qrCodeGenerator.RemovePeerQRCode(id)
	if err != nil {
		w.logger.Error("failed to remove QR Code file", zap.Error(err))
		return err
	}

	err = w.configGenerator.RemovePeerConfig(id)
	if err != nil {
		w.logger.Error("failed to remove peer config", zap.Error(err))
		return err
	}

	if err := w.db.Unscoped().Delete(&peer).Error; err != nil {
		w.logger.Error("failed to delete peer from database", zap.Error(err))
		return fmt.Errorf("failed to delete peer from database: %w", err)
	}

	return nil
}

func (w *WgPeer) GetPeersData() (*schema.PeerStatsResponse, error) {
	var dbPeers []model.Peer
	if err := w.db.Find(&dbPeers).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.logger.Warn("no peers found in database")
			return &schema.PeerStatsResponse{}, nil
		}
		w.logger.Error("failed to fetch peers from database", zap.Error(err))
		return nil, fmt.Errorf("fetch peers from database: %w", err)
	}

	type peerWithDuration struct {
		peer     *mikrotik.WireGuardPeer
		duration time.Duration
	}

	var (
		allOnlinePeers []peerWithDuration
		disabledPeers  []mikrotik.WireGuardPeer
	)

	for _, dbPeer := range dbPeers {
		peer, err := w.mikrotikAdaptor.FetchWgPeer(context.Background(), dbPeer.PeerID)
		if err != nil {
			w.logger.Error("failed to fetch peer from mikrotik", zap.String("peerID", dbPeer.PeerID), zap.Error(err))
			return nil, fmt.Errorf("fetch peer %s from mikrotik: %w", dbPeer.PeerID, err)
		}

		if peer.Disabled == "true" {
			disabledPeers = append(disabledPeers, *peer)
			continue
		}

		if peer.LastHandshake != nil {
			duration, err := time.ParseDuration(*peer.LastHandshake)
			if err != nil {
				w.logger.Error("invalid last handshake duration", zap.String("peerID", dbPeer.PeerID), zap.Error(err))
				continue
			}
			allOnlinePeers = append(allOnlinePeers, peerWithDuration{
				peer:     peer,
				duration: duration,
			})
		}
	}

	onlinePeersCount := len(allOnlinePeers)
	offlinePeersCount := len(dbPeers) - onlinePeersCount - len(disabledPeers)

	sort.Slice(allOnlinePeers, func(i, j int) bool {
		return allOnlinePeers[i].duration < allOnlinePeers[j].duration
	})

	var recentPeers []schema.RecentOnlinePeers
	for _, item := range allOnlinePeers[:min(5, len(allOnlinePeers))] {
		recentPeers = append(recentPeers, schema.RecentOnlinePeers{
			Name:     item.peer.Name,
			LastSeen: time.Unix(int64(item.duration.Seconds()), 0).UTC().Format("15:04:05"),
		})
	}

	return &schema.PeerStatsResponse{
		RecentOnlinePeers: &recentPeers,
		TotalPeers:        len(dbPeers),
		OnlinePeers:       onlinePeersCount,
		OfflinePeers:      offlinePeersCount,
		DisabledPeers:     len(disabledPeers),
	}, nil
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

func (w *WgPeer) ensureAllowedAddressIsUnique(address string) error {
	var existing model.Peer
	if err := w.db.Where("allowed_address = ?", address).First(&existing).Error; err == nil {
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
		bytes := utils.GBToBytes(utils.DerefString(req.TrafficLimit))
		trafficLimit = &bytes
	}

	dbPeer := model.Peer{
		UUID:                uuid.New().String(),
		PeerID:              mtPeer.ID,
		Disabled:            disabled,
		Comment:             mtPeer.Comment,
		Name:                mtPeer.Name,
		PrivateKey:          *mtPeer.PrivateKey,
		PublicKey:           mtPeer.PublicKey,
		Interface:           mtPeer.Interface,
		AllowedAddress:      mtPeer.AllowedAddress,
		Endpoint:            req.Endpoint,
		EndpointPort:        iface.ListenPort,
		PersistentKeepalive: keepalive,
		SchedulerID:         schedulerId,
		QueueID:             queueId,
		ExpireTime:          req.ExpireTime,
		TrafficLimit:        trafficLimit,
		DownloadBandwidth:   req.DownloadBandwidth,
		UploadBandwidth:     req.UploadBandwidth,
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

	if req.ExpireTime != nil {
		err := w.scheduler.updateScheduler(peer.SchedulerID, req.ExpireTime)
		if err != nil {
			w.logger.Error("failed to update scheduler for WireGuard peer", zap.Error(err))
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

	err := w.queue.updateQueue(queueID, download, upload)
	if err != nil {
		w.logger.Error("failed to update queue for wireguard peer", zap.Error(err))
		return queueID, err
	}

	return queueID, nil
}

func (w *WgPeer) preparePeerUpdate(req *schema.UpdatePeerRequest, schedulerID, queueID *string) map[string]interface{} {
	updateData := map[string]interface{}{}

	trafficLimit := utils.GBToBytes(utils.DerefString(req.TrafficLimit))
	if trafficLimit > 0 {
		updateData["traffic_limit"] = trafficLimit
	} else {
		updateData["traffic_limit"] = nil
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

	return schema.PeerResponse{
		Id:                peer.ID,
		UUID:              peer.UUID,
		Disabled:          peer.Disabled,
		Comment:           peer.Comment,
		Name:              peer.Name,
		Interface:         peer.Interface,
		AllowedAddress:    peer.AllowedAddress,
		TrafficLimit:      trafficLimit,
		ExpireTime:        peer.ExpireTime,
		DownloadBandwidth: peer.DownloadBandwidth,
		UploadBandwidth:   peer.UploadBandwidth,
		TotalUsage:        utils.BytesToGB(peer.DownloadUsage + peer.UploadUsage),
		Status:            statuses,
		IsShared:          peer.IsShared,
	}
}

func (w *WgPeer) transformPeerStatus(peer model.Peer) []schema.PeerStatus {
	var peerStatus []schema.PeerStatus

	if peer.Disabled {
		peerStatus = append(peerStatus, schema.InactivePeer)
	} else {
		peerStatus = append(peerStatus, schema.ActivePeer)
	}

	if peer.ExpireTime != nil {
		expireTime, err := time.Parse("2006-01-02", *peer.ExpireTime)
		if err == nil && time.Now().After(expireTime) {
			peerStatus = append(peerStatus, schema.ExpiredPeer)
		}
	}

	if peer.TrafficLimit != nil {
		totalUsedTraffic := peer.DownloadUsage + peer.UploadUsage
		if totalUsedTraffic > *peer.TrafficLimit {
			peerStatus = append(peerStatus, schema.SuspendedPeer)
		}
	}

	return peerStatus
}
