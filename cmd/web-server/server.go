package webserver

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"mikrotik-wg-go/adaptor/mikrotik"
	"mikrotik-wg-go/api"
	"mikrotik-wg-go/api/schema"
	"mikrotik-wg-go/config"
	"mikrotik-wg-go/dataservice/db"
	"mikrotik-wg-go/service"
	"mikrotik-wg-go/utils/httphelper"
	"mikrotik-wg-go/utils/log"
	"mikrotik-wg-go/utils/validate"
	"net/http"
)

func init() {
	log.InitLogger(config.GetAppConfig())
}

func StartHttpServer(db *db.Queries) error {
	logger := zap.L()

	client, err := httphelper.NewClient(httphelper.Config{
		BaseURL:            "http://192.168.88.237/rest",
		Username:           "maahdima",
		Password:           "M@hdima7731$$",
		InsecureSkipVerify: true,
	})
	if err != nil {
		logger.Panic("Failed to create HTTP client", zap.Error(err))
		return err
	}

	mikrotikAdaptor := mikrotik.NewAdaptor(client)
	configGenerator := service.NewConfigGenerator(db)
	qrCodeGenerator := service.NewQRCodeGenerator(db)
	peerService := service.NewWGPeer(db, mikrotikAdaptor, configGenerator)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Validator = &validate.CustomValidator{Validator: validator.New()}

	api.SetupMwpAPI(e, peerService, configGenerator, qrCodeGenerator)

	// TODO : move all these into controllers and api package
	e.GET("/device-info", func(c echo.Context) error {
		info, err := mikrotikAdaptor.FetchDeviceResource(c.Request().Context())
		if err != nil {
			logger.Error("Failed to fetch device info", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, schema.ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Status:     "error",
				Message:    "Failed to fetch device info: " + err.Error(),
			})
		}

		return c.JSON(http.StatusOK, info)
	})

	e.GET("/wg-interfaces", func(c echo.Context) error {
		info, err := mikrotikAdaptor.FetchWgInterfaces(c.Request().Context())
		if err != nil {
			logger.Error("Failed to fetch wireguard interfaces", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, info)
	})

	e.GET("/generate-key", func(c echo.Context) error {
		privateKey, publicKey, err := peerService.GenerateKeys()
		if err != nil {
			logger.Error("Failed to generate keys", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"private_key": privateKey,
			"public_key":  publicKey,
		})
	})

	e.Logger.Fatal(e.Start(":1323"))

	return nil
}
