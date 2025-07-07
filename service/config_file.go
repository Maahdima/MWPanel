package service

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"mikrotik-wg-go/dataservice/model"
	"os"
)

// TODO : Make these configurable
var (
	allowedIpsExcludeLocal = ""
	allowedIpsIncludeLocal = "0.0.0.0/0, ::/0"
	defaultDns             = "8.8.8.8, 1.1.1.1"
)

type ConfigGenerator struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewConfigGenerator(db *gorm.DB) *ConfigGenerator {
	return &ConfigGenerator{
		db:     db,
		logger: zap.L().Named("ConfigGenerator"),
	}
}

func (c *ConfigGenerator) GetPeerConfig(id uint) (configPath string, err error) {
	var peer model.Peer
	if err = c.db.First(&peer, "id = ?", id).Error; err != nil {
		c.logger.Error("failed to get peer from database", zap.Uint("id", id), zap.Error(err))
		return
	}

	configPath = fmt.Sprintf("./%s/%s.conf", peerConfigsPath, peer.Name)

	return configPath, nil
}

func (c *ConfigGenerator) BuildPeerConfig(config string, peerName string) error {
	dirPath := fmt.Sprintf("./%s", peerConfigsPath)

	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	filePath := fmt.Sprintf("%s/%s.conf", dirPath, peerName)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(config); err != nil {
		return fmt.Errorf("failed to write config to file: %w", err)
	}

	return nil
}
