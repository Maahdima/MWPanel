package main

import (
	"go.uber.org/zap"
	webserver "mikrotik-wg-go/cmd/web-server"
	"mikrotik-wg-go/config"
	"mikrotik-wg-go/dataservice"
	"mikrotik-wg-go/utils/log"
)

func init() {
	log.InitLogger(config.GetAppConfig())
}

func main() {
	logger := zap.L()

	db, err := dataservice.ConnectDB(config.GetDBConfig())
	if err != nil {
		logger.Panic("Failed to connect to database", zap.Error(err))
	}

	if err := dataservice.AutoMigrate(db); err != nil {
		logger.Panic("Failed to auto-migrate database", zap.Error(err))
	}

	if err := webserver.StartHttpServer(db); err != nil {
		logger.Panic("Failed to start HTTP server", zap.Error(err))
	}
}
