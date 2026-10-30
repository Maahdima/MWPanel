package service

import (
	"context"
	"math/big"
	"regexp"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice/model"
)

var peerBandwidthPattern = regexp.MustCompile(`^(?:[0-9]{1,9})(?:\.[0-9]{1,3})?[KMG]$`)

// Bandwidth uses the existing RouterOS decimal K/M/G bits-per-second units.
// Nil is unlimited; bounded positive values avoid malformed RouterOS commands.
func validPeerBandwidth(value *string) bool {
	if value == nil {
		return true
	}

	if !peerBandwidthPattern.MatchString(*value) {
		return false
	}

	amount, ok := new(big.Rat).SetString((*value)[:len(*value)-1])

	if !ok || amount.Sign() <= 0 {
		return false
	}

	factor := map[byte]int64{'K': 1_000, 'M': 1_000_000, 'G': 1_000_000_000}[(*value)[len(*value)-1]]

	return amount.Mul(amount, new(big.Rat).SetInt64(factor)).Cmp(new(big.Rat).SetInt64(1_000_000_000_000)) <= 0
}

func (s *PeerPurchases) applyBandwidth(ctx context.Context, router *mikrotik.Adaptor, peer model.Peer, download, upload *string) (*string, error) {
	queueID := peer.QueueID

	if queueID == nil {
		// A deterministic name recovers a successful creation after an ambiguous response.
		queues, err := router.FetchSimpleQueues(ctx)

		if err != nil {
			return nil, err
		}

		for _, queue := range queues {
			if queue.Name == common.QueueName+peer.UUID {
				if queueID != nil || queue.Target == nil || *queue.Target != peer.AllowedAddress {
					return nil, ErrRouterPeerIdentity
				}

				id := queue.ID

				queueID = &id
			}
		}

		if queueID == nil {
			bound, err := s.peers.ForServer(peer.ServerID)

			if err != nil {
				return nil, err
			}

			queueID, err = bound.queue.createQueueContext(ctx, peer.UUID, peer.AllowedAddress, download, upload)

			if err != nil {
				return nil, err
			}
		}
	}

	queue, err := router.FetchSimpleQueue(ctx, *queueID)

	if err != nil {
		return nil, err
	}

	if queue.Target == nil || *queue.Target != peer.AllowedAddress || queue.Disabled == "true" {
		return nil, ErrRouterPeerIdentity
	}

	down, up := "0", "0"

	if download != nil {
		down = *download
	}

	if upload != nil {
		up = *upload
	}

	limit := down + "/" + up

	if _, err = router.UpdateSimpleQueue(ctx, *queueID, mikrotik.Queue{MaxLimit: &limit}); err != nil {
		return nil, err
	}

	verified, err := router.FetchSimpleQueue(ctx, *queueID)

	if err != nil {
		return nil, err
	}

	if verified.Target == nil || *verified.Target != peer.AllowedAddress || verified.MaxLimit == nil || *verified.MaxLimit != limit || verified.Disabled == "true" {
		return nil, ErrRouterSyncPending
	}

	return queueID, nil
}
