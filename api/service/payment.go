package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

const MaxReceiptBytes int64 = 5 << 20

var (
	ErrReceiptImage = errors.New("receipt must be a valid PNG or JPEG up to 5 MiB and 20 megapixels")
	ErrReceiptState = errors.New("receipt cannot transition from its current state")
)

type Payments struct {
	billing *Billing
	root    string
}

func NewPayments(billing *Billing, root string) *Payments {
	return &Payments{billing: billing, root: root}
}

type ReceiptInput struct {
	IdempotencyKey       string
	AmountRial           string
	PaidAt               string
	TransactionReference string
}

type ReceiptView struct {
	ID                   uint   `json:"id"`
	ResellerID           uint   `json:"reseller_id"`
	AmountRial           string `json:"amount_rial"`
	PaidAt               int64  `json:"paid_at"`
	TransactionReference string `json:"transaction_reference"`
	State                string `json:"state"`
	ReviewReason         string `json:"review_reason"`
	ReviewedAt           *int64 `json:"reviewed_at"`
	CreatedAt            uint64 `json:"created_at"`
}

func receiptView(r model.PaymentReceipt) ReceiptView {
	return ReceiptView{r.ID, r.ResellerID, strconv.FormatInt(r.AmountRial, 10), r.PaidAt, r.TransactionReference, r.State, r.ReviewReason, r.ReviewedAt, r.CreatedAt}
}

func parseReceiptInput(input ReceiptInput) (int64, int64, string, error) {
	amount, err := strconv.ParseInt(input.AmountRial, 10, 64)
	if err != nil || amount <= 0 || !validBillingKey(input.IdempotencyKey) {
		return 0, 0, "", ErrInvalidBillingQuantity
	}
	paidAt, err := time.Parse(time.RFC3339, input.PaidAt)
	if err != nil {
		paidAt, err = time.Parse("2006-01-02", input.PaidAt)
	}
	if err != nil || paidAt.Unix() <= 0 {
		return 0, 0, "", ErrInvalidBillingQuantity
	}
	reference := strings.TrimSpace(input.TransactionReference)
	if reference == "" || len(reference) > 128 {
		return 0, 0, "", ErrInvalidBillingQuantity
	}
	for _, char := range reference {
		if unicode.IsControl(char) {
			return 0, 0, "", ErrInvalidBillingQuantity
		}
	}
	return amount, paidAt.UTC().Unix(), reference, nil
}

func canonicalReceipt(reader io.Reader) ([]byte, string, string, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxReceiptBytes+1))
	if err != nil || int64(len(data)) > MaxReceiptBytes {
		return nil, "", "", ErrReceiptImage
	}
	mime := http.DetectContentType(data)
	if mime != "image/png" && mime != "image/jpeg" {
		return nil, "", "", ErrReceiptImage
	}
	conf, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || conf.Width < 1 || conf.Height < 1 || int64(conf.Width)*int64(conf.Height) > 20_000_000 {
		return nil, "", "", ErrReceiptImage
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", "", ErrReceiptImage
	}
	var output bytes.Buffer
	extension := ".png"
	// Re-encoding strips metadata, appended payloads and polyglot file contents.
	if format == "jpeg" {
		extension = ".jpg"
		err = jpeg.Encode(&output, decoded, &jpeg.Options{Quality: 95})
	} else {
		err = png.Encode(&output, decoded)
	}
	if err != nil || int64(output.Len()) > MaxReceiptBytes {
		return nil, "", "", ErrReceiptImage
	}
	return output.Bytes(), mime, extension, nil
}

func (p *Payments) Submit(actor Principal, resellerID uint, input ReceiptInput, reader io.Reader) (ReceiptView, error) {
	if !billingAccess(actor, resellerID) {
		return ReceiptView{}, ErrRoleForbidden
	}
	amount, paidAt, reference, err := parseReceiptInput(input)
	if err != nil {
		return ReceiptView{}, err
	}
	data, mime, extension, err := canonicalReceipt(reader)
	if err != nil {
		return ReceiptView{}, err
	}
	imageDigest := sha256.Sum256(data)
	imageHash := hex.EncodeToString(imageDigest[:])
	request, _ := json.Marshal(struct {
		Amount, PaidAt       int64
		Reference, ImageHash string
	}{amount, paidAt, reference, imageHash})
	requestDigest := sha256.Sum256(request)
	hash := hex.EncodeToString(requestDigest[:])
	if err := os.MkdirAll(p.root, 0700); err != nil {
		return ReceiptView{}, err
	}
	storageName := uuid.NewString() + extension
	filename := filepath.Join(p.root, storageName)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return ReceiptView{}, err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		os.Remove(filename)
		return ReceiptView{}, err
	}
	if err := file.Close(); err != nil {
		os.Remove(filename)
		return ReceiptView{}, err
	}
	created := false
	defer func() {
		if !created {
			os.Remove(filename)
		}
	}()
	var result model.PaymentReceipt
	err = p.billing.accountTransaction(resellerID, func(tx *gorm.DB, _ *model.BillingAccount) error {
		var existing model.PaymentReceipt
		lookup := tx.Where("reseller_id = ? AND idempotency_key = ?", resellerID, input.IdempotencyKey).First(&existing).Error
		if lookup == nil {
			if existing.RequestHash != hash {
				return ErrIdempotencyConflict
			}
			result = existing
			return nil
		}
		if !errors.Is(lookup, gorm.ErrRecordNotFound) {
			return lookup
		}
		result = model.PaymentReceipt{ResellerID: resellerID, IdempotencyKey: input.IdempotencyKey, RequestHash: hash, AmountRial: amount, PaidAt: paidAt, TransactionReference: reference, StorageName: storageName, ImageHash: imageHash, MIME: mime, State: model.ReceiptPending}
		return tx.Create(&result).Error
	})
	if err != nil {
		return ReceiptView{}, err
	}
	created = result.StorageName == storageName
	return receiptView(result), nil
}

func (p *Payments) History(actor Principal, resellerID uint) ([]ReceiptView, error) {
	if !billingAccess(actor, resellerID) {
		return nil, ErrRoleForbidden
	}
	var rows []model.PaymentReceipt
	if err := p.billing.db.Where("reseller_id = ?", resellerID).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]ReceiptView, 0, len(rows))
	for _, row := range rows {
		result = append(result, receiptView(row))
	}
	return result, nil
}

func (p *Payments) OpenReceipt(actor Principal, id uint) (*os.File, string, error) {
	var receipt model.PaymentReceipt
	query := p.billing.db
	if !billingAdmin(actor) {
		if actor.Role != model.RoleReseller || actor.AccountID == 0 || actor.ResellerID == nil {
			return nil, "", ErrRoleForbidden
		}
		query = query.Where("reseller_id = ?", *actor.ResellerID)
	}
	if err := query.First(&receipt, id).Error; err != nil {
		return nil, "", err
	}
	name := receipt.StorageName
	extension := filepath.Ext(name)
	if filepath.Base(name) != name || (extension != ".png" && extension != ".jpg") {
		return nil, "", ErrReceiptImage
	}
	if _, err := uuid.Parse(strings.TrimSuffix(name, extension)); err != nil {
		return nil, "", ErrReceiptImage
	}
	filename := filepath.Join(p.root, name)
	info, err := os.Lstat(filename)
	if err != nil {
		return nil, "", err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, "", ErrReceiptImage
	}
	file, err := os.Open(filename)
	return file, receipt.MIME, err
}

func (p *Payments) Approve(actor Principal, id uint) error {
	if !billingAdmin(actor) {
		return ErrRoleForbidden
	}
	file, _, err := p.OpenReceipt(actor, id)
	if err != nil {
		return err
	}
	file.Close()
	var initial model.PaymentReceipt
	if err := p.billing.db.First(&initial, id).Error; err != nil {
		return err
	}
	if _, err := p.billing.ReconcileAccount(initial.ResellerID); err != nil {
		return err
	}
	return p.billing.accountTransaction(initial.ResellerID, func(tx *gorm.DB, a *model.BillingAccount) error {
		var receipt model.PaymentReceipt
		if err := tx.First(&receipt, id).Error; err != nil {
			return err
		}
		if receipt.State == model.ReceiptApproved {
			var payment model.ApprovedPayment
			if err := tx.Where("receipt_id = ?", id).First(&payment).Error; err != nil {
				return err
			}
			if payment.AmountRial != receipt.AmountRial || payment.ResellerID != receipt.ResellerID || payment.TransactionReference != receipt.TransactionReference {
				return ErrReceiptState
			}
			return nil
		}
		if receipt.State != model.ReceiptPending {
			return ErrReceiptState
		}
		balance, err := checkedRialAdd(a.BalanceRial, receipt.AmountRial)
		if err != nil {
			return err
		}
		now := p.billing.now().UTC().Unix()
		entryID := uuid.NewString()
		paymentID := uuid.NewString()
		payment := model.ApprovedPayment{ID: paymentID, CreatedAt: now, ResellerID: a.ResellerID, ReceiptID: id, TransactionReference: receipt.TransactionReference, AmountRial: receipt.AmountRial, ActorID: actor.AccountID, LedgerEntryID: entryID}
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}
		entry := model.LedgerEntry{ID: entryID, CreatedAt: now, ResellerID: a.ResellerID, Reference: "payment:" + paymentID, ActorID: actor.AccountID, Kind: "payment", DeltaRial: receipt.AmountRial, BalanceAfterRial: balance, PeriodStart: a.PeriodStart, Reason: "approved card-to-card receipt " + strconv.FormatUint(uint64(id), 10)}
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}
		if err := allocateInvoiceCredit(tx, a.ResellerID, entryID, receipt.AmountRial, now); err != nil {
			return err
		}
		if err := tx.Model(a).Update("balance_rial", balance).Error; err != nil {
			return err
		}
		a.BalanceRial = balance
		state, err := p.billing.accountState(tx, a, now)
		if err != nil {
			return err
		}
		if err := queueBillingRestriction(tx, a.ResellerID, state.BillingSuspended); err != nil {
			return err
		}
		if err := tx.Model(a).Update("billing_suspended", state.BillingSuspended).Error; err != nil {
			return err
		}
		if err := tx.Model(&receipt).Updates(map[string]any{"state": model.ReceiptApproved, "reviewed_by": actor.AccountID, "reviewed_at": now, "review_reason": ""}).Error; err != nil {
			return err
		}
		return billingAudit(tx, a.ResellerID, actor.AccountID, nil, "receipt_approved", strconv.FormatUint(uint64(id), 10))
	})
}

func (p *Payments) Reject(actor Principal, id uint, reason string) error {
	if !billingAdmin(actor) {
		return ErrRoleForbidden
	}
	if strings.TrimSpace(reason) == "" || len(reason) > 1000 {
		return ErrInvalidBillingQuantity
	}
	var initial model.PaymentReceipt
	if err := p.billing.db.First(&initial, id).Error; err != nil {
		return err
	}
	return p.billing.accountTransaction(initial.ResellerID, func(tx *gorm.DB, _ *model.BillingAccount) error {
		var receipt model.PaymentReceipt
		if err := tx.First(&receipt, id).Error; err != nil {
			return err
		}
		if receipt.State == model.ReceiptRejected && receipt.ReviewReason == reason {
			return nil
		}
		if receipt.State != model.ReceiptPending {
			return ErrReceiptState
		}
		if err := tx.Model(&receipt).Updates(map[string]any{"state": model.ReceiptRejected, "reviewed_by": actor.AccountID, "reviewed_at": p.billing.now().Unix(), "review_reason": reason}).Error; err != nil {
			return err
		}
		return billingAudit(tx, initial.ResellerID, actor.AccountID, nil, "receipt_rejected", fmt.Sprintf("%d: %s", id, reason))
	})
}
