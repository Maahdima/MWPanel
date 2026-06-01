package service

import (
	"errors"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/http/schema"
)

func (s *SyncService) GetSyncInterfaces() ([]schema.SyncInterfacePreviewResponse, error) {
	ifaces, err := s.fetchMikrotikInterfaces()
	if err != nil {
		return nil, err
	}

	return mapSyncInterfaces(ifaces), nil
}

func (s *SyncService) GetSyncPeers(interfaceName string) ([]schema.SyncPeerPreviewResponse, error) {
	peers, err := s.fetchMikrotikPeers()
	if err != nil {
		return nil, err
	}

	if interfaceName != "" {
		filtered := make([]mikrotik.WireGuardPeer, 0, len(peers))
		for _, peer := range peers {
			if peer.Interface == interfaceName {
				filtered = append(filtered, peer)
			}
		}
		peers = filtered
	}

	return mapSyncPeers(peers), nil
}

func (s *SyncService) SyncSelectedInterfaces(interfaceIDs []string) error {
	if len(interfaceIDs) == 0 {
		return errors.New("no interfaces selected")
	}

	mikrotikIfaces, err := s.fetchMikrotikInterfaces()
	if err != nil {
		return err
	}

	selectedIfaces := filterMikrotikInterfacesByID(mikrotikIfaces, interfaceIDs)
	if len(selectedIfaces) == 0 {
		s.logger.Info("no matching interfaces selected for sync")
		return nil
	}

	dbIfaces, err := s.fetchDBInterfaces()
	if err != nil {
		return err
	}

	return s.syncNewAndUpdatedInterfaces(s.mapMikrotikInterfaces(selectedIfaces), s.mapDBInterfaces(dbIfaces))
}

func (s *SyncService) SyncSelectedPeers(peerIDs []string) error {
	if len(peerIDs) == 0 {
		return errors.New("no peers selected")
	}

	mikrotikPeers, err := s.fetchMikrotikPeers()
	if err != nil {
		return err
	}

	selectedPeers := filterMikrotikPeersByID(mikrotikPeers, peerIDs)
	if len(selectedPeers) == 0 {
		s.logger.Info("no matching peers selected for sync")
		return nil
	}

	dbPeers, err := s.fetchDBPeers()
	if err != nil {
		return err
	}

	return s.syncNewAndUpdatedPeers(s.mapMikrotikPeers(selectedPeers), s.mapDBPeers(dbPeers))
}

func parseOptionalBool(value *string) bool {
	if value == nil {
		return false
	}

	return parseBool(*value)
}

func mapSyncInterfaces(ifaces []mikrotik.WireGuardInterface) []schema.SyncInterfacePreviewResponse {
	result := make([]schema.SyncInterfacePreviewResponse, 0, len(ifaces))
	for _, iface := range ifaces {
		result = append(result, schema.SyncInterfacePreviewResponse{
			ID:         iface.ID,
			Disabled:   parseBool(iface.Disabled),
			Comment:    iface.Comment,
			Name:       iface.Name,
			ListenPort: iface.ListenPort,
			MTU:        iface.MTU,
			IsRunning:  parseOptionalBool(iface.Running),
		})
	}

	return result
}

func mapSyncPeers(peers []mikrotik.WireGuardPeer) []schema.SyncPeerPreviewResponse {
	result := make([]schema.SyncPeerPreviewResponse, 0, len(peers))
	for _, peer := range peers {
		result = append(result, schema.SyncPeerPreviewResponse{
			ID:             peer.ID,
			Disabled:       parseBool(peer.Disabled),
			Comment:        peer.Comment,
			Name:           peer.Name,
			Interface:      peer.Interface,
			AllowedAddress: peer.AllowedAddress,
		})
	}

	return result
}

func filterMikrotikInterfacesByID(ifaces []mikrotik.WireGuardInterface, ids []string) []mikrotik.WireGuardInterface {
	selected := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		selected[id] = struct{}{}
	}

	result := make([]mikrotik.WireGuardInterface, 0, len(ids))
	for _, iface := range ifaces {
		if _, ok := selected[iface.ID]; ok {
			result = append(result, iface)
		}
	}

	return result
}

func filterMikrotikPeersByID(peers []mikrotik.WireGuardPeer, ids []string) []mikrotik.WireGuardPeer {
	selected := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		selected[id] = struct{}{}
	}

	result := make([]mikrotik.WireGuardPeer, 0, len(ids))
	for _, peer := range peers {
		if _, ok := selected[peer.ID]; ok {
			result = append(result, peer)
		}
	}

	return result
}
