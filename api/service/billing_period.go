package service

import (
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

var (
	ErrBillingOverdue        = errors.New("settle overdue obligations before starting a new billable operation")
	ErrBillingCatchupPending = errors.New("billing period reconciliation is still catching up")
)

func (b *Billing) WithClock(clock func() time.Time) *Billing {
	clone := *b
	clone.now = clock
	return &clone
}

type BillingState struct {
	AdministrativeSuspended bool   `json:"administrative_suspended"`
	Status                  string `json:"status"`
	BalanceRial             string `json:"balance_rial"`
	AvailableRial           string `json:"available_rial"`
	OverdueRial             string `json:"overdue_rial"`
	BillableBlocked         bool   `json:"billable_blocked"`
	BillingSuspended        bool   `json:"billing_suspended"`
	DueAt                   int64  `json:"due_at"`
	GraceUntil              int64  `json:"grace_until"`
	NextInvoiceAt           int64  `json:"next_invoice_at"`
	ActivationAnchor        int64  `json:"activation_anchor"`
	Reconciling             bool   `json:"reconciling"`
}

func billingDays(days int64) (int64, error) {
	if days < 0 || days > math.MaxInt64/86400 {
		return 0, ErrInvalidBillingQuantity
	}
	return days * 86400, nil
}

func (b *Billing) initializePeriod(tx *gorm.DB, a *model.BillingAccount) error {
	if a.IntervalDays <= 0 {
		a.IntervalDays = 30
	}
	if a.GraceDays < 0 || a.RecurringFeeRial < 0 {
		return ErrInvalidBillingQuantity
	}
	if a.ActivationAnchor == 0 {
		a.ActivationAnchor = int64(a.CreatedAt)
		if a.ActivationAnchor == 0 {
			a.ActivationAnchor = b.now().UTC().Unix()
		}
	}
	if a.PeriodStart == 0 {
		a.PeriodStart = a.ActivationAnchor
	}
	interval, err := billingDays(a.IntervalDays)
	if err != nil {
		return err
	}
	if a.NextInvoiceAt == 0 {
		a.NextInvoiceAt, err = checkedRialAdd(a.PeriodStart, interval)
		if err != nil {
			return err
		}
	}
	return tx.Model(a).Updates(map[string]any{"activation_anchor": a.ActivationAnchor, "period_start": a.PeriodStart, "next_invoice_at": a.NextInvoiceAt, "interval_days": a.IntervalDays}).Error
}

// Periods are fixed day durations, not calendar months. An early payment cannot
// change their anchor. Long downtime is processed in bounded, resumable batches.
func (b *Billing) advancePeriods(tx *gorm.DB, a *model.BillingAccount, now int64) error {
	if err := b.initializePeriod(tx, a); err != nil {
		return err
	}
	interval, err := billingDays(a.IntervalDays)
	if err != nil {
		return err
	}
	grace, err := billingDays(a.GraceDays)
	if err != nil {
		return err
	}
	for processed := 0; processed < 120 && now >= a.NextInvoiceAt; processed++ {
		end := a.NextInvoiceAt
		if a.RecurringFeeRial > 0 {
			balance, err := checkedRialAdd(a.BalanceRial, -a.RecurringFeeRial)
			if err != nil {
				return err
			}
			entry := model.LedgerEntry{ID: uuid.NewString(), CreatedAt: end, ResellerID: a.ResellerID, Reference: "period-fee:" + strconv.FormatInt(a.PeriodStart, 10), Kind: "recurring_fee", DeltaRial: -a.RecurringFeeRial, BalanceAfterRial: balance, PeriodStart: a.PeriodStart, Reason: "configured recurring period fee"}
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
			a.BalanceRial = balance
		}
		var entries []model.LedgerEntry
		if err := tx.Where("reseller_id = ? AND period_start = ?", a.ResellerID, a.PeriodStart).Find(&entries).Error; err != nil {
			return err
		}
		total := int64(0)
		for _, entry := range entries {
			if entry.DeltaRial < 0 {
				if entry.DeltaRial == math.MinInt64 {
					return ErrBillingOverflow
				}
				total, err = checkedRialAdd(total, -entry.DeltaRial)
				if err != nil {
					return err
				}
			}
		}
		var older []model.BillingInvoice
		if err := tx.Where("reseller_id = ? AND remaining_rial > 0", a.ResellerID).Find(&older).Error; err != nil {
			return err
		}
		outstanding := int64(0)
		for _, invoice := range older {
			outstanding, err = checkedRialAdd(outstanding, invoice.RemainingRial)
			if err != nil {
				return err
			}
		}
		due := int64(0)
		if a.BalanceRial < 0 {
			due = max(int64(0), -a.BalanceRial-outstanding)
		}
		graceUntil, err := checkedRialAdd(end, grace)
		if err != nil {
			return err
		}
		invoice := model.BillingInvoice{ResellerID: a.ResellerID, PeriodStart: a.PeriodStart, PeriodEnd: end, DueAt: end, GraceUntil: graceUntil, TotalRial: total, DueRial: due, RemainingRial: due}
		if err := tx.Create(&invoice).Error; err != nil {
			return err
		}
		a.PeriodStart = end
		a.NextInvoiceAt, err = checkedRialAdd(end, interval)
		if err != nil {
			return err
		}
		if err := tx.Model(a).Updates(map[string]any{"period_start": a.PeriodStart, "next_invoice_at": a.NextInvoiceAt, "balance_rial": a.BalanceRial}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (b *Billing) accountState(tx *gorm.DB, a *model.BillingAccount, now int64) (BillingState, error) {
	state := BillingState{Status: "active", BalanceRial: strconv.FormatInt(a.BalanceRial, 10), AvailableRial: "0", OverdueRial: "0", ActivationAnchor: a.ActivationAnchor, NextInvoiceAt: a.NextInvoiceAt, Reconciling: now >= a.NextInvoiceAt}
	if a.BalanceRial > a.ReservedRial {
		state.AvailableRial = strconv.FormatInt(a.BalanceRial-a.ReservedRial, 10)
	}
	var invoices []model.BillingInvoice
	if err := tx.Where("reseller_id = ? AND remaining_rial > 0 AND due_at <= ?", a.ResellerID, now).Order("due_at ASC").Find(&invoices).Error; err != nil {
		return state, err
	}
	overdue := int64(0)
	for _, invoice := range invoices {
		var err error
		overdue, err = checkedRialAdd(overdue, invoice.RemainingRial)
		if err != nil {
			return state, err
		}
		until := max(invoice.GraceUntil, a.GraceOverrideUntil)
		if state.DueAt == 0 || invoice.DueAt < state.DueAt {
			state.DueAt = invoice.DueAt
			state.GraceUntil = until
		}
		if now >= until {
			state.BillingSuspended = true
		}
	}
	state.OverdueRial = strconv.FormatInt(overdue, 10)
	if overdue > 0 {
		state.Status = "grace"
		state.BillableBlocked = true
		if state.BillingSuspended {
			state.Status = "suspended"
		}
		if a.BillingOverrideUntil > now {
			state.Status = "admin_override"
			state.BillableBlocked = false
			state.BillingSuspended = false
		}
	}
	var reseller model.Reseller
	if err := tx.Select("admin_suspended").First(&reseller, a.ResellerID).Error; err != nil {
		return state, err
	}
	state.AdministrativeSuspended = reseller.AdminSuspended
	if reseller.AdminSuspended {
		state.Status = "admin_suspended"
		state.BillableBlocked = true
	}
	if state.Reconciling {
		state.BillableBlocked = true
	}
	return state, nil
}

func (b *Billing) ReconcileAccount(resellerID uint) (BillingState, error) {
	var state BillingState
	err := b.accountTransaction(resellerID, func(tx *gorm.DB, a *model.BillingAccount) error {
		if err := b.advancePeriods(tx, a, b.now().UTC().Unix()); err != nil {
			return err
		}
		calculated, err := b.accountState(tx, a, b.now().UTC().Unix())
		if err != nil {
			return err
		}
		state = calculated
		if err := queueBillingRestriction(tx, a.ResellerID, state.BillingSuspended); err != nil {
			return err
		}
		if a.BillingSuspended != state.BillingSuspended {
			if err := tx.Model(a).Update("billing_suspended", state.BillingSuspended).Error; err != nil {
				return err
			}
			return billingAudit(tx, a.ResellerID, 0, nil, "billing_state_changed", state.Status)
		}
		return nil
	})
	return state, err
}

func allocateInvoiceCredit(tx *gorm.DB, resellerID uint, entryID string, amount, now int64) error {
	var invoices []model.BillingInvoice
	if err := tx.Where("reseller_id = ? AND remaining_rial > 0", resellerID).Order("due_at ASC").Order("id ASC").Find(&invoices).Error; err != nil {
		return err
	}
	for _, invoice := range invoices {
		if amount <= 0 {
			break
		}
		paid := min(amount, invoice.RemainingRial)
		allocation := model.InvoiceAllocation{ID: uuid.NewString(), CreatedAt: now, InvoiceID: invoice.ID, LedgerEntryID: entryID, AmountRial: paid}
		if err := tx.Create(&allocation).Error; err != nil {
			return err
		}
		if err := tx.Model(&invoice).Update("remaining_rial", invoice.RemainingRial-paid).Error; err != nil {
			return err
		}
		amount -= paid
	}
	return nil
}

func (b *Billing) State(actor Principal, resellerID uint) (BillingState, error) {
	if !billingAccess(actor, resellerID) {
		return BillingState{}, ErrRoleForbidden
	}
	return b.ReconcileAccount(resellerID)
}

func (b *Billing) ExtendGrace(actor Principal, resellerID uint, until int64, reason string) error {
	if !billingAdmin(actor) {
		return ErrRoleForbidden
	}
	if until <= b.now().Unix() || reason == "" {
		return ErrInvalidBillingQuantity
	}
	if _, err := b.ReconcileAccount(resellerID); err != nil {
		return err
	}
	return b.accountTransaction(resellerID, func(tx *gorm.DB, a *model.BillingAccount) error {
		if until <= a.GraceOverrideUntil {
			return ErrInvalidBillingQuantity
		}
		if err := tx.Model(a).Update("grace_override_until", until).Error; err != nil {
			return err
		}
		return billingAudit(tx, resellerID, actor.AccountID, nil, "grace_extended", reason+": "+strconv.FormatInt(until, 10))
	})
}

func (b *Billing) OverrideBilling(actor Principal, id uint, until int64, reason string) error {
	if !billingAdmin(actor) {
		return ErrRoleForbidden
	}
	if until <= b.now().Unix() || reason == "" || len(reason) > 1000 {
		return ErrInvalidBillingQuantity
	}
	if _, err := b.ReconcileAccount(id); err != nil {
		return err
	}
	return b.accountTransaction(id, func(tx *gorm.DB, a *model.BillingAccount) error {
		if err := tx.Model(a).Update("billing_override_until", until).Error; err != nil {
			return err
		}
		a.BillingOverrideUntil = until
		state, err := b.accountState(tx, a, b.now().Unix())
		if err != nil {
			return err
		}
		if err := tx.Model(a).Update("billing_suspended", state.BillingSuspended).Error; err != nil {
			return err
		}
		if err := queueBillingRestriction(tx, id, state.BillingSuspended); err != nil {
			return err
		}
		return billingAudit(tx, id, actor.AccountID, nil, "billing_override", reason+": "+strconv.FormatInt(until, 10))
	})
}
