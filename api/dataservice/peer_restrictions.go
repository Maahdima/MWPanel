package dataservice

import (
	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

// Legacy disable reasons cannot be inferred reliably. Keep them conservatively
// manual until explicitly enabled; billing recovery must never guess otherwise.
func MigratePeerRestrictions(db *gorm.DB) error {
	return db.Model(&model.Peer{}).Where("restrictions_initialized = ?", false).Updates(map[string]any{"manual_disabled": gorm.Expr("disabled"), "restrictions_initialized": true}).Error
}
