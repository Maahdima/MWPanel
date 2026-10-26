package model

import "gorm.io/gorm"

const (
	ReceiptPending  = "pending"
	ReceiptApproved = "approved"
	ReceiptRejected = "rejected"
)

type BillingSettings struct {
	ID           uint   `gorm:"primaryKey"`
	RateRial     int64  `gorm:"not null"` // integer rials per GiB/30 days
	CardNumber   string `gorm:"type:varchar(16)"`
	Cardholder   string `gorm:"type:varchar(255)"`
	Instructions string `gorm:"type:text"`
	UpdatedBy    uint
	UpdatedAt    int64
}

type PaymentReceipt struct {
	Model
	ResellerID           uint   `gorm:"not null;index;uniqueIndex:uidx_receipt_request"`
	IdempotencyKey       string `gorm:"type:varchar(128);not null;uniqueIndex:uidx_receipt_request"`
	RequestHash          string `gorm:"type:varchar(64);not null"`
	AmountRial           int64  `gorm:"not null;check:receipt_amount_positive,amount_rial > 0"`
	PaidAt               int64  `gorm:"not null"`
	TransactionReference string `gorm:"type:varchar(128);not null;index"`
	StorageName          string `gorm:"type:varchar(64);not null"`
	ImageHash            string `gorm:"type:varchar(64);not null"`
	MIME                 string `gorm:"type:varchar(32);not null"`
	State                string `gorm:"type:varchar(16);not null;default:pending"`
	ReviewedBy           *uint
	ReviewedAt           *int64
	ReviewReason         string `gorm:"type:text"`
}

type ApprovedPayment struct {
	ID                   string `gorm:"type:varchar(36);primaryKey"`
	CreatedAt            int64  `gorm:"autoCreateTime"`
	ResellerID           uint   `gorm:"not null;index"`
	ReceiptID            uint   `gorm:"not null;uniqueIndex"`
	TransactionReference string `gorm:"type:varchar(128);not null;uniqueIndex"`
	AmountRial           int64  `gorm:"not null"`
	ActorID              uint   `gorm:"not null"`
	LedgerEntryID        string `gorm:"type:varchar(36);not null;uniqueIndex"`
}

func (*ApprovedPayment) BeforeUpdate(*gorm.DB) error { return ErrImmutableAccounting }

func (*ApprovedPayment) BeforeDelete(*gorm.DB) error { return ErrImmutableAccounting }
