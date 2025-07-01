package service

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"mikrotik-wg-go/adaptor/mikrotik"
	"mikrotik-wg-go/api/schema"
	"mikrotik-wg-go/dataservice/db"
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
	db              *db.Queries
	mikrotikAdaptor *mikrotik.Adaptor
	configGenerator *ConfigGenerator
	qrCodeGenerator *QRCodeGenerator
	logger          *zap.Logger
}

func NewWGPeer(db *db.Queries, mikrotikAdaptor *mikrotik.Adaptor, configGenerator *ConfigGenerator) *WgPeer {
	return &WgPeer{
		db:              db,
		mikrotikAdaptor: mikrotikAdaptor,
		configGenerator: configGenerator,
		logger:          zap.L().Named("WgPeerService"),
	}
}

func (w *WgPeer) GetPeers() (*[]schema.WgPeerResponse, error) {
	peers, err := w.db.ListPeers(context.Background())
	if err != nil {
		w.logger.Error("failed to list wireguard peers from database", zap.Error(err))
		return nil, err
	}

	var wgPeers []schema.WgPeerResponse
	for _, peer := range peers {
		wgPeer := w.transformPeerToResponse(peer)
		wgPeers = append(wgPeers, wgPeer)
	}

	return &wgPeers, nil
}

func (w *WgPeer) CreatePeer(req *schema.WgPeerRequest) (*schema.WgPeerResponse, error) {
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
		Name:           req.Name,
		AllowedAddress: allowedAddress,
		Interface:      req.Interface,
		PresharedKey:   req.PresharedKey,
		PublicKey:      req.PublicKey,
	}

	peer, err := w.mikrotikAdaptor.CreateWgPeer(context.Background(), *wgPeer)
	if err != nil {
		return nil, err
	}

	schedulerId, err := w.createPeerScheduler(peer, req.ExpireTime)
	if err != nil {
		return nil, err
	}

	queueId, err := w.createPeerQueue(peer, req.DownloadBandwidth, req.UploadBandwidth)
	if err != nil {
		return nil, err
	}

	parsedTime, err := timehelper.ParseTime(*req.PersistentKeepAlive)
	createdPeer, err := w.db.CreatePeer(context.Background(), db.CreatePeerParams{
		PeerID:              peer.ID,
		Disabled:            peer.Disabled,
		Comment:             peer.Comment,
		PeerName:            peer.Name,
		PublicKey:           peer.PublicKey,
		Interface:           peer.Interface,
		AllowedAddress:      peer.AllowedAddress,
		Endpoint:            req.Endpoint,
		EndpointPort:        wgInterface.ListenPort,
		PersistentKeepalive: strconv.Itoa(parsedTime),
		SchedulerID:         schedulerId,
		QueueID:             queueId,
		TrafficLimit:        req.TrafficLimit,
		ExpireTime:          req.ExpireTime,
		DownloadBandwidth:   req.DownloadBandwidth,
		UploadBandwidth:     req.UploadBandwidth,
	})
	if err != nil {
		w.logger.Error("failed to insert peer into database", zap.Error(err))
		return nil, err
	}

	config := fmt.Sprintf(wireguard.Template, req.PrivateKey, createdPeer.AllowedAddress, defaultDns, wgInterface.PublicKey, createdPeer.Endpoint, createdPeer.EndpointPort, allowedIpsIncludeLocal, createdPeer.PersistentKeepalive)

	err = w.configGenerator.BuildPeerConfig(
		config,
		createdPeer.PeerName,
	)
	if err != nil {
		return nil, err
	}

	err = w.qrCodeGenerator.BuildPeerQRCode(config, createdPeer.PeerName)
	if err != nil {
		return nil, err
	}

	transformedPeer := w.transformPeerToResponse(createdPeer)
	return &transformedPeer, nil
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
				Name:     item.peer.Name,
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
		Comment:   schedulerComment + peer.Name,
		Name:      schedulerName + peer.Name,
		StartDate: *expireTime,
		StartTime: schedulerStartTime,
		Interval:  schedulerInterval,
		Policy:    schedulerPolicy,
		OnEvent:   schedulerEvent + peer.ID,
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
		Comment:  queueComment + peer.Name,
		Name:     queueName + peer.Name,
		Target:   peer.AllowedAddress,
		MaxLimit: *downloadLimit + "/" + *uploadLimit,
	}

	createdQueue, err := w.mikrotikAdaptor.CreateSimpleQueue(context.Background(), wgQueue)
	if err != nil {
		w.logger.Error("failed to create simple queue for WireGuard peer", zap.Error(err))
		return nil, err
	}

	return &createdQueue.ID, nil
}

func (w *WgPeer) transformPeerToResponse(peer db.WgPeer) schema.WgPeerResponse {
	return schema.WgPeerResponse{
		Id:                strconv.FormatInt(peer.ID, 10),
		Disabled:          peer.Disabled,
		Comment:           peer.Comment,
		Name:              peer.PeerName,
		Interface:         peer.Interface,
		AllowedAddress:    peer.AllowedAddress,
		TrafficLimit:      peer.TrafficLimit,
		ExpireTime:        peer.ExpireTime,
		DownloadBandwidth: peer.DownloadBandwidth,
		UploadBandwidth:   peer.UploadBandwidth,
	}
}
