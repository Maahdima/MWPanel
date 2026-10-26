package service

import (
	"context"
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/maahdima/mwp/api/dataservice/model"
)

// Complete histories are available independently of the dashboard's recent-100
// view. Rows stream into a private temporary file; incomplete exports are never
// sent as successful downloads, and request cancellation stops the SQL cursor.
func (b *Billing) ExportHistory(ctx context.Context, actor Principal, resellerID uint, kind string) (path string, err error) {
	if !billingAccess(actor, resellerID) {
		return "", ErrRoleForbidden
	}
	if kind != "ledger" && kind != "invoices" {
		return "", ErrInvalidBillingQuantity
	}
	file, err := os.CreateTemp("", "mwp-billing-*.csv")
	if err != nil {
		return "", err
	}
	temporary := file.Name()
	path = temporary
	success := false
	defer func() {
		file.Close()
		if !success {
			os.Remove(temporary)
		}
	}()
	writer := csv.NewWriter(file)
	query := b.db.WithContext(ctx).Where("reseller_id = ?", resellerID)
	if kind == "ledger" {
		if err = writer.Write([]string{"id", "created_at_utc", "kind", "change_rial", "balance_after_rial", "period_start_utc", "reason"}); err != nil {
			return "", err
		}
		rows, e := query.Model(&model.LedgerEntry{}).Order("created_at ASC").Order("id ASC").Rows()
		if e != nil {
			return "", e
		}
		defer rows.Close()
		for rows.Next() {
			var row model.LedgerEntry
			if e = b.db.ScanRows(rows, &row); e != nil {
				return "", e
			}
			if e = writer.Write([]string{row.ID, time.Unix(row.CreatedAt, 0).UTC().Format(time.RFC3339), csvText(row.Kind), strconv.FormatInt(row.DeltaRial, 10), strconv.FormatInt(row.BalanceAfterRial, 10), time.Unix(row.PeriodStart, 0).UTC().Format(time.RFC3339), csvText(row.Reason)}); e != nil {
				return "", e
			}
		}
		if e = rows.Err(); e != nil {
			return "", e
		}
	} else {
		if err = writer.Write([]string{"id", "period_start_utc", "period_end_utc", "due_at_utc", "grace_until_utc", "total_rial", "due_rial", "remaining_rial"}); err != nil {
			return "", err
		}
		rows, e := query.Model(&model.BillingInvoice{}).Order("id ASC").Rows()
		if e != nil {
			return "", e
		}
		defer rows.Close()
		for rows.Next() {
			var row model.BillingInvoice
			if e = b.db.ScanRows(rows, &row); e != nil {
				return "", e
			}
			if e = writer.Write([]string{strconv.FormatUint(uint64(row.ID), 10), time.Unix(row.PeriodStart, 0).UTC().Format(time.RFC3339), time.Unix(row.PeriodEnd, 0).UTC().Format(time.RFC3339), time.Unix(row.DueAt, 0).UTC().Format(time.RFC3339), time.Unix(row.GraceUntil, 0).UTC().Format(time.RFC3339), strconv.FormatInt(row.TotalRial, 10), strconv.FormatInt(row.DueRial, 10), strconv.FormatInt(row.RemainingRial, 10)}); e != nil {
				return "", e
			}
		}
		if e = rows.Err(); e != nil {
			return "", e
		}
	}
	writer.Flush()
	if err = writer.Error(); err != nil {
		return "", err
	}
	if err = file.Close(); err != nil {
		return "", err
	}
	success = true
	return path, nil
}
func csvText(value string) string {
	trimmed := strings.TrimLeft(value, " \t\r\n")
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}
