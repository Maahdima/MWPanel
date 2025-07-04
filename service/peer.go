package service

import (
	"context"
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
	queueComment = "Wg Bandwidth Queue: "
	queueName    = "Bandwidth Limit: "
)

type WgPeer struct {
	db              *gorm.DB
	mikrotikAdaptor *mikrotik.Adaptor
	configGenerator *ConfigGenerator
	qrCodeGenerator *QRCodeGenerator
	logger          *zap.Logger
}

func NewWGPeer(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor, configGenerator *ConfigGenerator) *WgPeer {
	return &WgPeer{
		db:              db,
		mikrotikAdaptor: mikrotikAdaptor,
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
	if err := w.db.Find(&peers).Error; err != nil {
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
	var allowedAddress string

	wgInterface, err := w.mikrotikAdaptor.FetchWgInterface(context.Background(), req.InterfaceId)
	if err != nil {
		return nil, err
	}

	if req.AllowedAddress == nil {
		allowedAddress = "0.0.0.0/0"
	} else {
		allowedAddress = *req.AllowedAddress
	}

	wgPeer := &mikrotik.WireGuardPeer{
		Comment:        req.Comment,
		Name:           &req.Name,
		AllowedAddress: &allowedAddress,
		Interface:      &req.Interface,
		PresharedKey:   req.PresharedKey,
		PublicKey:      &req.PublicKey,
	}

	mtPeer, err := w.mikrotikAdaptor.CreateWgPeer(context.Background(), *wgPeer)
	if err != nil {
		return nil, err
	}

	schedulerId, err := w.createPeerScheduler(mtPeer, req.ExpireTime)
	if err != nil {
		return nil, err
	}

	queueId, err := w.createPeerQueue(mtPeer, req.DownloadBandwidth, req.UploadBandwidth)
	if err != nil {
		return nil, err
	}

	parsedTime, err := timehelper.ParseTime(*req.PersistentKeepAlive)
	timeString := strconv.Itoa(parsedTime)

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

	_, err := w.mikrotikAdaptor.UpdateWgPeer(context.Background(), peer.PeerID, wgPeer)
	if err != nil {
		w.logger.Error("failed to update wireguard peer in Mikrotik", zap.Error(err))
		return nil, err
	}

	if err := w.db.Model(&peer).Updates(req).Error; err != nil {
		return nil, err
	}

	transformed := w.transformPeerToResponse(peer)
	return &transformed, nil
}

func (w *WgPeer) DeletePeer(id string) error {
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

func (w *WgPeer) createPeerScheduler(peer *mikrotik.WireGuardPeer, expireTime *string) (*string, error) {
	if expireTime == nil {
		return nil, nil
	}

	scheduler := mikrotik.Scheduler{
		Comment:   schedulerComment + *peer.Name,
		Name:      schedulerName + *peer.Name,
		StartDate: *expireTime,
		StartTime: schedulerStartTime,
		Interval:  schedulerInterval,
		Policy:    schedulerPolicy,
		OnEvent:   schedulerEvent + *peer.ID,
	}

	createdScheduler, err := w.mikrotikAdaptor.CreateScheduler(context.Background(), scheduler)
	if err != nil {
		w.logger.Error("failed to create scheduler for WireGuard peer", zap.Error(err))
		return nil, err
	}

	return &createdScheduler.ID, nil
}

func (w *WgPeer) createPeerQueue(peer *mikrotik.WireGuardPeer, downloadLimit, uploadLimit *string) (*string, error) {
	if downloadLimit == nil || uploadLimit == nil {
		return nil, nil
	}

	wgQueue := mikrotik.Queue{
		Comment:  queueComment + *peer.Name,
		Name:     queueName + *peer.Name,
		Target:   *peer.AllowedAddress,
		MaxLimit: *downloadLimit + "/" + *uploadLimit,
	}

	createdQueue, err := w.mikrotikAdaptor.CreateSimpleQueue(context.Background(), wgQueue)
	if err != nil {
		w.logger.Error("failed to create simple queue for WireGuard peer", zap.Error(err))
		return nil, err
	}

	return &createdQueue.ID, nil
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
