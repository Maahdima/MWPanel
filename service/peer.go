package service

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"mikrotik-wg-go/adaptor/mikrotik"
	"mikrotik-wg-go/api/schema"
	"mikrotik-wg-go/dataservice/model"
	"mikrotik-wg-go/utils/timehelper"
	"mikrotik-wg-go/utils/wireguard"
	"sort"
	"strconv"
	"time"
)

var (
	peerConfigsPath = "assets/config"
	peerQrCodesPath = "assets/qrcode"
)

var (
	schedulerComment   = "Expire WireGuard Peer: "
	schedulerName      = "Schedule: "
	schedulerStartTime = "12:00:00"
	schedulerInterval  = "00:00:00"
	schedulerPolicy    = "read,write"
	schedulerEvent     = "/interface/wireguard/peers/disable"
)

var (
	queueComment     = "Wg Bandwidth Queue: "
	queueName        = "Bandwidth Limit: "
	defaultKeepalive = "25"
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

func NewWGPeer(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor, scheduler *Scheduler, queue *Queue, configGenerator *ConfigGenerator) *WgPeer {
	return &WgPeer{
		db:              db,
		mikrotikAdaptor: mikrotikAdaptor,
		scheduler:       scheduler,
		queue:           queue,
		configGenerator: configGenerator,
		logger:          zap.L().Named("WgPeerService"),
	}
}

func (w *WgPeer) GetPeerKeys() (*schema.PeerKeyResponse, error) {
	privateKey, publicKey, err := w.GenerateKeys()
	if err != nil {
		w.logger.Error("failed to generate peer keys", zap.Error(err))
		return nil, err
	}

	return &schema.PeerKeyResponse{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

func (w *WgPeer) GetPeers() (*[]schema.PeerResponse, error) {
	var peers []model.Peer
	if err := w.db.Order("created_at ASC").Find(&peers).Error; err != nil {
		w.logger.Error("failed to get peers from database", zap.Error(err))
		return nil, err
	}

	var wgPeers []schema.PeerResponse
	for _, peer := range peers {
		wgPeer := w.transformPeerToResponse(peer)
		wgPeers = append(wgPeers, wgPeer)
	}

	return &wgPeers, nil
}

func (w *WgPeer) CreatePeer(req *schema.CreatePeerRequest) (*schema.PeerResponse, error) {
	wgInterface, err := w.mikrotikAdaptor.FetchWgInterface(context.Background(), req.InterfaceId)
	if err != nil {
		return nil, err
	}

	var existingPeer model.Peer
	if err := w.db.Where("allowed_address = ?", req.AllowedAddress).First(&existingPeer).Error; err == nil {
		return nil, fmt.Errorf("allowed address %s is already in use by peer %s", req.AllowedAddress, existingPeer.PeerName)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		w.logger.Error("failed to check allowed address uniqueness", zap.Error(err))
		return nil, err
	}

	wgPeer := &mikrotik.WireGuardPeer{
		Comment:        req.Comment,
		Name:           &req.Name,
		AllowedAddress: &req.AllowedAddress,
		Interface:      &req.Interface,
		PresharedKey:   req.PresharedKey,
		PublicKey:      &req.PublicKey,
	}

	mtPeer, err := w.mikrotikAdaptor.CreateWgPeer(context.Background(), *wgPeer)
	if err != nil {
		return nil, err
	}

	schedulerId, err := w.scheduler.createScheduler(*mtPeer.Name, *mtPeer.ID, req.ExpireTime)
	if err != nil {
		return nil, err
	}

	queueId, err := w.queue.createQueue(*mtPeer.Name, *mtPeer.AllowedAddress, req.DownloadBandwidth, req.UploadBandwidth)
	if err != nil {
		return nil, err
	}

	var timeString = defaultKeepalive
	if req.PersistentKeepAlive != nil {
		parsedTime, err := timehelper.ParseTime(*req.PersistentKeepAlive)
		if err != nil {
			w.logger.Error("failed to parse persistent keepalive time", zap.Error(err))
			return nil, err
		}
		timeString = strconv.Itoa(parsedTime)
	}

	disabled, err := strconv.ParseBool(*mtPeer.Disabled)
	if err != nil {
		w.logger.Error("failed to parse disabled field from Mikrotik peer", zap.Error(err))
		return nil, err
	}

	dbPeer := model.Peer{
		PeerID:              *mtPeer.ID,
		Disabled:            disabled,
		Comment:             mtPeer.Comment,
		PeerName:            *mtPeer.Name,
		PublicKey:           *mtPeer.PublicKey,
		Interface:           *mtPeer.Interface,
		AllowedAddress:      *mtPeer.AllowedAddress,
		Endpoint:            req.Endpoint,
		EndpointPort:        wgInterface.ListenPort,
		PersistentKeepalive: timeString,
		SchedulerID:         schedulerId,
		QueueID:             queueId,
		TrafficLimit:        req.TrafficLimit,
		ExpireTime:          req.ExpireTime,
		DownloadBandwidth:   req.DownloadBandwidth,
		UploadBandwidth:     req.UploadBandwidth,
	}
	if err := w.db.Create(&dbPeer).Error; err != nil {
		w.logger.Error("failed to create peer in database", zap.Error(err))
		return nil, err
	}

	config := fmt.Sprintf(wireguard.Template, req.PrivateKey, dbPeer.AllowedAddress, defaultDns, wgInterface.PublicKey, dbPeer.Endpoint, dbPeer.EndpointPort, allowedIpsIncludeLocal, dbPeer.PersistentKeepalive)

	err = w.configGenerator.BuildPeerConfig(
		config,
		dbPeer.PeerName,
	)
	if err != nil {
		return nil, err
	}

	err = w.qrCodeGenerator.BuildPeerQRCode(config, dbPeer.PeerName)
	if err != nil {
		return nil, err
	}

	transformedPeer := w.transformPeerToResponse(dbPeer)
	return &transformedPeer, nil
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

	// TODO : delete peer config and QR code files

	if err := w.db.Unscoped().Delete(&peer).Error; err != nil {
		w.logger.Error("failed to delete peer from database", zap.Error(err))
		return fmt.Errorf("failed to delete peer from database: %w", err)
	}

	return nil
}

func (w *WgPeer) GetPeersData() (recentOnlinePeers *[]schema.RecentOnlinePeers, totalPeers int, onlinePeers int, err error) {
	peers, err := w.mikrotikAdaptor.FetchWgPeers(context.Background())
	if err != nil {
		return nil, 0, 0, err
	}

	var peerList []struct {
		peer     mikrotik.WireGuardPeer
		duration time.Duration
	}
	var wgPeers []schema.RecentOnlinePeers

	for _, peer := range *peers {
		if peer.LastHandshake != nil {
			duration, err := time.ParseDuration(*peer.LastHandshake)
			if err != nil {
				w.logger.Error("failed to parse last handshake duration", zap.Error(err))
				return nil, 0, 0, err
			}
			peerList = append(peerList, struct {
				peer     mikrotik.WireGuardPeer
				duration time.Duration
			}{peer, duration})
		}
	}

	sort.Slice(peerList, func(i, j int) bool {
		return peerList[i].duration < peerList[j].duration
	})

	count := 0
	for _, item := range peerList {
		if item.duration < 150*time.Second {
			wgPeers = append(wgPeers, schema.RecentOnlinePeers{
				Name:     *item.peer.Name,
				LastSeen: time.Unix(int64(item.duration.Seconds()), 0).UTC().Format("15:04:05")})
			count++
			if count == 5 {
				break
			}
		}
	}

	return &wgPeers, len(*peers), len(peerList), nil
}

func (w *WgPeer) GenerateKeys() (privateKey, publicKey string, err error) {
	privKey, privateKey, err := wireguard.GeneratePrivateKey()
	if err != nil {
		w.logger.Error("failed to generate private key", zap.Error(err))
		return
	}

	publicKey, err = wireguard.GeneratePublicKey(privKey)
	if err != nil {
		w.logger.Error("failed to generate public key from private key", zap.Error(err))
		return "", "", err
	}

	return
}

func (w *WgPeer) updateMikrotikPeer(peerID string, req *schema.UpdatePeerRequest) error {
	wgPeer := mikrotik.WireGuardPeer{}

	if req.Disabled != nil {
		disabledStr := strconv.FormatBool(*req.Disabled)
		wgPeer.Disabled = &disabledStr
	}
	if req.Comment != nil {
		wgPeer.Comment = req.Comment
	}
	if req.Name != nil {
		wgPeer.Name = req.Name
	}
	if req.AllowedAddress != nil {
		wgPeer.AllowedAddress = req.AllowedAddress
	}
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
		return w.scheduler.createScheduler(peer.PeerName, peer.PeerID, req.ExpireTime)
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
		newQueueID, err := w.queue.createQueue(peer.PeerName, peer.AllowedAddress, download, upload)
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

	updateData["disabled"] = req.Disabled
	updateData["comment"] = req.Comment
	updateData["peer_name"] = req.Name
	updateData["allowed_address"] = req.AllowedAddress
	updateData["persistent_keepalive"] = req.PersistentKeepAlive
	updateData["expire_time"] = req.ExpireTime
	updateData["traffic_limit"] = req.TrafficLimit
	updateData["download_bandwidth"] = req.DownloadBandwidth
	updateData["upload_bandwidth"] = req.UploadBandwidth
	updateData["scheduler_id"] = schedulerID
	updateData["queue_id"] = queueID

	return updateData
}

func (w *WgPeer) transformPeerToResponse(peer model.Peer) schema.PeerResponse {
	statuses := w.transformPeerStatus(peer)
	return schema.PeerResponse{
		Id:                peer.ID,
		Disabled:          peer.Disabled,
		Comment:           peer.Comment,
		Name:              peer.PeerName,
		Interface:         peer.Interface,
		AllowedAddress:    peer.AllowedAddress,
		TrafficLimit:      peer.TrafficLimit,
		ExpireTime:        peer.ExpireTime,
		DownloadBandwidth: peer.DownloadBandwidth,
		UploadBandwidth:   peer.UploadBandwidth,
		Status:            statuses,
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
		downloadUsage, err := strconv.Atoi(peer.DownloadUsage)
		if err != nil {
			w.logger.Error("failed to parse download usage", zap.Error(err))
			return peerStatus
		}

		uploadUsage, err := strconv.Atoi(peer.UploadUsage)
		if err != nil {
			w.logger.Error("failed to parse upload usage", zap.Error(err))
			return peerStatus
		}

		trafficLimit, err := strconv.Atoi(*peer.TrafficLimit)
		if err != nil {
			w.logger.Error("failed to parse traffic limit", zap.Error(err))
			return peerStatus
		}

		totalUsedTraffic := downloadUsage + uploadUsage
		if totalUsedTraffic > trafficLimit {
			peerStatus = append(peerStatus, schema.SuspendedPeer)
		}
	}

	return peerStatus
}
