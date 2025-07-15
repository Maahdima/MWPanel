package main

import (
	"fmt"
	"time"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/cmd/traffic-job"
	"github.com/maahdima/mwp/api/cmd/web-server"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice"
	"github.com/maahdima/mwp/api/dataservice/seeds"
	"github.com/maahdima/mwp/api/utils/log"

	"go.uber.org/zap"
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

	// TODO: add db migration (gorm)

	err = seeds.AdminSeed(db)
	if err != nil {
		fmt.Printf("cannot seed admin [%s]", err.Error())
		logger.Panic("cannot seed admin", zap.Error(err))
	}

	// Initialize http client for Mikrotik API
	mwpClients := common.NewMwpClients(db)
	mwpClients.InitClient()

	mikrotikAdaptor := mikrotik.NewAdaptor(mwpClients)
	trafficCalculator := traffic.NewTrafficCalculator(db, mikrotikAdaptor)

	// Start the traffic calculation job
	go func() {
		// TODO : get checker interval from config
		for range time.Tick(30 * time.Second) {
			httpClient := mwpClients.GetClient(nil)
			if httpClient != nil {
				trafficCalculator.CalculateTraffic()
			}
		}
	}()

	// Start the HTTP server
	if err := webserver.StartHttpServer(db, mwpClients, mikrotikAdaptor, trafficCalculator); err != nil {
		logger.Panic("Failed to start HTTP server", zap.Error(err))
	}
}
