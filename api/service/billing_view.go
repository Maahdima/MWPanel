package service

import (
	"strconv"

	"github.com/maahdima/mwp/api/dataservice/model"
)

type InvoiceView struct {
	ID            uint   `json:"id"`
	PeriodStart   int64  `json:"period_start"`
	PeriodEnd     int64  `json:"period_end"`
	DueAt         int64  `json:"due_at"`
	GraceUntil    int64  `json:"grace_until"`
	TotalRial     string `json:"total_rial"`
	DueRial       string `json:"due_rial"`
	RemainingRial string `json:"remaining_rial"`
}

type LedgerView struct {
	ID               string `json:"id"`
	CreatedAt        int64  `json:"created_at"`
	Kind             string `json:"kind"`
	DeltaRial        string `json:"delta_rial"`
	BalanceAfterRial string `json:"balance_after_rial"`
	Reason           string `json:"reason"`
}

type BillingBook struct {
	ResellerID       uint          `json:"reseller_id"`
	Mode             string        `json:"billing_mode"`
	State            BillingState  `json:"state"`
	RateRial         string        `json:"rate_rial"`
	ReservedRial     string        `json:"reserved_rial"`
	IntervalDays     int64         `json:"interval_days"`
	GraceDays        int64         `json:"grace_days"`
	RecurringFeeRial string        `json:"recurring_fee_rial"`
	Invoices         []InvoiceView `json:"recent_invoices"`
	Ledger           []LedgerView  `json:"recent_ledger"`
}

func (b *Billing) Book(actor Principal, resellerID uint) (BillingBook, error) {
	var result BillingBook

	state, err := b.State(actor, resellerID)

	if err != nil {
		return result, err
	}

	account, err := b.Account(actor, resellerID)

	if err != nil {
		return result, err
	}

	rate, err := billingRate(b.db, &account)

	if err != nil {
		return result, err
	}

	result = BillingBook{ResellerID: resellerID, Mode: account.Mode, State: state, RateRial: strconv.FormatInt(rate, 10), ReservedRial: strconv.FormatInt(account.ReservedRial, 10), IntervalDays: account.IntervalDays, GraceDays: account.GraceDays, RecurringFeeRial: strconv.FormatInt(account.RecurringFeeRial, 10), Invoices: []InvoiceView{}, Ledger: []LedgerView{}}

	var invoices []model.BillingInvoice

	if err := b.db.Where("reseller_id = ?", resellerID).Order("id DESC").Limit(100).Find(&invoices).Error; err != nil {
		return result, err
	}

	for _, invoice := range invoices {
		result.Invoices = append(result.Invoices, InvoiceView{invoice.ID, invoice.PeriodStart, invoice.PeriodEnd, invoice.DueAt, invoice.GraceUntil, strconv.FormatInt(invoice.TotalRial, 10), strconv.FormatInt(invoice.DueRial, 10), strconv.FormatInt(invoice.RemainingRial, 10)})
	}

	var entries []model.LedgerEntry

	if err := b.db.Where("reseller_id = ?", resellerID).Order("created_at DESC").Order("id DESC").Limit(100).Find(&entries).Error; err != nil {
		return result, err
	}

	for _, entry := range entries {
		result.Ledger = append(result.Ledger, LedgerView{entry.ID, entry.CreatedAt, entry.Kind, strconv.FormatInt(entry.DeltaRial, 10), strconv.FormatInt(entry.BalanceAfterRial, 10), entry.Reason})
	}

	return result, nil
}
