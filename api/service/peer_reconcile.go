package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ErrRouterSyncPending = errors.New("router synchronization is pending")
var ErrBilledOperationRequired = errors.New("use a billed reseller operation for this change")
var ErrPeerRestricted = errors.New("peer has an active expiration, quota, billing or administrative restriction")
var ErrRouterPeerIdentity = errors.New("stored router object identity does not match the peer")

type PeerReconciler struct {
	runMu         sync.Mutex
	accountCursor uint
	db            *gorm.DB
	adaptor       *mikrotik.Adaptor
	billing       *Billing
	now           func() time.Time
	cursor        uint
}

func NewPeerReconciler(db *gorm.DB, adaptor *mikrotik.Adaptor, billing *Billing) *PeerReconciler {
	return &PeerReconciler{db: db, adaptor: adaptor, billing: billing, now: time.Now}
}

func peerOtherRestriction(tx *gorm.DB, peer model.Peer, now time.Time) (bool, error) {
	if utils.IsPeerExpired(peer.ExpireTime, now) || (peer.TrafficLimit != nil && peer.DownloadUsage+peer.UploadUsage >= *peer.TrafficLimit) {
		return true, nil
	}
	if peer.ResellerID != nil {
		var reseller model.Reseller
		if err := tx.First(&reseller, *peer.ResellerID).Error; err != nil {
			return false, err
		}
		var account model.BillingAccount
		if err := tx.Where("reseller_id = ?", *peer.ResellerID).First(&account).Error; err != nil {
			return false, err
		}
		return reseller.AdminSuspended || account.BillingSuspended, nil
	}
	return false, nil
}

// Run is bounded per tick. It revisits every persisted peer after restart and
// uses durable leases/backoff; a failed router does not stop other routers.
func (r *PeerReconciler) Run() {
	if !r.runMu.TryLock() {
		return
	}
	defer r.runMu.Unlock()
	var accounts []model.BillingAccount
	if err := r.db.Where("id > ?", r.accountCursor).Order("id ASC").Limit(200).Find(&accounts).Error; err != nil {
		return
	}
	if len(accounts) == 0 {
		r.accountCursor = 0
	}
	for _, account := range accounts {
		r.accountCursor = account.ID
		if _, err := r.billing.ReconcileAccount(account.ResellerID); err != nil {
			zap.L().Warn("billing reconciliation failed", zap.Uint("reseller_id", account.ResellerID), zap.Error(err))
		}
	}

	var peers []model.Peer
	if err := r.db.Where("id > ?", r.cursor).Order("id ASC").Limit(200).Find(&peers).Error; err != nil {
		return
	}
	if len(peers) == 0 {
		r.cursor = 0
		return
	}
	for _, peer := range peers {
		r.cursor = peer.ID
		if err := r.ReconcilePeer(peer.ID); err != nil {
			zap.L().Warn("peer reconciliation pending", zap.Uint("peer_id", peer.ID), zap.Error(err))
		}
	}
}

func (r *PeerReconciler) prepare(id uint) (model.Peer, bool, error) {
	var peer model.Peer
	claimed := false
	now := r.now().UTC().Unix()
	err := r.db.Transaction(func(tx *gorm.DB) error {
		lock := tx.Model(&model.Peer{}).Where("id = ?", id).UpdateColumn("restriction_revision", gorm.Expr("restriction_revision"))
		if lock.Error != nil {
			return lock.Error
		}
		if lock.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.First(&peer, id).Error; err != nil {
			return err
		}
		manual := peer.ManualDisabled
		if !peer.RestrictionsInitialized {
			manual = peer.Disabled
		}
		expiry := utils.IsPeerExpired(peer.ExpireTime, r.now())
		quota := peer.TrafficLimit != nil && peer.DownloadUsage+peer.UploadUsage >= *peer.TrafficLimit
		billing, admin := false, false
		if peer.ResellerID != nil {
			var reseller model.Reseller
			if err := tx.First(&reseller, *peer.ResellerID).Error; err != nil {
				return err
			}
			admin = reseller.AdminSuspended
			var account model.BillingAccount
			if err := tx.Where("reseller_id = ?", *peer.ResellerID).First(&account).Error; err != nil {
				return err
			}
			billing = account.BillingSuspended
		}
		desired := manual || expiry || quota || billing || admin
		changed := !peer.RestrictionsInitialized || peer.ManualDisabled != manual || peer.ExpiryBlocked != expiry || peer.QuotaBlocked != quota || peer.BillingBlocked != billing || peer.AdminBlocked != admin || peer.Disabled != desired
		if changed {
			peer.RestrictionRevision++
			peer.Disabled = desired
			peer.ManualDisabled = manual
			peer.ExpiryBlocked = expiry
			peer.QuotaBlocked = quota
			peer.BillingBlocked = billing
			peer.AdminBlocked = admin
			peer.RestrictionsInitialized = true
			peer.SyncNextAttemptAt = 0
			if err := tx.Model(&model.Peer{}).Where("id = ?", id).Updates(map[string]any{"restrictions_initialized": true, "manual_disabled": manual, "expiry_blocked": expiry, "quota_blocked": quota, "billing_blocked": billing, "admin_blocked": admin, "disabled": desired, "restriction_revision": peer.RestrictionRevision, "sync_next_attempt_at": 0}).Error; err != nil {
				return err
			}
		}
		if peer.SyncLeaseUntil > now || peer.SyncNextAttemptAt > now {
			return nil
		}
		if !changed && peer.SyncedDisabled != nil && *peer.SyncedDisabled == desired && peer.AppliedRevision == peer.RestrictionRevision && now-peer.LastSyncedAt < 300 {
			return nil
		}
		result := tx.Model(&model.Peer{}).Where("id = ? AND sync_lease_until <= ?", id, now).UpdateColumn("sync_lease_until", now+120)
		if result.Error != nil {
			return result.Error
		}
		claimed = result.RowsAffected == 1
		return nil
	})
	return peer, claimed, err
}

func (r *PeerReconciler) ReconcilePeer(id uint) error {
	peer, claimed, err := r.prepare(id)
	if err != nil || !claimed {
		return err
	}
	err = r.apply(peer)
	now := r.now().UTC().Unix()
	if err != nil {
		attempts := peer.SyncAttempts + 1
		delay := int64(5) * (int64(1) << min(attempts, int64(6)))
		delay = min(delay, int64(300))
		update := r.db.Model(&model.Peer{}).Where("id = ? AND restriction_revision = ?", id, peer.RestrictionRevision).Updates(map[string]any{"sync_lease_until": 0, "sync_attempts": attempts, "sync_next_attempt_at": now + delay, "sync_error": err.Error()})
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected == 0 {
			r.db.Model(&model.Peer{}).Where("id = ?", id).UpdateColumn("sync_lease_until", 0)
		}
		return err
	}
	result := r.db.Model(&model.Peer{}).Where("id = ? AND restriction_revision = ? AND server_id = ? AND peer_id = ?", id, peer.RestrictionRevision, peer.ServerID, peer.PeerID).Updates(map[string]any{"sync_lease_until": 0, "sync_attempts": 0, "sync_next_attempt_at": 0, "sync_error": "", "applied_revision": peer.RestrictionRevision, "synced_disabled": peer.Disabled, "last_synced_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return r.db.Model(&model.Peer{}).Where("id = ?", id).UpdateColumn("sync_lease_until", 0).Error
	}
	return nil
}

func (r *PeerReconciler) apply(peer model.Peer) error {
	adaptor, err := r.adaptor.ForServer(peer.ServerID)
	if err != nil {
		return err
	}
	ctx := context.Background()
	observed, err := adaptor.FetchWgPeer(ctx, peer.PeerID)
	if err != nil {
		return err
	}
	if observed.PublicKey != peer.PublicKey || observed.Interface != peer.Interface {
		return ErrRouterPeerIdentity
	}
	disabled := fmt.Sprintf("%t", peer.Disabled)
	// Block connectivity first when suspending. When restoring, enable only after
	// queue/scheduler ownership has been verified and other objects are ready.
	if peer.Disabled {
		if _, err := adaptor.UpdateWgPeer(ctx, peer.PeerID, mikrotik.WireGuardPeer{Disabled: disabled}); err != nil {
			return err
		}
	}
	if peer.QueueID != nil {
		queue, err := adaptor.FetchSimpleQueue(ctx, *peer.QueueID)
		if err != nil {
			return err
		}
		if queue.Target == nil || strings.TrimSuffix(*queue.Target, "/32") != strings.TrimSuffix(peer.AllowedAddress, "/32") {
			return ErrRouterPeerIdentity
		}
		if _, err := adaptor.UpdateSimpleQueue(ctx, *peer.QueueID, mikrotik.Queue{Disabled: disabled}); err != nil {
			return err
		}
	}
	if peer.SchedulerID != nil {
		scheduler, err := adaptor.FetchScheduler(ctx, *peer.SchedulerID)
		if err != nil {
			return err
		}
		if scheduler.OnEvent == nil || !strings.Contains(*scheduler.OnEvent, fmt.Sprintf(`.id="%s"`, peer.PeerID)) {
			return ErrRouterPeerIdentity
		}
		if _, err := adaptor.UpdateScheduler(ctx, *peer.SchedulerID, mikrotik.Scheduler{Disabled: fmt.Sprintf("%t", peer.ExpiryBlocked)}); err != nil {
			return err
		}
	}
	if !peer.Disabled {
		var current model.Peer
		if err := r.db.First(&current, peer.ID).Error; err != nil {
			return err
		}
		if current.RestrictionRevision != peer.RestrictionRevision || current.ServerID != peer.ServerID || current.PeerID != peer.PeerID {
			return errors.New("peer changed during reconciliation")
		}
		restricted, err := peerOtherRestriction(r.db, current, r.now())
		if err != nil {
			return err
		}
		if restricted || current.ManualDisabled {
			return ErrPeerRestricted
		}
		if _, err := adaptor.UpdateWgPeer(ctx, peer.PeerID, mikrotik.WireGuardPeer{Disabled: disabled}); err != nil {
			return err
		}
	}
	verified, err := adaptor.FetchWgPeer(ctx, peer.PeerID)
	if err != nil {
		return err
	}
	if verified.Disabled != disabled {
		return errors.New("router did not converge to the requested peer state")
	}
	return nil
}
