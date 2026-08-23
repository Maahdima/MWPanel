package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	httpserver "github.com/maahdima/mwp/api/cmd/http-server"
	traffic "github.com/maahdima/mwp/api/cmd/jobs"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice"
	"github.com/maahdima/mwp/api/dataservice/seeds"
	"github.com/maahdima/mwp/api/service"
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
	telegramCfg := config.GetTelegramConfig()
	telegramClient := service.NewTelegramClient(telegramCfg)
	telegramNotifier := service.NewTelegramNotifier(db, telegramCfg, telegramClient)
	configGenerator := service.NewConfigGenerator(db)
	qrCodeGenerator := service.NewQRCodeGenerator(db)
	telegramBot := service.NewTelegramBot(db, telegramCfg, telegramClient, configGenerator, qrCodeGenerator, mikrotikAdaptor)
	telegramBot.Start(context.Background())
	trafficCalculator := traffic.NewTrafficCalculator(db, mikrotikAdaptor, telegramNotifier)

	// Start the traffic calculation job
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		logger.Panic("Failed to create scheduler", zap.Error(err))
	}

	trafficJobInterval, _ := strconv.Atoi(appCfg.TrafficJobInterval)
	sessionJobInterval, _ := strconv.Atoi(appCfg.SessionJobInterval)
	if sessionJobInterval <= 0 {
		sessionJobInterval = 30
	}

	_, err = scheduler.NewJob(
		gocron.DurationJob(
			time.Duration(trafficJobInterval)*time.Second),
		gocron.NewTask(trafficCalculator.CalculatePeerTraffic))
	if err != nil {
		logger.Panic("Failed to create peer traffic calculation job", zap.Error(err))
	}

	_, err = scheduler.NewJob(
		gocron.DurationJob(
			time.Duration(sessionJobInterval)*time.Second),
		gocron.NewTask(trafficCalculator.TrackPeerSessions),
		gocron.WithStartAt(gocron.WithStartImmediately()),
	)
	if err != nil {
		logger.Panic("Failed to create peer session tracking job", zap.Error(err))
	}

	_, err = scheduler.NewJob(
		gocron.DurationJob(
			time.Duration(trafficJobInterval)*time.Second),
		gocron.NewTask(trafficCalculator.ExpireOverduePeers),
		gocron.WithStartAt(gocron.WithStartImmediately()),
	)
	if err != nil {
		logger.Panic("Failed to create peer expiration job", zap.Error(err))
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
	if err := httpserver.StartHttpServer(db, mwpClients, mikrotikAdaptor, trafficCalculator, telegramBot); err != nil {
		logger.Panic("Failed to start HTTP server", zap.Error(err))
	}
}
