package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils/wireguard"
)

type SyncService struct {
	db              *gorm.DB
	mikrotikAdaptor *mikrotik.Adaptor
	configService   *ConfigGenerator
	qrCodeService   *QRCodeGenerator
	logger          *zap.Logger
}

func NewSyncService(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor, configService *ConfigGenerator, qrCodeService *QRCodeGenerator) *SyncService {
	return &SyncService{
		db:              db,
		mikrotikAdaptor: mikrotikAdaptor,
		configService:   configService,
		qrCodeService:   qrCodeService,
		logger:          zap.L().Named("SyncService"),
	}
}

func (s *SyncService) SyncPeers() error {
	mikrotikPeers, err := s.fetchMikrotikPeers()
	if err != nil {
		return err
	}

	dbPeers, err := s.fetchDBPeers()
	if err != nil {
		return err
	}

	if len(mikrotikPeers) == len(dbPeers) {
		s.logger.Info("no changes detected in peers, skipping sync")
		return nil
	}

	mikrotikMap := s.mapMikrotikPeers(mikrotikPeers)
	dbMap := s.mapDBPeers(dbPeers)

	if err := s.syncNewAndUpdatedPeers(mikrotikMap); err != nil {
		return err
	}

	if err := s.removeStalePeers(mikrotikMap, dbMap); err != nil {
		return err
	}

	s.logger.Info("successfully synced peers with mikrotik")
	return nil
}

func (s *SyncService) SyncInterfaces() error {
	mikrotikIfaces, err := s.fetchMikrotikInterfaces()
	if err != nil {
		return err
	}

	dbIfaces, err := s.fetchDBInterfaces()
	if err != nil {
		return err
	}

	if len(mikrotikIfaces) == len(dbIfaces) {
		s.logger.Info("no changes detected in interfaces, skipping sync")
		return nil
	}

	mikrotikMap := s.mapMikrotikInterfaces(mikrotikIfaces)
	dbMap := s.mapDBInterfaces(dbIfaces)

	if err := s.syncNewAndUpdatedInterfaces(mikrotikMap); err != nil {
		return err
	}

	if err := s.removeStaleInterfaces(mikrotikMap, dbMap); err != nil {
		return err
	}

	s.logger.Info("successfully synced interfaces with mikrotik")
	return nil
}

func (s *SyncService) fetchMikrotikPeers() ([]mikrotik.WireGuardPeer, error) {
	peers, err := s.mikrotikAdaptor.FetchWgPeers(context.Background())
	if err != nil {
		s.logger.Error("failed to get peers from Mikrotik", zap.Error(err))
	}
	return peers, err
}

func (s *SyncService) fetchDBPeers() ([]model.Peer, error) {
	var peers []model.Peer
	err := s.db.Find(&peers).Error
	if err != nil {
		s.logger.Error("failed to get peers from database", zap.Error(err))
	}
	return peers, err
}

func (s *SyncService) mapMikrotikPeers(peers []mikrotik.WireGuardPeer) map[string]mikrotik.WireGuardPeer {
	m := make(map[string]mikrotik.WireGuardPeer)
	for _, p := range peers {
		m[p.ID] = p
	}
	return m
}

func (s *SyncService) mapDBPeers(peers []model.Peer) map[string]model.Peer {
	m := make(map[string]model.Peer)
	for _, p := range peers {
		m[p.PeerID] = p
	}
	return m
}

func (s *SyncService) syncNewAndUpdatedPeers(peers map[string]mikrotik.WireGuardPeer) error {
	for id, peer := range peers {
		if peer.PrivateKey == nil {
			s.logger.Error("missing private key", zap.String("peer", id))
			continue
		}

		server, err := s.fetchServer()
		if err != nil {
			return err
		}

		dbIface, err := s.fetchInterface(peer.Interface, id)
		if err != nil {
			return err
		}

		dbPeer := s.buildDBPeer(peer, server, dbIface)
		config := s.buildConfig(peer, dbPeer, dbIface)

		if err := s.configService.BuildPeerConfig(config, dbPeer.UUID); err != nil {
			s.logger.Error("failed to build peer config", zap.String("id", id), zap.Error(err))
			return err
		}

		if err := s.qrCodeService.BuildPeerQRCode(config, dbPeer.UUID); err != nil {
			s.logger.Error("failed to build QR code", zap.String("id", id), zap.Error(err))
			return err
		}

		if err := s.db.Save(&dbPeer).Error; err != nil {
			s.logger.Error("failed to upsert peer", zap.String("id", id), zap.Error(err))
			return err
		}
	}
	return nil
}

func (s *SyncService) removeStalePeers(mikrotikMap map[string]mikrotik.WireGuardPeer, dbMap map[string]model.Peer) error {
	for id, peer := range dbMap {
		if _, found := mikrotikMap[id]; !found {
			if err := s.db.Delete(peer.ID).Unscoped().Error; err != nil {
				s.logger.Error("failed to delete peer", zap.String("peerId", id), zap.Error(err))
				return err
			}
			s.logger.Info("deleted stale peer from DB", zap.String("peerId", id))
		}
	}
	return nil
}

func (s *SyncService) fetchServer() (model.Server, error) {
	var server model.Server
	err := s.db.First(&server).Error
	if err != nil {
		s.logger.Error("failed to get server", zap.Error(err))
	}
	return server, err
}

func (s *SyncService) fetchInterface(name, peerID string) (model.Interface, error) {
	var iface model.Interface
	err := s.db.Where("name = ?", name).First(&iface).Error
	if err != nil {
		if errors.Is(gorm.ErrRecordNotFound, err) {
			s.logger.Error("interface not found", zap.String("peerId", peerID), zap.String("interfaceId", name))
			return iface, fmt.Errorf("interface %s not found for peer %s", name, peerID)
		}
		s.logger.Error("failed to get interface", zap.String("peerId", peerID), zap.Error(err))
	}
	return iface, err
}

func (s *SyncService) buildDBPeer(peer mikrotik.WireGuardPeer, server model.Server, iface model.Interface) model.Peer {
	return model.Peer{
		UUID:                uuid.New().String(),
		PeerID:              peer.ID,
		Disabled:            parseBool(peer.Disabled),
		Comment:             peer.Comment,
		Name:                peer.Name,
		PrivateKey:          *peer.PrivateKey,
		PublicKey:           peer.PublicKey,
		Interface:           peer.Interface,
		AllowedAddress:      peer.AllowedAddress,
		Endpoint:            server.IPAddress,
		EndpointPort:        iface.ListenPort,
		PersistentKeepalive: common.DefaultKeepalive,
	}
}

func (s *SyncService) buildConfig(peer mikrotik.WireGuardPeer, dbPeer model.Peer, iface model.Interface) string {
	return fmt.Sprintf(wireguard.Template,
		*peer.PrivateKey,
		dbPeer.AllowedAddress,
		common.DefaultDns,
		iface.PublicKey,
		dbPeer.Endpoint,
		dbPeer.EndpointPort,
		common.AllowedIpsIncludeLocal,
		dbPeer.PersistentKeepalive,
	)
}

func (s *SyncService) fetchMikrotikInterfaces() ([]mikrotik.WireGuardInterface, error) {
	ifaces, err := s.mikrotikAdaptor.FetchWgInterfaces(context.Background())
	if err != nil {
		s.logger.Error("failed to get interfaces from Mikrotik", zap.Error(err))
	}
	return ifaces, err
}

func (s *SyncService) fetchDBInterfaces() ([]model.Interface, error) {
	var ifaces []model.Interface
	err := s.db.Find(&ifaces).Error
	if err != nil {
		s.logger.Error("failed to get interfaces from database", zap.Error(err))
	}
	return ifaces, err
}

func (s *SyncService) mapMikrotikInterfaces(ifaces []mikrotik.WireGuardInterface) map[string]mikrotik.WireGuardInterface {
	m := make(map[string]mikrotik.WireGuardInterface)
	for _, iface := range ifaces {
		m[iface.ID] = iface
	}
	return m
}

func (s *SyncService) mapDBInterfaces(ifaces []model.Interface) map[string]model.Interface {
	m := make(map[string]model.Interface)
	for _, iface := range ifaces {
		m[iface.InterfaceID] = iface
	}
	return m
}

func (s *SyncService) syncNewAndUpdatedInterfaces(ifaceMap map[string]mikrotik.WireGuardInterface) error {
	for id, mikrotikIface := range ifaceMap {
		dbIface := model.Interface{
			InterfaceID: mikrotikIface.ID,
			Disabled:    parseBool(mikrotikIface.Disabled),
			Comment:     mikrotikIface.Comment,
			Name:        mikrotikIface.Name,
			PrivateKey:  mikrotikIface.PrivateKey,
			PublicKey:   mikrotikIface.PublicKey,
			ListenPort:  mikrotikIface.ListenPort,
		}

		if err := s.db.Save(&dbIface).Error; err != nil {
			s.logger.Error("failed to upsert interface", zap.String("id", id), zap.Error(err))
			return err
		}
	}
	return nil
}

func (s *SyncService) removeStaleInterfaces(mikrotikMap map[string]mikrotik.WireGuardInterface, dbMap map[string]model.Interface) error {
	for id, dbIface := range dbMap {
		if _, exists := mikrotikMap[id]; !exists {
			if err := s.db.Delete(dbIface.ID).Unscoped().Error; err != nil {
				s.logger.Error("failed to delete interface", zap.String("interfaceId", id), zap.Error(err))
				return err
			}
			s.logger.Info("deleted stale interface from DB", zap.String("interfaceId", id))
		}
	}
	return nil
}

func parseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}
