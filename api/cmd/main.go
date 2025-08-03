package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/cmd/http-server"
	"github.com/maahdima/mwp/api/cmd/jobs"
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

	appCfg := config.GetAppConfig()

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
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		logger.Panic("Failed to create scheduler", zap.Error(err))
	}

	trafficJobInterval, _ := strconv.Atoi(appCfg.TrafficJobInterval)

	_, err = scheduler.NewJob(
		gocron.DurationJob(
			time.Duration(trafficJobInterval)*time.Second),
		gocron.NewTask(trafficCalculator.CalculatePeerTraffic))
	if err != nil {
		logger.Panic("Failed to create peer traffic calculation job", zap.Error(err))
	}

	_, err = scheduler.NewJob(
		gocron.DailyJob(
			1, gocron.NewAtTimes(
				gocron.NewAtTime(00, 00, 00))),
		gocron.NewTask(trafficCalculator.CalculateDailyTraffic))
	if err != nil {
		logger.Panic("Failed to create daily traffic calculation job", zap.Error(err))
	}

	scheduler.Start()

	// Start the HTTP server
	if err := httpserver.StartHttpServer(db, mwpClients, mikrotikAdaptor, trafficCalculator); err != nil {
		logger.Panic("Failed to start HTTP server", zap.Error(err))
	}
}
