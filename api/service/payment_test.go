package service

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

func testReceiptPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var data bytes.Buffer
	if err := png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}
func TestReceiptApprovalIsAtomicIdempotentAndPreservesRestrictions(t *testing.T) {
	b, admin, tenant, id, clock := periodFixture(t, model.BillingPostpaid)
	anchor := *clock
	if _, err := b.ReserveCharge(tenant, id, "create", nil, BillingCreate, BillingGiB, 30); err != nil {
		t.Fatal(err)
	}
	if err := b.CommitCharge(tenant, id, "create", nil); err != nil {
		t.Fatal(err)
	}
	*clock = anchor.Add(35 * 24 * time.Hour)
	if _, err := b.State(tenant, id); err != nil {
		t.Fatal(err)
	}
	p := NewPayments(b, t.TempDir())
	imageData := testReceiptPNG(t)
	input := ReceiptInput{IdempotencyKey: "receipt-1", AmountRial: "40000", PaidAt: clock.Format(time.RFC3339), TransactionReference: "bank-123"}
	first, err := p.Submit(tenant, id, input, bytes.NewReader(imageData))
	if err != nil {
		t.Fatal(err)
	}
	repeat, err := p.Submit(tenant, id, input, bytes.NewReader(imageData))
	if err != nil || repeat.ID != first.ID {
		t.Fatal("receipt replay duplicated submission")
	}
	account, _ := b.Account(tenant, id)
	if account.BalanceRial != -100000 {
		t.Fatal("unapproved receipt credited balance")
	}
	if err := p.Approve(tenant, first.ID); !errors.Is(err, ErrRoleForbidden) {
		t.Fatal("reseller approved own receipt")
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := p.Approve(admin, first.ID); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	state, err := b.State(tenant, id)
	if err != nil || !state.BillingSuspended || state.OverdueRial != "60000" {
		t.Fatalf("partial approval lifted billing suspension: %+v %v", state, err)
	}
	if err := b.db.Model(&model.Reseller{}).Where("id = ?", id).Update("admin_suspended", true).Error; err != nil {
		t.Fatal(err)
	}
	input.IdempotencyKey = "receipt-2"
	input.AmountRial = "60000"
	input.TransactionReference = "bank-456"
	second, err := p.Submit(tenant, id, input, bytes.NewReader(imageData))
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Approve(admin, second.ID); err != nil {
		t.Fatal(err)
	}
	state, err = b.State(tenant, id)
	if err != nil || state.BillingSuspended || !state.AdministrativeSuspended || !state.BillableBlocked || state.BalanceRial != "0" {
		t.Fatalf("full payment cleared unrelated restriction: %+v %v", state, err)
	}
	var count int64
	b.db.Model(&model.ApprovedPayment{}).Count(&count)
	if count != 2 {
		t.Fatalf("payment count=%d", count)
	}
	b.db.Model(&model.InvoiceAllocation{}).Count(&count)
	if count != 2 {
		t.Fatalf("allocation count=%d", count)
	}
}

func TestReceiptDuplicateBankReferenceAndRejectedResubmission(t *testing.T) {
	b, admin, tenant, id, clock := periodFixture(t, model.BillingPrepaid)
	p := NewPayments(b, t.TempDir())
	data := testReceiptPNG(t)
	input := ReceiptInput{IdempotencyKey: "rejected", AmountRial: "100", PaidAt: clock.Format(time.RFC3339), TransactionReference: "same-bank-reference"}
	rejected, err := p.Submit(tenant, id, input, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Reject(admin, rejected.ID, "unreadable transaction details"); err != nil {
		t.Fatal(err)
	}
	input.IdempotencyKey = "resubmit"
	first, err := p.Submit(tenant, id, input, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Approve(admin, first.ID); err != nil {
		t.Fatal(err)
	}
	input.IdempotencyKey = "duplicate-bank"
	duplicate, err := p.Submit(tenant, id, input, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Approve(admin, duplicate.ID); err == nil {
		t.Fatal("bank reference credited twice")
	}
	account, _ := b.Account(tenant, id)
	if account.BalanceRial != 100 {
		t.Fatalf("duplicate approval was not rolled back: %d", account.BalanceRial)
	}
	var stored model.PaymentReceipt
	b.db.First(&stored, duplicate.ID)
	if stored.State != model.ReceiptPending {
		t.Fatal("failed approval changed receipt state")
	}
	input.AmountRial = "101"
	if _, err := p.Submit(tenant, id, input, bytes.NewReader(data)); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("receipt payload replay changed amount: %v", err)
	}
}

func TestReceiptImageAndOwnershipProtection(t *testing.T) {
	b, admin, tenant, id, clock := periodFixture(t, model.BillingPrepaid)
	root := t.TempDir()
	p := NewPayments(b, root)
	input := ReceiptInput{IdempotencyKey: "image", AmountRial: "10", PaidAt: clock.Format(time.RFC3339), TransactionReference: "secure-image"}
	for _, invalid := range [][]byte{[]byte("<html>not an image</html>"), make([]byte, MaxReceiptBytes+1), testReceiptPNG(t)[:20]} {
		if _, err := p.Submit(tenant, id, input, bytes.NewReader(invalid)); !errors.Is(err, ErrReceiptImage) {
			t.Fatalf("invalid image accepted: %v", err)
		}
	}
	data := append(testReceiptPNG(t), []byte("<script>appended payload</script>")...)
	receipt, err := p.Submit(tenant, id, input, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	otherID := id + 100
	other := Principal{AccountID: 9, Role: model.RoleReseller, ResellerID: &otherID}
	if _, _, err := p.OpenReceipt(other, receipt.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal("cross-tenant receipt visible")
	}
	if _, err := p.History(tenant, otherID); !errors.Is(err, ErrRoleForbidden) {
		t.Fatal("cross-tenant receipt history visible")
	}
	file, _, err := p.OpenReceipt(tenant, receipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()
	var stored model.PaymentReceipt
	b.db.First(&stored, receipt.ID)
	saved, err := os.ReadFile(filepath.Join(root, stored.StorageName))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(saved, []byte("<script>")) {
		t.Fatal("appended payload survived canonicalization")
	}
	if err := b.db.Model(&model.PaymentReceipt{}).Where("id = ?", receipt.ID).Update("storage_name", "../secret.png").Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := p.OpenReceipt(admin, receipt.ID); !errors.Is(err, ErrReceiptImage) {
		t.Fatal("path traversal accepted")
	}
}
