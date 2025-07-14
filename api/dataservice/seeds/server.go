package seeds

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"mikrotik-wg-go/config"
	"mikrotik-wg-go/dataservice/model"
)

func ServerSeed(db *gorm.DB) error {
	var serverCount int64
	err := db.Find(&model.Server{}).Count(&serverCount).Error
	if err != nil {
		fmt.Printf("Error when count servers: %s\n", err.Error())
		return err
	}

	if serverCount > 0 {
		return nil
	}

	serverConfig := config.GetServerConfig()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(serverConfig.Password), 10)
	if err != nil {
		fmt.Printf("Error when hashing server password: %s\n", err.Error())
		return err
	}

	server := []model.Server{
		{
			Comment:   serverConfig.Comment,
			Name:      serverConfig.Name,
			IPAddress: serverConfig.IPAddress,
			APIPort:   serverConfig.APIPort,
			Username:  serverConfig.Username,
			Password:  string(hashedPassword),
		},
	}

	err = db.Save(&server).Error
	if err != nil {
		fmt.Printf("Error when create server record: %s\n", err.Error())
		return err
	}

	return nil
}
