package service

import (
	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

// queueBillingRestriction records desired changes in the same transaction as a
// financial state transition. Router failures can never erase that intent.
func queueBillingRestriction(tx *gorm.DB, resellerID uint, blocked bool) error {
	if err := tx.Model(&model.Peer{}).Where("reseller_id = ? AND restrictions_initialized = ?", resellerID, false).Updates(map[string]any{"manual_disabled": gorm.Expr("disabled"), "restrictions_initialized": true}).Error; err != nil {
		return err
	}
	return tx.Model(&model.Peer{}).Where("reseller_id = ? AND billing_blocked <> ?", resellerID, blocked).Updates(map[string]any{
		"billing_blocked": blocked, "disabled": gorm.Expr("manual_disabled OR quota_blocked OR expiry_blocked OR admin_blocked OR ?", blocked), "restriction_revision": gorm.Expr("restriction_revision + 1"), "sync_next_attempt_at": 0,
	}).Error
}
