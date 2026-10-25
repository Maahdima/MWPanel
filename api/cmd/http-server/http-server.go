package httpserver

import (
	"fmt"
	"path/filepath"

	"golang.org/x/crypto/acme/autocert"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	traffic "github.com/maahdima/mwp/api/cmd/jobs"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/http"
	"github.com/maahdima/mwp/api/service"
	"github.com/maahdima/mwp/api/shopclient"
	"github.com/maahdima/mwp/api/utils/validate"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

func StartHttpServer(
	db *gorm.DB,
	mwpClients *common.MwpClient,
	mikrotikAdaptor *mikrotik.Adaptor,
	trafficCalculator *traffic.Calculator,
	shopAPI *shopclient.Client,
) error {
	appCfg := config.GetAppConfig()

	authenticationService := service.NewAuthentication(db)
	schedulerService := service.NewScheduler(mikrotikAdaptor)
	queueService := service.NewQueue(mikrotikAdaptor)
	configGenerator := service.NewConfigGenerator(db)
	qrCodeGenerator := service.NewQRCodeGenerator(db)
	excelGenerator := service.NewExcelGenerator(db)
	serverService := service.NewServerService(db, mwpClients, mikrotikAdaptor)
	peerService := service.NewWGPeer(db, mikrotikAdaptor, schedulerService, queueService, configGenerator, qrCodeGenerator)
	peerService.SetShopLinks(shopAPI)
	interfaceService := service.NewWgInterface(db, mikrotikAdaptor, peerService)
	ipPoolService := service.NewIPPool(db, mikrotikAdaptor)
	deviceDataService := service.NewDeviceData(db, mikrotikAdaptor, serverService, interfaceService, peerService)
	syncService := service.NewSyncService(db, mikrotikAdaptor, configGenerator, qrCodeGenerator, peerService, interfaceService)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Validator = &validate.CustomValidator{Validator: validator.New()}

	http.SetupMwpUI(e, appCfg.UIAssetsFs, appCfg.AdminPanelPath)
	http.SetupMwpAPI(
		e,
		mwpClients,
		authenticationService,
		serverService,
		interfaceService,
		ipPoolService,
		peerService,
		configGenerator,
		qrCodeGenerator,
		excelGenerator,
		deviceDataService,
		trafficCalculator,
		syncService,
		shopAPI,
		service.NewResellers(db, peerService),
		service.NewBilling(db),
		service.NewPayments(service.NewBilling(db), filepath.Join(appCfg.DataDirPath, "receipts")),
	)

	if appCfg.Port == "443" || appCfg.Port == "8443" {
		e.AutoTLSManager.Cache = autocert.DirCache(appCfg.DataDirPath)
		e.Logger.Fatal(e.StartAutoTLS(fmt.Sprintf("%s:%s", appCfg.Host, appCfg.Port)))
	} else {
		e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", appCfg.Host, appCfg.Port)))
	}

	return nil
}
