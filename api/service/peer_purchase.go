package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"
	"gorm.io/gorm"
)

// PeerPurchaseRequest carries only purchased quantities. Ownership, usage, rate,
// router, current quota and expiration are always loaded on the server.
type PeerPurchaseRequest struct {
	Operation BillingOperation `json:"operation"`
	Bytes     string           `json:"bytes"`
	Days      string           `json:"days"`
}

type PeerPurchaseView struct {
	ID         string `json:"id"`
	PeerID     uint   `json:"peer_id"`
	State      string `json:"state"`
	AmountRial string `json:"amount_rial"`
	Attempts   int64  `json:"attempts"`
}

type PeerPurchases struct {
	runMu   sync.Mutex
	cursor  string
	billing *Billing
	peers   *WgPeer
}

func NewPeerPurchases(billing *Billing, peers *WgPeer) *PeerPurchases {
	return &PeerPurchases{billing: billing, peers: peers}
}

func (s *PeerPurchases) view(op model.PeerPurchase) (PeerPurchaseView, error) {
	var charge model.BillingCharge

	if op.ChargeID == nil {
		return PeerPurchaseView{op.ID, op.PeerID, op.State, "0", op.Attempts}, nil
	}

	err := s.billing.db.First(&charge, *op.ChargeID).Error

	return PeerPurchaseView{op.ID, op.PeerID, op.State, strconv.FormatInt(charge.AmountRial, 10), op.Attempts}, err
}

func parsePurchaseQuantity(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}

	if len(value) > 19 {
		return 0, ErrInvalidBillingQuantity
	}

	for _, c := range value {
		if c < '0' || c > '9' {
			return 0, ErrInvalidBillingQuantity
		}
	}

	result, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		return 0, ErrInvalidBillingQuantity
	}

	return result, nil
}

// Begin snapshots exact quantities and creates the financial hold and durable
// operation together. No external mutation occurs until that transaction commits.
func (s *PeerPurchases) Begin(actor Principal, resellerID, peerID uint, key string, req PeerPurchaseRequest) (PeerPurchaseView, error) {
	var empty PeerPurchaseView

	if !billingAccess(actor, resellerID) {
		return empty, ErrRoleForbidden
	}

	if !validBillingKey(key) {
		return empty, ErrInvalidBillingQuantity
	}

	payload, _ := json.Marshal(struct {
		PeerID  uint
		Request PeerPurchaseRequest
	}{peerID, req})

	digest := sha256.Sum256(payload)

	hash := hex.EncodeToString(digest[:])

	var existing model.PeerPurchase

	err := s.billing.db.Where("reseller_id = ? AND idempotency_key = ?", resellerID, key).First(&existing).Error

	if err == nil {
		if existing.RequestHash != hash {
			return empty, ErrIdempotencyConflict
		}

		return s.view(existing)
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return empty, err
	}

	var peer model.Peer

	if err = s.billing.db.Where("reseller_id = ?", resellerID).First(&peer, peerID).Error; err != nil {
		return empty, err
	}

	if peer.PurchaseID != nil || peer.MigrationID != nil {
		return empty, ErrPendingPeerOperation
	}

	if peer.TrafficLimit == nil || *peer.TrafficLimit <= 0 || peer.ExpireTime == nil {
		return empty, ErrInvalidBillingQuantity
	}

	bytes, err := parsePurchaseQuantity(req.Bytes)

	if err != nil {
		return empty, err
	}

	days, err := parsePurchaseQuantity(req.Days)

	if err != nil {
		return empty, err
	}

	desired := peer

	quota := *peer.TrafficLimit

	expiry := *peer.ExpireTime

	capturedTx, capturedRx := peer.LastTx, peer.LastRx

	reset := false

	pricingBytes, pricingDays := bytes, days

	switch req.Operation {
	case BillingAddTraffic:
		if bytes <= 0 || days != 0 {
			return empty, ErrInvalidBillingQuantity
		}

		quota, err = checkedRialAdd(quota, bytes)

		if err != nil {
			return empty, err
		}

	case BillingExtendExpiry:
		if bytes != 0 || days <= 0 || days > 36500 {
			return empty, ErrInvalidBillingQuantity
		}

		end, e := time.Parse("2006-01-02", expiry)

		if e != nil {
			return empty, ErrInvalidBillingQuantity
		}

		today := s.billing.now().UTC().Truncate(24 * time.Hour)

		if end.Before(today) {
			end = today
		}

		expiry = end.AddDate(0, 0, int(days)).Format("2006-01-02")

		pricingBytes = quota

	case BillingResetUsage, BillingFullReset:
		if bytes != 0 || days != 0 {
			return empty, ErrInvalidBillingQuantity
		}

		router, e := s.peers.mikrotikAdaptor.ForServer(peer.ServerID)

		if e != nil {
			return empty, e
		}

		observed, e := router.FetchWgPeer(context.Background(), peer.PeerID)

		if e != nil {
			return empty, e
		}

		if observed.PublicKey != peer.PublicKey || observed.Interface != peer.Interface {
			return empty, ErrRouterPeerIdentity
		}

		capturedTx, e = strconv.ParseInt(observed.TransferTx, 10, 64)

		if e != nil || capturedTx < 0 {
			return empty, ErrInvalidBillingQuantity
		}

		capturedRx, e = strconv.ParseInt(observed.TransferRx, 10, 64)

		if e != nil || capturedRx < 0 {
			return empty, ErrInvalidBillingQuantity
		}

		txDelta, _ := utils.PeerCounterDelta(peer.LastTx, capturedTx, 1<<32)

		rxDelta, _ := utils.PeerCounterDelta(peer.LastRx, capturedRx, 1<<32)

		consumed, e := checkedRialAdd(peer.DownloadUsage, peer.UploadUsage)

		if e != nil {
			return empty, e
		}

		consumed, e = checkedRialAdd(consumed, txDelta)

		if e != nil {
			return empty, e
		}

		consumed, e = checkedRialAdd(consumed, rxDelta)

		if e != nil {
			return empty, e
		}

		pricingBytes = consumed

		if req.Operation == BillingFullReset {
			if peer.ConfiguredDurationDays <= 0 || peer.ConfiguredDurationDays > 36500 {
				return empty, ErrInvalidBillingQuantity
			}

			pricingBytes = quota

			pricingDays = peer.ConfiguredDurationDays

			expiry = s.billing.now().UTC().AddDate(0, 0, int(pricingDays)).Format("2006-01-02")
		}

		reset = true

	default:
		return empty, ErrInvalidBillingQuantity
	}

	desired.TrafficLimit = &quota

	desired.ExpireTime = &expiry

	if reset {
		desired.DownloadUsage = 0

		desired.UploadUsage = 0
	}

	desired.Disabled = desired.ManualDisabled || utils.IsPeerExpired(&expiry, s.billing.now()) || desired.DownloadUsage+desired.UploadUsage >= quota || desired.BillingBlocked || desired.AdminBlocked

	op := model.PeerPurchase{ID: uuid.NewString(), ResellerID: resellerID, IdempotencyKey: key, RequestHash: hash, PeerID: peerID, ActorID: actor.AccountID, Kind: string(req.Operation), State: "pending", QuotaBytes: quota, ExpireTime: expiry, ResetUsage: reset, CapturedTx: capturedTx, CapturedRx: capturedRx}

	_, err = s.billing.reserveCharge(actor, resellerID, key, &peerID, req.Operation, pricingBytes, pricingDays, func(tx *gorm.DB, charge model.BillingCharge) error {
		if e := migrationUnlocked(tx, resellerID); e != nil {
			return e
		}

		// CAS rejects a concurrent sample or mutation: retry with a new snapshot;
		// the transaction rolls back the hold as well.
		lock := tx.Model(&model.Peer{}).Where("id = ? AND reseller_id = ? AND purchase_id IS NULL AND restriction_revision = ? AND updated_at = ? AND last_tx = ? AND last_rx = ? AND download_usage = ? AND upload_usage = ? AND sync_lease_until <= ?", peer.ID, resellerID, peer.RestrictionRevision, peer.UpdatedAt, peer.LastTx, peer.LastRx, peer.DownloadUsage, peer.UploadUsage, s.billing.now().Unix()).UpdateColumn("purchase_id", op.ID)

		if lock.Error != nil {
			return lock.Error
		}

		if lock.RowsAffected != 1 {
			return ErrPendingPeerOperation
		}

		if e := ValidatePeerResource(tx, resellerID, desired, peerID); e != nil {
			return e
		}

		op.ChargeID = &charge.ID

		return tx.Create(&op).Error
	})

	if err != nil {
		return empty, err
	}

	// Another identical caller may have won the same key.
	op = model.PeerPurchase{}

	if err = s.billing.db.Where("reseller_id = ? AND idempotency_key = ?", resellerID, key).First(&op).Error; err != nil {
		return empty, err
	}

	if op.RequestHash != hash {
		return empty, ErrIdempotencyConflict
	}

	return s.view(op)
}

// Run processes a bounded batch and durable leases. On timeout or database
// failure the hold remains pending; a restart verifies and finishes the same
// operation, never a second charge or an unverified refund.
func (s *PeerPurchases) Run() {
	if !s.runMu.TryLock() {
		return
	}

	defer s.runMu.Unlock()

	var ops []model.PeerPurchase

	now := s.billing.now().UTC().Unix()

	if s.billing.db.Where("state = ? AND next_attempt_at <= ? AND lease_until <= ? AND id > ?", "pending", now, now, s.cursor).Order("id ASC").Limit(50).Find(&ops).Error != nil {
		return
	}

	if len(ops) == 0 {
		s.cursor = ""
	}

	for _, op := range ops {
		s.cursor = op.ID

		_ = s.Reconcile(op.ID)
	}
}

func (s *PeerPurchases) Reconcile(id string) error {
	now := s.billing.now().UTC().Unix()

	claim := s.billing.db.Model(&model.PeerPurchase{}).Where("id = ? AND state = ? AND lease_until <= ? AND next_attempt_at <= ?", id, "pending", now, now).UpdateColumn("lease_until", now+120)

	if claim.Error != nil {
		return claim.Error
	}

	if claim.RowsAffected == 0 {
		return nil
	}

	var op model.PeerPurchase

	if err := s.billing.db.First(&op, "id = ?", id).Error; err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)

	defer cancel()

	worker := &PeerPurchases{billing: s.billing.withContext(ctx), peers: s.peers}

	err := worker.apply(ctx, op)

	if err != nil {
		delay := min(int64(300), int64(1)<<min(op.Attempts+1, int64(8)))

		_ = s.billing.db.Model(&model.PeerPurchase{}).Where("id = ? AND state = ?", id, "pending").Updates(map[string]any{"lease_until": 0, "attempts": gorm.Expr("attempts + 1"), "next_attempt_at": now + delay, "last_error": "router or storage verification pending"}).Error
	}

	return err
}

func (s *PeerPurchases) apply(ctx context.Context, op model.PeerPurchase) error {
	var peer model.Peer

	if err := s.billing.db.Where("reseller_id = ? AND purchase_id = ?", op.ResellerID, op.ID).First(&peer, op.PeerID).Error; err != nil {
		return err
	}

	if op.ChargeID == nil {
		return s.applyFree(ctx, op, peer)
	}

	if op.Kind == string(BillingCreate) {
		return s.applyCreate(ctx, op, peer)
	}

	router, err := s.peers.mikrotikAdaptor.ForServer(peer.ServerID)

	if err != nil {
		return err
	}

	observed, err := router.FetchWgPeer(ctx, peer.PeerID)

	if err != nil {
		return err
	}

	if observed.PublicKey != peer.PublicKey || observed.Interface != peer.Interface {
		return ErrRouterPeerIdentity
	}

	if op.ExpireTime != utils.DerefString(peer.ExpireTime) {
		if peer.SchedulerID == nil {
			return ErrRouterPeerIdentity
		}

		scheduler, err := router.FetchScheduler(ctx, *peer.SchedulerID)

		if err != nil {
			return err
		}

		if scheduler.OnEvent == nil || *scheduler.OnEvent != schedulerOnEvent(peer.PeerID) {
			return ErrRouterPeerIdentity
		}

		bound, err := s.peers.ForServer(peer.ServerID)

		if err != nil {
			return err
		}

		if err = bound.scheduler.updateSchedulerContext(ctx, peer.SchedulerID, peer.PeerID, &op.ExpireTime); err != nil {
			return err
		}

		verify, err := router.FetchScheduler(ctx, *peer.SchedulerID)

		if err != nil {
			return err
		}

		expected, err := toMikrotikDate(&op.ExpireTime)

		if err != nil {
			return err
		}

		if verify.StartDate == nil || expected == nil || *verify.StartDate != *expected || verify.OnEvent == nil || *verify.OnEvent != schedulerOnEvent(peer.PeerID) {
			return ErrRouterPeerIdentity
		}
	}

	actor := Principal{AccountID: op.ActorID, Role: model.RoleReseller, ResellerID: &op.ResellerID}

	return s.billing.CommitCharge(actor, op.ResellerID, op.IdempotencyKey, func(tx *gorm.DB) error {
		updates := map[string]any{"traffic_limit": op.QuotaBytes, "expire_time": op.ExpireTime, "purchase_id": nil, "restriction_revision": gorm.Expr("restriction_revision + 1"), "sync_next_attempt_at": 0, "expire_first_notify": false, "expire_second_notify": false, "expire_third_notify": false}

		if op.ResetUsage {
			updates["download_usage"] = 0

			updates["upload_usage"] = 0

			updates["last_tx"] = op.CapturedTx

			updates["last_rx"] = op.CapturedRx

			updates["traffic_first_notify"] = false

			updates["traffic_second_notify"] = false

			updates["traffic_third_notify"] = false
		}

		result := tx.Model(&model.Peer{}).Where("id = ? AND reseller_id = ? AND purchase_id = ?", op.PeerID, op.ResellerID, op.ID).Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected != 1 {
			return ErrPendingPeerOperation
		}

		return tx.Model(&model.PeerPurchase{}).Where("id = ? AND state = ?", op.ID, "pending").Updates(map[string]any{"state": "applied", "lease_until": 0, "last_error": ""}).Error
	})
}

func (s *Resellers) PurchasePeer(actor Principal, resellerID, peerID uint, key string, req PeerPurchaseRequest) (PeerPurchaseView, error) {
	return NewPeerPurchases(NewBilling(s.db), s.peers).Begin(actor, resellerID, peerID, key, req)
}

func (s *Resellers) PeerPurchaseHistory(actor Principal, resellerID, peerID uint) ([]PeerPurchaseView, error) {
	if !billingAccess(actor, resellerID) {
		return nil, ErrRoleForbidden
	}

	var peer model.Peer

	if err := s.db.Where("reseller_id = ?", resellerID).First(&peer, peerID).Error; err != nil {
		return nil, err
	}

	var rows []model.PeerPurchase

	if err := s.db.Where("reseller_id = ? AND peer_id = ?", resellerID, peerID).Order("created_at DESC").Limit(50).Find(&rows).Error; err != nil {
		return nil, err
	}

	worker := NewPeerPurchases(NewBilling(s.db), s.peers)

	result := make([]PeerPurchaseView, 0, len(rows))

	for _, row := range rows {
		v, err := worker.view(row)

		if err != nil {
			return nil, err
		}

		result = append(result, v)
	}

	return result, nil
}
