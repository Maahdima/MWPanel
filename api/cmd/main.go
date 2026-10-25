package main

import (
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
	"github.com/maahdima/mwp/api/shopclient"
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
	mwpClients := common.NewMwpClient(db)
	mwpClients.InitClient()

	mikrotikAdaptor := mikrotik.NewAdaptor(mwpClients)
	integration := config.GetIntegrationConfig()
	shopAPI := shopclient.New(integration.ShopURL, integration.ServiceToken)
	trafficCalculator := traffic.NewTrafficCalculator(db, mikrotikAdaptor, shopclient.NewNotifier(shopAPI))

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

	peerReconciler := service.NewPeerReconciler(db, mikrotikAdaptor, service.NewBilling(db))
	_, err = scheduler.NewJob(gocron.DurationJob(10*time.Second), gocron.NewTask(peerReconciler.Run), gocron.WithStartAt(gocron.WithStartImmediately()))
	if err != nil {
		logger.Panic("Failed to create peer reconciliation job", zap.Error(err))
	}
	purchasePeers := service.NewWGPeer(db, mikrotikAdaptor, service.NewScheduler(mikrotikAdaptor), service.NewQueue(mikrotikAdaptor), service.NewConfigGenerator(db), service.NewQRCodeGenerator(db))
	purchases := service.NewPeerPurchases(service.NewBilling(db), purchasePeers)
	_, err = scheduler.NewJob(gocron.DurationJob(10*time.Second), gocron.NewTask(purchases.Run), gocron.WithStartAt(gocron.WithStartImmediately()))
	if err != nil {
		logger.Panic("Failed to create purchase reconciliation job", zap.Error(err))
	}
	scheduler.Start()

	// Start the HTTP server
	if err := httpserver.StartHttpServer(db, mwpClients, mikrotikAdaptor, trafficCalculator, shopAPI); err != nil {
		logger.Panic("Failed to start HTTP server", zap.Error(err))
	}
}
