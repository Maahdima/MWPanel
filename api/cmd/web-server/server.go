package webserver

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"mikrotik-wg-go/adaptor/mikrotik"
	"mikrotik-wg-go/config"
	"mikrotik-wg-go/http"
	"mikrotik-wg-go/service"
	"mikrotik-wg-go/utils/httphelper"
	"mikrotik-wg-go/utils/validate"
	"path/filepath"
)

func StartHttpServer(db *gorm.DB) error {
	appCfg := config.GetAppConfig()
	logger := zap.L()

	// TODO : initial check for connection to Mikrotik device
	client, err := httphelper.NewClient(httphelper.Config{
		BaseURL:            "http://192.168.64.2/rest",
		Username:           "admin",
		Password:           "admin1234$",
		InsecureSkipVerify: true,
	})
	if err != nil {
		logger.Panic("Failed to create HTTP client", zap.Error(err))
		return err
	}

	mikrotikAdaptor := mikrotik.NewAdaptor(client)
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

	publicDir := echo.MustSubFS(e.Filesystem, appCfg.PublicDir)
	staticFilesHandler := echo.StaticDirectoryHandler(publicDir, false)

	e.GET(
		"/*",
		func(c echo.Context) error {
			if err := staticFilesHandler(c); err != nil {
				return c.File(filepath.Join(appCfg.PublicDir, "index.html"))
			}

			return nil
		},
	)

	http.SetupMwpAPI(e, authenticationService, serverService, interfaceService, peerService, configGenerator, qrCodeGenerator, deviceDataService)

	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", appCfg.Host, appCfg.Port)))

	return nil
}
