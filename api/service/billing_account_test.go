package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/internal/testdb"
	"gorm.io/gorm"
)

func billingFixture(t *testing.T, mode string) (*Billing, *gorm.DB, Principal, Principal, uint, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "billing.db")
	db := billingTestDB(t, path)
	if err := db.AutoMigrate(&model.Reseller{}, &model.Peer{}, &model.BillingAccount{}, &model.BillingCharge{}, &model.LedgerEntry{}, &model.BillingAuditEvent{}, &model.BillingInvoice{}, &model.InvoiceAllocation{}, &model.BillingSettings{}, &model.PaymentReceipt{}, &model.ApprovedPayment{}); err != nil {
		t.Fatal(err)
	}
	reseller := model.Reseller{Name: "reseller"}
	if err := db.Create(&reseller).Error; err != nil {
		t.Fatal(err)
	}
	admin := Principal{AccountID: 1, Role: model.RoleSuperadmin}
	tenant := Principal{AccountID: 2, Role: model.RoleReseller, ResellerID: &reseller.ID}
	b := NewBilling(db)
	if err := b.CreateAccount(admin, reseller.ID, mode); err != nil {
		t.Fatal(err)
	}
	return b, db, admin, tenant, reseller.ID, path
}

func billingTestDB(t *testing.T, path string) *gorm.DB {
	t.Helper()
	return testdb.Open(t, path)
}

func TestConcurrentPrepaidReservationsCannotOverspend(t *testing.T) {
	b, db, admin, tenant, id, path := billingFixture(t, model.BillingPrepaid)
	if err := b.Adjust(admin, id, "credit", 100000, "initial credit"); err != nil {
		t.Fatal(err)
	}
	independent := NewBilling(billingTestDB(t, path))
	results := make(chan error, 2)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i, instance := range []*Billing{b, independent} {
		wg.Add(1)
		go func(i int, instance *Billing) {
			defer wg.Done()
			<-start
			_, err := instance.ReserveCharge(tenant, id, fmt.Sprintf("request-%d", i), nil, BillingCreate, BillingGiB, 30)
			results <- err
		}(i, instance)
	}
	close(start)
	wg.Wait()
	close(results)
	success, rejected := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, ErrInsufficientCredit) {
			rejected++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || rejected != 1 {
		t.Fatalf("success=%d insufficient=%d", success, rejected)
	}
	account, err := b.Account(tenant, id)
	if err != nil {
		t.Fatal(err)
	}
	if account.BalanceRial != 100000 || account.ReservedRial != 100000 {
		t.Fatalf("invalid holds: %+v", account)
	}
	var charges []model.BillingCharge
	if err := db.Find(&charges).Error; err != nil {
		t.Fatal(err)
	}
	if len(charges) != 1 {
		t.Fatalf("charges=%d", len(charges))
	}
	if err := b.CommitCharge(tenant, id, charges[0].IdempotencyKey, nil); err != nil {
		t.Fatal(err)
	}
	account, err = b.Account(tenant, id)
	if err != nil {
		t.Fatal(err)
	}
	if account.BalanceRial != 0 || account.ReservedRial != 0 {
		t.Fatalf("invalid final balance: %+v", account)
	}
}

func TestChargeRetriesPreserveSnapshotAndPostOnlyOnce(t *testing.T) {
	b, db, admin, tenant, id, _ := billingFixture(t, model.BillingPrepaid)
	if err := b.Adjust(admin, id, "credit", 1000000, "initial credit"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		got, err := b.ReserveCharge(tenant, id, "same-key", nil, BillingCreate, BillingGiB, 30)
		if err != nil {
			t.Fatal(err)
		}
		if got.AmountRial != 100000 {
			t.Fatalf("rate changed historical intent: %+v", got)
		}
		if err := db.Model(&model.BillingAccount{}).Where("reseller_id = ?", id).Update("rate_rial", 200000).Error; err != nil {
			t.Fatal(err)
		}
	}
	if _, err := b.ReserveCharge(tenant, id, "same-key", nil, BillingCreate, 2*BillingGiB, 30); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("changed replay accepted: %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := b.CommitCharge(tenant, id, "same-key", nil); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	account, err := b.Account(tenant, id)
	if err != nil {
		t.Fatal(err)
	}
	if account.BalanceRial != 900000 || account.ReservedRial != 0 {
		t.Fatalf("duplicate debit: %+v", account)
	}
	var count int64
	if err := db.Model(&model.LedgerEntry{}).Where("kind = ?", "charge").Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	if err := b.ReleaseCharge(tenant, id, "same-key", "already committed"); !errors.Is(err, ErrChargeState) {
		t.Fatalf("released a finalized charge: %v", err)
	}
}

func TestFailedPeerCommitRollsBackLedgerAndHoldCanRecoverAfterRestart(t *testing.T) {
	b, db, admin, tenant, id, path := billingFixture(t, model.BillingPrepaid)
	if err := b.Adjust(admin, id, "credit", 100000, "initial credit"); err != nil {
		t.Fatal(err)
	}
	if _, err := b.ReserveCharge(tenant, id, "operation", nil, BillingCreate, BillingGiB, 30); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("peer update failed")
	if err := b.CommitCharge(tenant, id, "operation", func(tx *gorm.DB) error { return sentinel }); !errors.Is(err, sentinel) {
		t.Fatalf("expected rollback: %v", err)
	}
	var count int64
	if err := db.Model(&model.LedgerEntry{}).Where("kind = ?", "charge").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("unjustified charge: %d %v", count, err)
	}
	recovered := NewBilling(billingTestDB(t, path))
	for i := 0; i < 2; i++ {
		if err := recovered.ReleaseCharge(tenant, id, "operation", "verified router operation absent"); err != nil {
			t.Fatal(err)
		}
	}
	account, err := recovered.Account(tenant, id)
	if err != nil {
		t.Fatal(err)
	}
	if account.ReservedRial != 0 || account.BalanceRial != 100000 {
		t.Fatalf("hold not recovered: %+v", account)
	}
	if err := b.CommitCharge(tenant, id, "operation", nil); !errors.Is(err, ErrChargeState) {
		t.Fatalf("released intent committed: %v", err)
	}
}

func TestPostpaidModeChangePreservesLiabilitiesAndCredit(t *testing.T) {
	b, _, admin, tenant, id, _ := billingFixture(t, model.BillingPostpaid)
	if _, err := b.ReserveCharge(tenant, id, "charge", nil, BillingCreate, BillingGiB, 30); err != nil {
		t.Fatal(err)
	}
	if err := b.CommitCharge(tenant, id, "charge", nil); err != nil {
		t.Fatal(err)
	}
	if err := b.SetMode(admin, id, model.BillingPrepaid, "admin switch"); err != nil {
		t.Fatal(err)
	}
	account, err := b.Account(tenant, id)
	if err != nil {
		t.Fatal(err)
	}
	if account.BalanceRial != -100000 {
		t.Fatalf("liability lost: %+v", account)
	}
	if _, err := b.ReserveCharge(tenant, id, "blocked", nil, BillingCreate, BillingGiB, 30); !errors.Is(err, ErrInsufficientCredit) {
		t.Fatalf("debt authorized prepaid spend: %v", err)
	}
	if err := b.Adjust(admin, id, "settle", 150000, "manual settlement adjustment"); err != nil {
		t.Fatal(err)
	}
	if err := b.SetMode(admin, id, model.BillingPostpaid, "admin switch back"); err != nil {
		t.Fatal(err)
	}
	account, err = b.Account(tenant, id)
	if err != nil {
		t.Fatal(err)
	}
	if account.BalanceRial != 50000 {
		t.Fatalf("credit lost: %+v", account)
	}
}

func TestFinancialAuthorizationAndImmutableHistory(t *testing.T) {
	b, db, admin, tenant, id, _ := billingFixture(t, model.BillingPrepaid)
	if err := b.Adjust(tenant, id, "credit", 100000, "forged payment"); !errors.Is(err, ErrRoleForbidden) {
		t.Fatalf("reseller credited wallet: %v", err)
	}
	if err := b.SetMode(tenant, id, model.BillingPostpaid, "forged mode"); !errors.Is(err, ErrRoleForbidden) {
		t.Fatalf("reseller changed mode: %v", err)
	}
	other := uint(999)
	if _, err := b.Account(tenant, other); !errors.Is(err, ErrRoleForbidden) {
		t.Fatalf("cross-tenant account visible: %v", err)
	}
	if err := b.Adjust(admin, id, "credit", 100000, "admin credit"); err != nil {
		t.Fatal(err)
	}
	if err := b.Adjust(admin, id, "credit", 100000, "admin credit"); err != nil {
		t.Fatal(err)
	}
	if err := b.Adjust(admin, id, "credit", 200000, "admin credit"); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("adjustment replay changed amount: %v", err)
	}
	var entry model.LedgerEntry
	if err := db.First(&entry).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&entry).Update("delta_rial", 1).Error; !errors.Is(err, model.ErrImmutableAccounting) {
		t.Fatalf("edited history: %v", err)
	}
	if err := db.Delete(&entry).Error; !errors.Is(err, model.ErrImmutableAccounting) {
		t.Fatalf("deleted history: %v", err)
	}
	var event model.BillingAuditEvent
	if err := db.First(&event).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Delete(&event).Error; !errors.Is(err, model.ErrImmutableAccounting) {
		t.Fatalf("deleted audit: %v", err)
	}
}

func TestBillingCannotChargeOrCommitAnotherTenantsPeer(t *testing.T) {
	b, db, admin, tenant, id, _ := billingFixture(t, model.BillingPrepaid)
	if err := b.Adjust(admin, id, "credit", 1000000, "initial credit"); err != nil {
		t.Fatal(err)
	}
	other := uint(999)
	peer := model.Peer{UUID: "owned-peer", PeerID: "*1", Name: "owned", AllowedAddress: "10.0.0.2", ResellerID: &other}
	if err := db.Create(&peer).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := b.ReserveCharge(tenant, id, "forged-peer", &peer.ID, BillingExtendExpiry, BillingGiB, 30); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("foreign peer billed: %v", err)
	}
	if err := db.Model(&model.Peer{}).Where("id = ?", peer.ID).Update("reseller_id", id).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := b.ReserveCharge(tenant, id, "extension", &peer.ID, BillingExtendExpiry, BillingGiB, 30); err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&model.Peer{}).Where("id = ?", peer.ID).Update("reseller_id", other).Error; err != nil {
		t.Fatal(err)
	}
	if err := b.CommitCharge(tenant, id, "extension", nil); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("ownership changed but charge finalized: %v", err)
	}
	if err := b.ReleaseCharge(admin, id, "extension", "verified operation absent after ownership change"); err != nil {
		t.Fatal(err)
	}
	account, err := b.Account(tenant, id)
	if err != nil {
		t.Fatal(err)
	}
	if account.BalanceRial != 1000000 || account.ReservedRial != 0 {
		t.Fatalf("unjustified charge after transfer: %+v", account)
	}
}
