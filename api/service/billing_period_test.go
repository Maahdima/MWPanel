package service

import (
	"errors"
	"testing"
	"time"

	"github.com/maahdima/mwp/api/dataservice/model"
)

func periodFixture(t *testing.T, mode string) (*Billing, Principal, Principal, uint, *time.Time) {
	t.Helper()

	b, db, admin, tenant, id, _ := billingFixture(t, mode)

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	if err := db.Model(&model.BillingAccount{}).Where("reseller_id = ?", id).Updates(map[string]any{"activation_anchor": now.Unix(), "period_start": now.Unix(), "next_invoice_at": now.Add(30 * 24 * time.Hour).Unix()}).Error; err != nil {
		t.Fatal(err)
	}

	return b.WithClock(func() time.Time { return now }), admin, tenant, id, &now
}

func TestPostpaidBoundaryGracePartialSettlementAndAnchor(t *testing.T) {
	b, admin, tenant, id, clock := periodFixture(t, model.BillingPostpaid)

	anchor := *clock

	if _, err := b.ReserveCharge(tenant, id, "create", nil, BillingCreate, BillingGiB, 30); err != nil {
		t.Fatal(err)
	}

	if err := b.CommitCharge(tenant, id, "create", nil); err != nil {
		t.Fatal(err)
	}

	*clock = anchor.Add(30*24*time.Hour - time.Second)

	state, err := b.State(tenant, id)

	if err != nil || state.Status != "active" {
		t.Fatalf("premature due: %+v %v", state, err)
	}

	*clock = anchor.Add(30 * 24 * time.Hour)

	state, err = b.State(tenant, id)

	if err != nil || state.Status != "grace" || !state.BillableBlocked || state.OverdueRial != "100000" {
		t.Fatalf("boundary: %+v %v", state, err)
	}

	if _, err := b.ReserveCharge(tenant, id, "blocked", nil, BillingCreate, BillingGiB, 30); !errors.Is(err, ErrBillingOverdue) {
		t.Fatalf("overdue authorized a new operation: %v", err)
	}

	*clock = anchor.Add(35 * 24 * time.Hour)

	state, err = b.State(tenant, id)

	if err != nil || !state.BillingSuspended {
		t.Fatalf("grace expired: %+v %v", state, err)
	}

	if err := b.Adjust(admin, id, "partial", 40000, "verified partial credit"); err != nil {
		t.Fatal(err)
	}

	state, err = b.State(tenant, id)

	if err != nil || !state.BillingSuspended || state.OverdueRial != "60000" {
		t.Fatalf("partial credit lifted suspension: %+v %v", state, err)
	}

	if err := b.Adjust(admin, id, "settle", 60000, "remaining credit"); err != nil {
		t.Fatal(err)
	}

	if err := b.Adjust(admin, id, "settle", 60000, "remaining credit"); err != nil {
		t.Fatal(err)
	}

	state, err = b.State(tenant, id)

	if err != nil || state.BillingSuspended || state.BillableBlocked || state.BalanceRial != "0" {
		t.Fatalf("full settlement: %+v %v", state, err)
	}

	if state.ActivationAnchor != anchor.Unix() || state.NextInvoiceAt != anchor.Add(60*24*time.Hour).Unix() {
		t.Fatal("payment moved billing dates")
	}

	var invoice model.BillingInvoice

	if err := b.db.First(&invoice).Error; err != nil {
		t.Fatal(err)
	}

	if invoice.DueRial != 100000 || invoice.RemainingRial != 0 {
		t.Fatal("historical invoice was rewritten")
	}

	var count int64

	b.db.Model(&model.InvoiceAllocation{}).Count(&count)

	if count != 2 {
		t.Fatalf("duplicate allocation: %d", count)
	}
}

func TestEarlyCreditDoesNotMoveAnchorAndPrepaidZeroCreditStaysActive(t *testing.T) {
	b, admin, tenant, id, clock := periodFixture(t, model.BillingPostpaid)

	anchor := *clock

	if err := b.Adjust(admin, id, "early", 150000, "early verified credit"); err != nil {
		t.Fatal(err)
	}

	if _, err := b.ReserveCharge(tenant, id, "create", nil, BillingCreate, BillingGiB, 30); err != nil {
		t.Fatal(err)
	}

	if err := b.CommitCharge(tenant, id, "create", nil); err != nil {
		t.Fatal(err)
	}

	*clock = anchor.Add(30 * 24 * time.Hour)

	state, err := b.State(tenant, id)

	if err != nil || state.BillableBlocked || state.BalanceRial != "50000" || state.ActivationAnchor != anchor.Unix() {
		t.Fatalf("early credit: %+v %v", state, err)
	}

	var invoice model.BillingInvoice

	b.db.First(&invoice)

	if invoice.TotalRial != 100000 || invoice.DueRial != 0 {
		t.Fatalf("early credit invoice: %+v", invoice)
	}

	prepaid, _, prepaidTenant, prepaidID, prepaidClock := periodFixture(t, model.BillingPrepaid)

	*prepaidClock = prepaidClock.Add(60 * 24 * time.Hour)

	state, err = prepaid.State(prepaidTenant, prepaidID)

	if err != nil || state.BillableBlocked || state.BillingSuspended || state.AvailableRial != "0" {
		t.Fatalf("zero prepaid credit suspended existing service: %+v %v", state, err)
	}
}

func TestRecurringFeesAndDowntimeCatchupAreExactlyOnce(t *testing.T) {
	b, _, tenant, id, clock := periodFixture(t, model.BillingPostpaid)

	anchor := *clock

	if err := b.db.Model(&model.BillingAccount{}).Where("reseller_id = ?", id).Update("recurring_fee_rial", 10).Error; err != nil {
		t.Fatal(err)
	}

	*clock = anchor.Add(90 * 24 * time.Hour)

	for i := 0; i < 3; i++ {
		state, err := b.State(tenant, id)

		if err != nil || state.BalanceRial != "-30" {
			t.Fatalf("fee replay: %+v %v", state, err)
		}
	}

	var count int64

	b.db.Model(&model.LedgerEntry{}).Where("kind = ?", "recurring_fee").Count(&count)

	if count != 3 {
		t.Fatalf("fee count=%d", count)
	}

	b.db.Model(&model.BillingInvoice{}).Count(&count)

	if count != 3 {
		t.Fatalf("invoice count=%d", count)
	}

	long, _, longTenant, longID, longClock := periodFixture(t, model.BillingPostpaid)

	*longClock = longClock.Add(121 * 30 * 24 * time.Hour)

	state, err := long.State(longTenant, longID)

	if err != nil || !state.Reconciling {
		t.Fatalf("unbounded/unfinished catchup: %+v %v", state, err)
	}

	state, err = long.State(longTenant, longID)

	if err != nil || state.Reconciling {
		t.Fatalf("catchup did not resume: %+v %v", state, err)
	}

	long.db.Model(&model.BillingInvoice{}).Count(&count)

	if count != 121 {
		t.Fatalf("skipped overdue periods: %d", count)
	}
}

func TestGraceExtensionDoesNotEraseDebtOrAuthorizeNewCharges(t *testing.T) {
	b, admin, tenant, id, clock := periodFixture(t, model.BillingPostpaid)

	anchor := *clock

	if _, err := b.ReserveCharge(tenant, id, "create", nil, BillingCreate, BillingGiB, 30); err != nil {
		t.Fatal(err)
	}

	if err := b.CommitCharge(tenant, id, "create", nil); err != nil {
		t.Fatal(err)
	}

	*clock = anchor.Add(36 * 24 * time.Hour)

	if _, err := b.State(tenant, id); err != nil {
		t.Fatal(err)
	}

	if err := b.ExtendGrace(admin, id, anchor.Add(40*24*time.Hour).Unix(), "admin extension"); err != nil {
		t.Fatal(err)
	}

	state, err := b.State(tenant, id)

	if err != nil || state.BillingSuspended || !state.BillableBlocked || state.OverdueRial != "100000" {
		t.Fatalf("extension bypassed financial restriction: %+v %v", state, err)
	}

	if err := b.ExtendGrace(tenant, id, anchor.Add(50*24*time.Hour).Unix(), "forged"); !errors.Is(err, ErrRoleForbidden) {
		t.Fatal("reseller extended own grace")
	}
}
