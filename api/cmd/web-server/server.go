package webserver

import (
	"fmt"

	"golang.org/x/crypto/acme/autocert"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/cmd/jobs"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/http"
	"github.com/maahdima/mwp/api/service"
	"github.com/maahdima/mwp/api/utils/validate"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

func StartHttpServer(db *gorm.DB, mwpClients *common.MwpClients, mikrotikAdaptor *mikrotik.Adaptor, trafficCalculator *traffic.Calculator) error {
	appCfg := config.GetAppConfig()

	authenticationService := service.NewAuthentication(db)
	schedulerService := service.NewScheduler(mikrotikAdaptor)
	queueService := service.NewQueue(mikrotikAdaptor)
	configGenerator := service.NewConfigGenerator(db)
	qrCodeGenerator := service.NewQRCodeGenerator(db)
	serverService := service.NewServerService(db, mwpClients, mikrotikAdaptor)
	interfaceService := service.NewWgInterface(db, mikrotikAdaptor)
	ipPoolService := service.NewIPPool(db)
	peerService := service.NewWGPeer(db, mikrotikAdaptor, schedulerService, queueService, configGenerator, qrCodeGenerator)
	deviceDataService := service.NewDeviceData(db, mikrotikAdaptor, serverService, interfaceService, peerService)
	syncService := service.NewSyncService(db, mikrotikAdaptor, configGenerator, qrCodeGenerator)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Validator = &validate.CustomValidator{Validator: validator.New()}

	http.SetupMwpUI(e, appCfg.UIAssetsFs)
	http.SetupMwpAPI(e, mwpClients, authenticationService, serverService, interfaceService, ipPoolService, peerService, configGenerator, qrCodeGenerator, deviceDataService, trafficCalculator, syncService)

	if appCfg.Port == "443" || appCfg.Port == "8443" {
		e.AutoTLSManager.Cache = autocert.DirCache(appCfg.DataDirPath)
		e.Logger.Fatal(e.StartAutoTLS(fmt.Sprintf("%s:%s", appCfg.Host, appCfg.Port)))
	} else {
		e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", appCfg.Host, appCfg.Port)))
	}

	return nil
}
