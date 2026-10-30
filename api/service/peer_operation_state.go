package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils/httphelper"
	"gorm.io/gorm"
	"net"
	"net/http"
	"strings"
	"time"
)

type provisioningFailure struct {
	stage string
	cause error
}

func (e *provisioningFailure) Error() string { return e.stage + ": " + e.cause.Error() }

func (e *provisioningFailure) Unwrap() error { return e.cause }

// Never publish router response bodies, URLs or credentials in status messages.
func provisioningMessage(err error) string {
	stage := "router synchronization"

	var failure *provisioningFailure

	if errors.As(err, &failure) {
		stage = failure.stage
	}

	reason := "Verification failed; check router connectivity, resources and server logs."

	var status *httphelper.StatusError

	var network net.Error

	switch {
	case errors.As(err, &status):
		if status.StatusCode == http.StatusUnauthorized || status.StatusCode == http.StatusForbidden {
			reason = "RouterOS rejected access; check REST credentials and permissions."
		} else {
			reason = fmt.Sprintf("RouterOS returned HTTP %d; verify the resource and retry.", status.StatusCode)
		}

	case errors.Is(err, context.DeadlineExceeded), errors.As(err, &network):
		reason = "Router connection failed or timed out; automatic retry is scheduled."

	case errors.Is(err, ErrRouterPeerIdentity):
		reason = "Router object identity or configuration does not match; review the assigned interface and resources."

	case errors.Is(err, gorm.ErrRecordNotFound):
		reason = "An assigned router or resource is missing or inactive."

	case errors.Is(err, ErrPeerRestricted):
		reason = "Another peer or account restriction prevents activation."
	}

	return stage + ": " + reason
}

func safeStoredOperationError(message string) string {
	for _, prefix := range []string{"interface: ", "peer: ", "queue: ", "scheduler: ", "configuration files: ", "accounting: ", "router synchronization: "} {
		if strings.HasPrefix(message, prefix) {
			return message
		}
	}

	if message != "" {
		return "Verification failed; check router connectivity, resources and server logs."
	}

	return ""
}

// Retry only releases backoff. It never releases a live lease or changes money,
// provisioning success, ownership, manual disablement or other restrictions.
func (s *Resellers) RetryPeer(actor Principal, resellerID, peerID uint) error {
	if !billingAdmin(actor) {
		return ErrRoleForbidden
	}

	return NewBilling(s.db).accountTransaction(resellerID, func(tx *gorm.DB, _ *model.BillingAccount) error {
		var peer model.Peer

		if err := tx.Where("reseller_id = ?", resellerID).First(&peer, peerID).Error; err != nil {
			return err
		}

		now := NewBilling(tx).now().UTC().Unix()

		if peer.MigrationID != nil || peer.SyncLeaseUntil > now {
			return ErrPendingPeerOperation
		}

		if peer.PurchaseID != nil {
			result := tx.Model(&model.PeerPurchase{}).Where("id = ? AND reseller_id = ? AND peer_id = ? AND state = ? AND lease_until <= ?", *peer.PurchaseID, resellerID, peerID, "pending", now).UpdateColumn("next_attempt_at", 0)

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected != 1 {
				return ErrPendingPeerOperation
			}
		} else if peer.ProvisioningBlocked {
			return ErrPendingPeerOperation
		}

		if err := tx.Model(&peer).Updates(map[string]any{"sync_next_attempt_at": 0, "synced_disabled": nil}).Error; err != nil {
			return err
		}

		return billingAudit(tx, resellerID, actor.AccountID, nil, "peer_retry_requested", fmt.Sprintf("peer:%d", peerID))
	})
}

func (w *WgPeer) peerOperationStatus(peer model.Peer) (string, string, int64) {
	status := "ready"

	message := safeStoredOperationError(peer.SyncError)

	next := peer.SyncNextAttemptAt

	if peer.ProvisioningBlocked {
		status = "pending"
	}

	if peer.PurchaseID != nil {
		var op model.PeerPurchase

		if err := w.db.Where("id = ? AND peer_id = ?", *peer.PurchaseID, peer.ID).First(&op).Error; err != nil {
			return "error", "The persisted provisioning operation is missing; administrator review is required.", 0
		}

		next = op.NextAttemptAt

		if op.LastError != "" {
			if peer.ProvisioningBlocked {
				status = "error"
			}

			return status, safeStoredOperationError(op.LastError), next
		}

		if peer.ProvisioningBlocked && op.LeaseUntil > time.Now().UTC().Unix() {
			status = "running"
		}
	} else if peer.ProvisioningBlocked {
		return "error", "No provisioning operation is linked; administrator review is required.", 0
	}

	return status, message, next
}
