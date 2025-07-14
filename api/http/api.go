package http

import (
	"net/http"

	"github.com/maahdima/mwp/api/cmd/traffic-job"
	"github.com/maahdima/mwp/api/service"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func SetupMwpAPI(app *echo.Echo, authenticationService *service.Authentication, serverService *service.Server, interfaceService *service.WgInterface, peerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator, deviceDataService *service.DeviceData, trafficCalculator *traffic.Calculator) {
	router := app.Group("/api")

	// TODO : read from config environment variables
	jwtConfig := echojwt.Config{
		SigningKey: []byte("access_secret"),
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		},
	}

	setupAuthenticationRoutes(router, authenticationService)
	setupServerRoutes(router, jwtConfig, serverService)
	setupInterfaceRoutes(router, jwtConfig, interfaceService)
	setupPeerRoutes(router, jwtConfig, peerService, peerConfigService, peerQrCodeService, trafficCalculator)
	setupDeviceInfoRoutes(router, jwtConfig, deviceDataService)
}

func setupAuthenticationRoutes(router *echo.Group, authService *service.Authentication) {
	authController := NewAuthController(authService)

	authGroup := router.Group("/auth")

	authGroup.POST("/login", authController.Login)
	//router.POST("/refresh-token", authController.RefreshToken)
	//router.GET("/logout", authController.Logout)
}

func setupServerRoutes(router *echo.Group, jwtConfig echojwt.Config, serverService *service.Server) {
	serverController := NewServerController(serverService)

	serverGroup := router.Group("/server")
	serverGroup.Use(echojwt.WithConfig(jwtConfig))

	serverGroup.GET("", serverController.GetServers)
	serverGroup.POST("", serverController.CreateServer)
	//serverGroup.GET("/server/:id", serverController.GetServerByID)
	serverGroup.POST("/:id/status", serverController.UpdateServerStatus)
	serverGroup.PATCH("/:id", serverController.UpdateServer)
	serverGroup.DELETE("/:id", serverController.DeleteServer)
}

func setupInterfaceRoutes(router *echo.Group, jwtConfig echojwt.Config, interfaceService *service.WgInterface) {
	wgInterfaceController := NewWgInterfaceController(interfaceService)

	interfaceGroup := router.Group("/interface")
	interfaceGroup.Use(echojwt.WithConfig(jwtConfig))

	interfaceGroup.GET("", wgInterfaceController.GetInterfaces)
	interfaceGroup.POST("", wgInterfaceController.CreateInterface)
	interfaceGroup.POST("/:id/status", wgInterfaceController.UpdateInterfaceStatus)
	interfaceGroup.PATCH("/:id", wgInterfaceController.UpdateInterface)
	interfaceGroup.DELETE("/:id", wgInterfaceController.DeleteInterface)
	//interfaceGroup.GET("/wg-interface/:id", wgInterfaceController.GetWgInterfaceByID)
}

func setupPeerRoutes(router *echo.Group, jwtConfig echojwt.Config, peerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator, trafficCalculator *traffic.Calculator) {
	wgPeerController := NewWgPeerController(peerService, peerConfigService, peerQrCodeService, trafficCalculator)

	peerGroup := router.Group("/peer")

	peerGroup.GET("/keys", wgPeerController.GetPeerKeys, echojwt.WithConfig(jwtConfig))
	peerGroup.GET("", wgPeerController.GetPeers, echojwt.WithConfig(jwtConfig))
	peerGroup.POST("", wgPeerController.CreatePeer, echojwt.WithConfig(jwtConfig))
	peerGroup.PATCH("/:id/status", wgPeerController.UpdatePeerStatus, echojwt.WithConfig(jwtConfig))
	peerGroup.PATCH("/:id/reset-usage", wgPeerController.ResetPeerUsage, echojwt.WithConfig(jwtConfig))
	peerGroup.PUT("/:id", wgPeerController.UpdatePeer, echojwt.WithConfig(jwtConfig))
	peerGroup.DELETE("/:id", wgPeerController.DeletePeer, echojwt.WithConfig(jwtConfig))
	peerGroup.GET("/:uuid/config", wgPeerController.GetPeerConfig)
	peerGroup.GET("/:uuid/qrcode", wgPeerController.GetPeerQRCode)
	peerGroup.GET("/:uuid/details", wgPeerController.GetPeerDetails)
}

func setupDeviceInfoRoutes(router *echo.Group, jwtConfig echojwt.Config, deviceDataService *service.DeviceData) {
	deviceInfoController := NewDeviceDataController(deviceDataService)

	deviceGroup := router.Group("/device")
	deviceGroup.Use(echojwt.WithConfig(jwtConfig))

	deviceGroup.GET("/stats", deviceInfoController.GetDeviceInfo)
}
