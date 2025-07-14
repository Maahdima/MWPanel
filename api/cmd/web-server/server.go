package webserver

import (
	"fmt"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/cmd/traffic-job"
	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/http"
	"github.com/maahdima/mwp/api/service"
	"github.com/maahdima/mwp/api/utils/validate"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

func StartHttpServer(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor, trafficCalculator *traffic.Calculator) error {
	appCfg := config.GetAppConfig()

	// TODO : initial check for connection to Mikrotik device
	authenticationService := service.NewAuthentication(db)
	schedulerService := service.NewScheduler(mikrotikAdaptor)
	queueService := service.NewQueue(mikrotikAdaptor)
	configGenerator := service.NewConfigGenerator(db)
	qrCodeGenerator := service.NewQRCodeGenerator(db)
	serverService := service.NewServerService(db, mikrotikAdaptor)
	interfaceService := service.NewWgInterface(db, mikrotikAdaptor)
	peerService := service.NewWGPeer(db, mikrotikAdaptor, schedulerService, queueService, configGenerator)
	deviceDataService := service.NewDeviceData(mikrotikAdaptor, serverService, interfaceService, peerService)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Validator = &validate.CustomValidator{Validator: validator.New()}

	http.SetupMwpUI(e, appCfg.UIAssetsFs)
	http.SetupMwpAPI(e, authenticationService, serverService, interfaceService, peerService, configGenerator, qrCodeGenerator, deviceDataService, trafficCalculator)

	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", appCfg.Host, appCfg.Port)))

	return nil
}
