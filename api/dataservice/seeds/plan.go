package seeds

import (
	"fmt"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"

	"gorm.io/gorm"
)

// defaultPlanGB is the traffic sizes offered in the Telegram shop.
var defaultPlanGB = []int{2, 5, 10, 20, 30, 50, 100}

func PlanSeed(db *gorm.DB) error {
	for i, gb := range defaultPlanGB {
		var count int64
		err := db.Model(&model.Plan{}).
			Where("traffic_gb = ? AND duration_days = ?", gb, 30).
			Count(&count).Error
		if err != nil {
			return fmt.Errorf("count vpn plans: %w", err)
		}
		if count > 0 {
			continue
		}

		faName := utils.PersianPlanTitle(gb, 30)
		plan := model.Plan{
			Name:         fmt.Sprintf("%d GB / 30 days", gb),
			FaName:       faName,
			Title:        faName,
			TrafficGB:    gb,
			DurationDays: 30,
			Enabled:      true,
			SortOrder:    i,
		}
		if err := db.Create(&plan).Error; err != nil {
			return fmt.Errorf("seed vpn plan %dGB: %w", gb, err)
		}
	}
	return nil
}
