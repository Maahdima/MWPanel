package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

var (
	ErrInsufficientCredit  = errors.New("insufficient available credit")
	ErrIdempotencyConflict = errors.New("idempotency key was used with a different request")
	ErrChargeState         = errors.New("charge cannot transition from its current state")
)

type Billing struct{ db *gorm.DB }

func NewBilling(db *gorm.DB) *Billing { return &Billing{db: db} }

func billingAdmin(actor Principal) bool {
	return actor.AccountID != 0 && actor.Role == model.RoleSuperadmin && actor.ResellerID == nil
}
func billingAccess(actor Principal, resellerID uint) bool {
	return billingAdmin(actor) || (actor.AccountID != 0 && actor.Role == model.RoleReseller && actor.ResellerID != nil && *actor.ResellerID == resellerID)
}

func (b *Billing) CreateAccount(actor Principal, resellerID uint, mode string) error {
	if !billingAdmin(actor) {
		return ErrRoleForbidden
	}
	if resellerID == 0 || (mode != model.BillingPrepaid && mode != model.BillingPostpaid) {
		return ErrInvalidBillingQuantity
	}
	return b.db.Transaction(func(tx *gorm.DB) error {
		var reseller model.Reseller
		if err := tx.First(&reseller, resellerID).Error; err != nil {
			return err
		}
		account := model.BillingAccount{ResellerID: resellerID, Mode: mode}
		if err := tx.Create(&account).Error; err != nil {
			return err
		}
		return billingAudit(tx, resellerID, actor.AccountID, nil, "account_created", mode)
	})
}

// accountTransaction takes a write lock before reading the balance. UPDATE locks
// the row on PostgreSQL and acquires SQLite's single-writer lock. All state and
// idempotency checks occur under that lock, with bounded lock-contention retries.
// No RouterOS or file operation is permitted inside this transaction callback.
func (b *Billing) accountTransaction(resellerID uint, fn func(*gorm.DB, *model.BillingAccount) error) error {
	for attempt := 0; attempt < 8; attempt++ {
		err := b.db.Transaction(func(tx *gorm.DB) error {
			result := tx.Model(&model.BillingAccount{}).Where("reseller_id = ?", resellerID).UpdateColumn("revision", gorm.Expr("revision + 1"))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
			var account model.BillingAccount
			if err := tx.Where("reseller_id = ?", resellerID).First(&account).Error; err != nil {
				return err
			}
			return fn(tx, &account)
		})
		if err == nil {
			return nil
		}
		message := strings.ToLower(err.Error())
		if !strings.Contains(message, "database is locked") && !strings.Contains(message, "sqlite_busy") && !strings.Contains(message, "sqlstate 40001") && !strings.Contains(message, "sqlstate 40p01") {
			return err
		}
		if attempt == 7 {
			return err
		}
		time.Sleep(time.Duration(attempt+1) * 5 * time.Millisecond)
	}
	return errors.New("billing transaction retry limit exceeded")
}

func checkedRialAdd(a, b int64) (int64, error) {
	if (b > 0 && a > math.MaxInt64-b) || (b < 0 && a < math.MinInt64-b) {
		return 0, ErrBillingOverflow
	}
	return a + b, nil
}

func billingAudit(tx *gorm.DB, resellerID, actorID uint, chargeID *uint, event, detail string) error {
	return tx.Create(&model.BillingAuditEvent{ID: uuid.NewString(), ResellerID: resellerID, ActorID: actorID, ChargeID: chargeID, Event: event, Detail: detail}).Error
}

func (b *Billing) Account(actor Principal, resellerID uint) (model.BillingAccount, error) {
	var account model.BillingAccount
	if !billingAccess(actor, resellerID) {
		return account, ErrRoleForbidden
	}
	err := b.db.Where("reseller_id = ?", resellerID).First(&account).Error
	return account, err
}

func validBillingKey(key string) bool {
	return strings.TrimSpace(key) == key && key != "" && len(key) <= 128
}

// ReserveCharge persists a durable intent and holds prepaid credit before any
// router mutation. A hold is never a finalized ledger charge. Retries return the
// original snapshot even after pricing changes; changed payloads are conflicts.
func (b *Billing) ReserveCharge(actor Principal, resellerID uint, key string, peerID *uint, operation BillingOperation, bytes, days int64) (model.BillingCharge, error) {
	var result model.BillingCharge
	if !billingAccess(actor, resellerID) {
		return result, ErrRoleForbidden
	}
	if !validBillingKey(key) {
		return result, ErrInvalidBillingQuantity
	}
	if (operation == BillingCreate && peerID != nil) || (operation != BillingCreate && peerID == nil) {
		return result, ErrInvalidBillingQuantity
	}
	request, _ := json.Marshal(struct {
		PeerID      *uint
		Operation   BillingOperation
		Bytes, Days int64
	}{peerID, operation, bytes, days})
	digest := sha256.Sum256(request)
	hash := hex.EncodeToString(digest[:])
	err := b.accountTransaction(resellerID, func(tx *gorm.DB, account *model.BillingAccount) error {
		var existing model.BillingCharge
		err := tx.Where("reseller_id = ? AND idempotency_key = ?", resellerID, key).First(&existing).Error
		if err == nil {
			if existing.RequestHash != hash {
				return ErrIdempotencyConflict
			}
			result = existing
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if peerID != nil {
			var peer model.Peer
			if err := tx.Where("reseller_id = ?", resellerID).First(&peer, *peerID).Error; err != nil {
				return err
			}
		}
		rate := DefaultPricePerGiB30DaysRial
		if account.RateRial != nil {
			rate = *account.RateRial
		}
		snapshot, err := PriceBillingOperation(operation, bytes, days, rate)
		if err != nil {
			return err
		}
		held, err := checkedRialAdd(account.ReservedRial, snapshot.AmountRial)
		if err != nil {
			return err
		}
		if account.Mode == model.BillingPrepaid && account.BalanceRial < held {
			return ErrInsufficientCredit
		}
		if account.Mode != model.BillingPrepaid && account.Mode != model.BillingPostpaid {
			return ErrInvalidBillingQuantity
		}
		if _, err := checkedRialAdd(account.BalanceRial, -held); err != nil {
			return err
		}
		encoded, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		result = model.BillingCharge{ResellerID: resellerID, IdempotencyKey: key, RequestHash: hash, PeerID: peerID, ActorID: actor.AccountID, State: model.ChargeReserved, AmountRial: snapshot.AmountRial, PricingSnapshot: string(encoded)}
		if err := tx.Create(&result).Error; err != nil {
			return err
		}
		if err := tx.Model(account).Update("reserved_rial", held).Error; err != nil {
			return err
		}
		return billingAudit(tx, resellerID, actor.AccountID, &result.ID, "charge_reserved", string(encoded))
	})
	return result, err
}

// CommitCharge is called only after the intended router state has been verified.
// apply writes the corresponding peer changes within the same transaction as the
// ledger posting. On failure the hold remains recoverable, with no finalized debit.
func (b *Billing) CommitCharge(actor Principal, resellerID uint, key string, apply func(*gorm.DB) error) error {
	if !billingAccess(actor, resellerID) {
		return ErrRoleForbidden
	}
	return b.accountTransaction(resellerID, func(tx *gorm.DB, account *model.BillingAccount) error {
		var charge model.BillingCharge
		if err := tx.Where("reseller_id = ? AND idempotency_key = ?", resellerID, key).First(&charge).Error; err != nil {
			return err
		}
		if charge.State == model.ChargeCommitted {
			return nil
		}
		if charge.State != model.ChargeReserved || account.ReservedRial < charge.AmountRial {
			return ErrChargeState
		}
		balance, err := checkedRialAdd(account.BalanceRial, -charge.AmountRial)
		if err != nil {
			return err
		}
		if account.Mode == model.BillingPrepaid && balance < 0 {
			return ErrInsufficientCredit
		}
		if charge.PeerID != nil {
			var peer model.Peer
			if err := tx.Where("reseller_id = ?", resellerID).First(&peer, *charge.PeerID).Error; err != nil {
				return err
			}
		}
		if apply != nil {
			if err := apply(tx); err != nil {
				return err
			}
		}
		if err := tx.Create(&model.LedgerEntry{ID: uuid.NewString(), ResellerID: resellerID, Reference: "charge:" + key, ActorID: actor.AccountID, Kind: "charge", DeltaRial: -charge.AmountRial, BalanceAfterRial: balance, Reason: charge.PricingSnapshot, ChargeID: &charge.ID}).Error; err != nil {
			return err
		}
		if err := tx.Model(account).Updates(map[string]any{"balance_rial": balance, "reserved_rial": account.ReservedRial - charge.AmountRial}).Error; err != nil {
			return err
		}
		if err := tx.Model(&charge).Update("state", model.ChargeCommitted).Error; err != nil {
			return err
		}
		return billingAudit(tx, resellerID, actor.AccountID, &charge.ID, "charge_committed", "")
	})
}

// ReleaseCharge is used only after verifying that the operation did not occur or
// was safely compensated. Ambiguous router timeouts must retain their intent.
func (b *Billing) ReleaseCharge(actor Principal, resellerID uint, key, reason string) error {
	if !billingAccess(actor, resellerID) {
		return ErrRoleForbidden
	}
	if strings.TrimSpace(reason) == "" || len(reason) > 1000 {
		return ErrInvalidBillingQuantity
	}
	return b.accountTransaction(resellerID, func(tx *gorm.DB, account *model.BillingAccount) error {
		var charge model.BillingCharge
		if err := tx.Where("reseller_id = ? AND idempotency_key = ?", resellerID, key).First(&charge).Error; err != nil {
			return err
		}
		if charge.State == model.ChargeReleased {
			return nil
		}
		if charge.State != model.ChargeReserved || account.ReservedRial < charge.AmountRial {
			return ErrChargeState
		}
		if err := tx.Model(account).Update("reserved_rial", account.ReservedRial-charge.AmountRial).Error; err != nil {
			return err
		}
		if err := tx.Model(&charge).Update("state", model.ChargeReleased).Error; err != nil {
			return err
		}
		return billingAudit(tx, resellerID, actor.AccountID, &charge.ID, "charge_released", reason)
	})
}

// Adjust posts an explicit signed superadmin adjustment with a mandatory reason.
// Receipt approval must call a separate atomic payment workflow, never this method.
func (b *Billing) Adjust(actor Principal, resellerID uint, key string, delta int64, reason string) error {
	if !billingAdmin(actor) {
		return ErrRoleForbidden
	}
	if !validBillingKey(key) || delta == 0 || strings.TrimSpace(reason) == "" || len(reason) > 1000 {
		return ErrInvalidBillingQuantity
	}
	return b.accountTransaction(resellerID, func(tx *gorm.DB, account *model.BillingAccount) error {
		var existing model.LedgerEntry
		err := tx.Where("reseller_id = ? AND reference = ?", resellerID, "adjustment:"+key).First(&existing).Error
		if err == nil {
			if existing.DeltaRial != delta || existing.Reason != reason {
				return ErrIdempotencyConflict
			}
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		balance, err := checkedRialAdd(account.BalanceRial, delta)
		if err != nil {
			return err
		}
		if delta < 0 && account.Mode == model.BillingPrepaid && balance < account.ReservedRial {
			return ErrInsufficientCredit
		}
		entry := model.LedgerEntry{ID: uuid.NewString(), ResellerID: resellerID, Reference: "adjustment:" + key, ActorID: actor.AccountID, Kind: "adjustment", DeltaRial: delta, BalanceAfterRial: balance, Reason: reason}
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}
		return tx.Model(account).Update("balance_rial", balance).Error
	})
}

func (b *Billing) SetMode(actor Principal, resellerID uint, mode, reason string) error {
	if !billingAdmin(actor) {
		return ErrRoleForbidden
	}
	if (mode != model.BillingPrepaid && mode != model.BillingPostpaid) || strings.TrimSpace(reason) == "" || len(reason) > 1000 {
		return ErrInvalidBillingQuantity
	}
	return b.accountTransaction(resellerID, func(tx *gorm.DB, account *model.BillingAccount) error {
		if account.Mode == mode {
			return nil
		}
		before := account.Mode
		if err := tx.Model(account).Update("mode", mode).Error; err != nil {
			return err
		}
		return billingAudit(tx, resellerID, actor.AccountID, nil, "mode_changed", before+" -> "+mode+": "+reason)
	})
}
