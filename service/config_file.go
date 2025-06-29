package service

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"mikrotik-wg-go/dataservice/db"
	"os"
)

// TODO : Make these configurable
var (
	allowedIpsExcludeLocal = ""
	allowedIpsIncludeLocal = "0.0.0.0/0, ::/0"
	defaultDns             = "8.8.8.8, 1.1.1.1"
)

type ConfigGenerator struct {
	db     *db.Queries
	logger *zap.Logger
}

func NewConfigGenerator(db *db.Queries) *ConfigGenerator {
	return &ConfigGenerator{
		db:     db,
		logger: zap.L().Named("ConfigGenerator"),
	}
}

func (c *ConfigGenerator) GetPeerConfig(id int64) (configPath string, err error) {
	peer, err := c.db.GetPeer(context.Background(), id)
	if err != nil {
		c.logger.Error("failed to get peer from database", zap.Int64("id", id), zap.Error(err))
		return
	}

	configPath = fmt.Sprintf("./%s/%s.conf", peerConfigsPath, peer.PeerName)

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
