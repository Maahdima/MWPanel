package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"
)

var (
	peerConfigsPath string
)

func init() {
	appCfg := config.GetAppConfig()
	peerConfigsPath = filepath.Join(appCfg.PeerFilesDir, "config")
	if err := os.MkdirAll(peerConfigsPath, os.ModePerm); err != nil {
		panic(fmt.Sprintf("failed to create peer config directory: %v", err))
	}
}

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Error("peer not found in database", zap.Uint("id", id))
			return
		}
		c.logger.Error("failed to get peer from database", zap.Uint("id", id), zap.Error(err))
		return
	}

	configPath = fmt.Sprintf("%s/%s.conf", peerConfigsPath, peer.UUID)

	return configPath, nil
}

func (c *ConfigGenerator) GetUserConfig(uuid string) (configPath string, err error) {
	var peer model.Peer

	if err = c.db.First(&peer, "uuid = ?", uuid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Error("peer not found in database", zap.String("uuid", uuid))
			return
		}
		c.logger.Error("failed to get peer from database", zap.String("uuid", uuid), zap.Error(err))
		return
	}

	utils.IsPeerSharable(peer.IsShared, peer.ShareExpireTime)
	if !peer.IsShared {
		return "", common.ErrPeerNotShared
	}

	configPath = fmt.Sprintf("%s/%s.conf", peerConfigsPath, peer.UUID)

	return configPath, nil
}

func (c *ConfigGenerator) BuildPeerConfig(config string, uuid string) error {
	filePath := fmt.Sprintf("%s/%s.conf", peerConfigsPath, uuid)

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

func (c *ConfigGenerator) RemovePeerConfig(id uint) error {
	var peer model.Peer

	if err := c.db.First(&peer, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Error("peer not found in database", zap.Uint("id", id))
			return err
		}
		c.logger.Error("failed to get peer from database", zap.Uint("id", id), zap.Error(err))
		return err
	}

	configPath := fmt.Sprintf("%s/%s.conf", peerConfigsPath, peer.UUID)

	err := os.Remove(configPath)
	if err != nil {
		c.logger.Error("failed to remove Config file", zap.String("path", configPath), zap.Error(err))
		return err
	}

	return nil
}
