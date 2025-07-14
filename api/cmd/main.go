package main

import (
	"fmt"
	"time"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/cmd/traffic-job"
	"github.com/maahdima/mwp/api/cmd/web-server"
	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice"
	"github.com/maahdima/mwp/api/dataservice/seeds"
	"github.com/maahdima/mwp/api/utils/httphelper"
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

	// TODO: add atlas migration
	//if err = dataservice.AtlasMigrate(config.GetDBConfig()); err != nil {
	//	fmt.Printf("cannot apply atlas migration [%s]", err.Error())
	//	logger.Panic("cannot apply atlas migration", zap.Error(err))
	//}

	err = seeds.AdminSeed(db)
	if err != nil {
		fmt.Printf("cannot seed admin [%s]", err.Error())
		logger.Panic("cannot seed admin", zap.Error(err))
	}

	err = seeds.ServerSeed(db)
	if err != nil {
		fmt.Printf("cannot seed server [%s]", err.Error())
		logger.Panic("cannot seed server", zap.Error(err))
	}

	serverConfig := config.GetServerConfig()

	client, err := httphelper.NewClient(httphelper.Config{
		BaseURL:            fmt.Sprintf("%s://%s:%d/rest", "http", serverConfig.IPAddress, serverConfig.APIPort),
		Username:           serverConfig.Username,
		Password:           serverConfig.Password,
		InsecureSkipVerify: true,
	})
	if err != nil {
		logger.Panic("Failed to create HTTP client", zap.Error(err))
	}

	mikrotikAdaptor := mikrotik.NewAdaptor(client)
	trafficCalculator := traffic.NewTrafficCalculator(db, mikrotikAdaptor)

	// Start the traffic calculation job
	go func() {
		// TODO : get checker interval from config
		for range time.Tick(30 * time.Second) {
			trafficCalculator.CalculateTraffic()
		}
	}()

	// Start the HTTP server
	if err := webserver.StartHttpServer(db, mikrotikAdaptor, trafficCalculator); err != nil {
		logger.Panic("Failed to start HTTP server", zap.Error(err))
	}
}
