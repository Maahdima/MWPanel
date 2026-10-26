package model

import (
	"errors"

	"gorm.io/gorm"
)

const (
	BillingPrepaid  = "prepaid"
	BillingPostpaid = "postpaid"
	ChargeReserved  = "reserved"
	ChargeCommitted = "committed"
	ChargeReleased  = "released"
)

type BillingAccount struct {
	Model
	ResellerID uint     `gorm:"not null;uniqueIndex"`
	Reseller   Reseller `gorm:"constraint:OnDelete:RESTRICT"`
	Mode       string   `gorm:"type:varchar(16);not null;default:prepaid"`
	// Balance is signed: positive = credit, negative = liabilities. Holds are not
	// finalized charges and reduce only available prepaid credit.
	BalanceRial          int64 `gorm:"not null;default:0"`
	ReservedRial         int64 `gorm:"not null;default:0;check:billing_reservations_nonnegative,reserved_rial >= 0"`
	RateRial             *int64
	ActivationAnchor     int64 `gorm:"not null;default:0"`
	PeriodStart          int64 `gorm:"not null;default:0"`
	NextInvoiceAt        int64 `gorm:"not null;default:0"`
	IntervalDays         int64 `gorm:"not null;default:30"`
	GraceDays            int64 `gorm:"not null;default:5"`
	RecurringFeeRial     int64 `gorm:"not null;default:0"`
	BillingSuspended     bool  `gorm:"not null;default:false"`
	GraceOverrideUntil   int64 `gorm:"not null;default:0"`
	BillingOverrideUntil int64 `gorm:"not null;default:0"`
	Revision             int64 `gorm:"not null;default:0"`
}

type BillingCharge struct {
	Model
	ResellerID      uint   `gorm:"not null;uniqueIndex:idx_charge_request"`
	IdempotencyKey  string `gorm:"type:varchar(128);not null;uniqueIndex:idx_charge_request"`
	RequestHash     string `gorm:"type:varchar(64);not null"`
	PeerID          *uint
	ActorID         uint   `gorm:"not null"`
	State           string `gorm:"type:varchar(16);not null"`
	AmountRial      int64  `gorm:"not null;check:billing_charge_nonnegative,amount_rial >= 0"`
	PricingSnapshot string `gorm:"type:text;not null"`
}

// LedgerEntry and BillingAuditEvent are append-only in application code. Correct
// an accounting mistake through an explicit compensating entry, never an edit.
type LedgerEntry struct {
	ID               string `gorm:"type:varchar(36);primaryKey"`
	CreatedAt        int64  `gorm:"autoCreateTime"`
	ResellerID       uint   `gorm:"not null;index;uniqueIndex:idx_ledger_reference"`
	Reference        string `gorm:"type:varchar(160);not null;uniqueIndex:idx_ledger_reference"`
	ActorID          uint   `gorm:"not null"`
	Kind             string `gorm:"type:varchar(32);not null"`
	DeltaRial        int64  `gorm:"not null"`
	BalanceAfterRial int64  `gorm:"not null"`
	Reason           string `gorm:"type:text;not null"`
	PeriodStart      int64  `gorm:"not null;default:0;index"`
	ChargeID         *uint
}

type BillingAuditEvent struct {
	ID         string `gorm:"type:varchar(36);primaryKey"`
	CreatedAt  int64  `gorm:"autoCreateTime"`
	ResellerID uint   `gorm:"not null;index"`
	ActorID    uint   `gorm:"not null"`
	ChargeID   *uint
	Event      string `gorm:"type:varchar(32);not null"`
	Detail     string `gorm:"type:text;not null"`
}

var ErrImmutableAccounting = errors.New("accounting history is immutable; post a compensating entry")

func (*LedgerEntry) BeforeUpdate(*gorm.DB) error { return ErrImmutableAccounting }

func (*LedgerEntry) BeforeDelete(*gorm.DB) error { return ErrImmutableAccounting }

func (*BillingAuditEvent) BeforeUpdate(*gorm.DB) error { return ErrImmutableAccounting }

func (*BillingAuditEvent) BeforeDelete(*gorm.DB) error { return ErrImmutableAccounting }

type BillingInvoice struct {
	Model
	ResellerID    uint  `gorm:"not null;uniqueIndex:uidx_invoice_period"`
	PeriodStart   int64 `gorm:"not null;uniqueIndex:uidx_invoice_period"`
	PeriodEnd     int64 `gorm:"not null"`
	DueAt         int64 `gorm:"not null;index"`
	GraceUntil    int64 `gorm:"not null"`
	TotalRial     int64 `gorm:"not null"`
	DueRial       int64 `gorm:"not null"`
	RemainingRial int64 `gorm:"not null;check:invoice_remaining_nonnegative,remaining_rial >= 0"`
}

type InvoiceAllocation struct {
	ID            string `gorm:"type:varchar(36);primaryKey"`
	CreatedAt     int64  `gorm:"autoCreateTime"`
	InvoiceID     uint   `gorm:"not null;uniqueIndex:uidx_invoice_allocation"`
	LedgerEntryID string `gorm:"type:varchar(36);not null;uniqueIndex:uidx_invoice_allocation"`
	AmountRial    int64  `gorm:"not null;check:invoice_allocation_positive,amount_rial > 0"`
}

func (i *BillingInvoice) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("ResellerID", "PeriodStart", "PeriodEnd", "DueAt", "GraceUntil", "TotalRial", "DueRial") {
		return ErrImmutableAccounting
	}

	return nil
}

func (*BillingInvoice) BeforeDelete(*gorm.DB) error { return ErrImmutableAccounting }

func (*InvoiceAllocation) BeforeUpdate(*gorm.DB) error { return ErrImmutableAccounting }

func (*InvoiceAllocation) BeforeDelete(*gorm.DB) error { return ErrImmutableAccounting }
