package seeds

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"mikrotik-wg-go/config"
	"mikrotik-wg-go/dataservice/model"
)

func AdminSeed(db *gorm.DB) error {
	var adminCount int64
	err := db.Find(&model.Admin{}).Count(&adminCount).Error
	if err != nil {
		fmt.Printf("Error when count admins: %s\n", err.Error())
		return err
	}

	if adminCount > 0 {
		return nil
	}

	adminConfig := config.GetAdminConfig()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminConfig.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error when hashing admin password: %s\n", err.Error())
		return err
	}

	admin := []model.Admin{
		{
			Username: adminConfig.Username,
			Password: string(hashedPassword),
		},
	}

	err = db.Save(&admin).Error
	if err != nil {
		fmt.Printf("Error when create admin record: %s\n", err.Error())
		return err
	}

	return nil
}
