package traffic

import (
	"context"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/dataservice/model"
	"go.uber.org/zap"
)

func (c *Calculator) forServer(serverID uint) (*Calculator, error) {
	adaptor, err := c.mikrotikAdaptor.ForServer(serverID)

	if err != nil {
		return nil, err
	}

	selected := *c

	selected.mikrotikAdaptor = adaptor

	return &selected, nil
}

// observedPeers keeps router-local IDs partitioned by server. Unreachable routers
// are absent, so session tracking never invents disconnects from a router outage.
func (c *Calculator) observedPeers(peers []model.Peer) map[uint]map[string]mikrotik.WireGuardPeer {
	result := make(map[uint]map[string]mikrotik.WireGuardPeer)

	attempted := make(map[uint]bool)

	for _, peer := range peers {
		if attempted[peer.ServerID] {
			continue
		}

		attempted[peer.ServerID] = true

		adaptor, err := c.mikrotikAdaptor.ForServer(peer.ServerID)

		if err != nil {
			continue
		}

		snapshots, err := adaptor.FetchWgPeers(context.Background())

		if err != nil {
			c.logger.Warn("router unavailable for peer observations", zap.Uint("server_id", peer.ServerID), zap.Error(err))

			continue
		}

		result[peer.ServerID] = make(map[string]mikrotik.WireGuardPeer, len(snapshots))

		for _, snapshot := range snapshots {
			result[peer.ServerID][snapshot.ID] = snapshot
		}
	}

	return result
}
