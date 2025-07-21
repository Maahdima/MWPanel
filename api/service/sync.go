package service

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice/model"
)

type SyncService struct {
	db              *gorm.DB
	mikrotikAdaptor *mikrotik.Adaptor
	logger          *zap.Logger
}

func NewSyncService(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor) *SyncService {
	return &SyncService{
		db:              db,
		mikrotikAdaptor: mikrotikAdaptor,
		logger:          zap.L().Named("SyncService"),
	}
}

func (s *SyncService) SyncPeers() error {
	mikrotikPeers, err := s.mikrotikAdaptor.FetchWgPeers(context.Background())
	if err != nil {
		s.logger.Error("failed to get peers from Mikrotik", zap.Error(err))
		return err
	}

	var dbPeers []model.Peer
	if err := s.db.Find(&dbPeers).Error; err != nil {
		s.logger.Error("failed to get peers from database", zap.Error(err))
		return err
	}

	if len(mikrotikPeers) == len(dbPeers) {
		s.logger.Info("no changes detected in peers, skipping sync")
		return nil
	}

	mikrotikPeerMap := make(map[string]mikrotik.WireGuardPeer)
	for _, peer := range mikrotikPeers {
		mikrotikPeerMap[peer.ID] = peer
	}

	dbPeerMap := make(map[string]model.Peer)
	for _, peer := range dbPeers {
		dbPeerMap[peer.PeerID] = peer
	}

	for id, mikrotikPeer := range mikrotikPeerMap {
		var dbPeer model.Peer

		dbPeer.UUID = uuid.New().String()
		dbPeer.PeerID = mikrotikPeer.ID
		dbPeer.Disabled, _ = strconv.ParseBool(mikrotikPeer.Disabled)
		dbPeer.Comment = mikrotikPeer.Comment
		dbPeer.Name = mikrotikPeer.Name
		dbPeer.PublicKey = mikrotikPeer.PublicKey
		dbPeer.Interface = mikrotikPeer.Interface
		dbPeer.AllowedAddress = mikrotikPeer.AllowedAddress
		dbPeer.Endpoint = "127.0.0.1" // TODO: fix this
		dbPeer.EndpointPort = "13231" // TODO: fix this
		dbPeer.PersistentKeepalive = common.DefaultKeepalive

		if err = s.db.Save(&dbPeer).Error; err != nil {
			s.logger.Error("failed to upsert peer", zap.String("id", id), zap.Error(err))
			return err
		}
	}

	for peerId, dbPeer := range dbPeerMap {
		if _, exists := mikrotikPeerMap[peerId]; !exists {
			if err = s.db.Delete(dbPeer.ID).Unscoped().Error; err != nil {
				s.logger.Error("failed to delete peer", zap.String("peerId", peerId), zap.Error(err))
				return err
			}
			s.logger.Info("deleted stale peer from DB", zap.String("peerId", peerId))
		}
	}

	s.logger.Info("successfully synced peers with mikrotik")
	return nil
}

func (s *SyncService) SyncInterfaces() error {
	mikrotikInterfaces, err := s.mikrotikAdaptor.FetchWgInterfaces(context.Background())
	if err != nil {
		s.logger.Error("failed to get interfaces from Mikrotik", zap.Error(err))
		return err
	}

	var dbInterfaces []model.Interface
	if err := s.db.Find(&dbInterfaces).Error; err != nil {
		s.logger.Error("failed to get interfaces from database", zap.Error(err))
		return err
	}

	if len(mikrotikInterfaces) == len(dbInterfaces) {
		s.logger.Info("no changes detected in interfaces, skipping sync")
		return nil
	}

	mikrotikInterfaceMap := make(map[string]mikrotik.WireGuardInterface)
	for _, iface := range mikrotikInterfaces {
		mikrotikInterfaceMap[iface.ID] = iface
	}

	dbInterfaceMap := make(map[string]model.Interface)
	for _, iface := range dbInterfaces {
		dbInterfaceMap[iface.InterfaceID] = iface
	}

	for id, mikrotikIface := range mikrotikInterfaceMap {
		var dbIface model.Interface

		dbIface.InterfaceID = mikrotikIface.ID
		dbIface.Disabled, _ = strconv.ParseBool(mikrotikIface.Disabled)
		dbIface.Comment = mikrotikIface.Comment
		dbIface.Name = mikrotikIface.Name
		dbIface.ListenPort = mikrotikIface.ListenPort

		if err = s.db.Save(dbIface).Error; err != nil {
			s.logger.Error("failed to upsert interface", zap.String("id", id), zap.Error(err))
			return err
		}
	}

	for ifaceId, dbIface := range dbInterfaceMap {
		if _, exists := mikrotikInterfaceMap[ifaceId]; !exists {
			if err = s.db.Delete(dbIface.ID).Unscoped().Error; err != nil {
				s.logger.Error("failed to delete interface", zap.String("interfaceId", ifaceId), zap.Error(err))
				return err
			}
			s.logger.Info("deleted stale interface from DB", zap.String("interfaceId", ifaceId))
		}
	}

	s.logger.Info("successfully synced interfaces with mikrotik")
	return nil
}
