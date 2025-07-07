package api

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"mikrotik-wg-go/service"
	"net/http"
)

func SetupMwpAPI(app *echo.Echo, authenticationService *service.Authentication, serverService *service.Server, interfaceService *service.WgInterface, peerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator, deviceDataService *service.DeviceData) {
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
	setupPeerRoutes(router, jwtConfig, peerService, peerConfigService, peerQrCodeService)
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
	interfaceGroup.GET("/stats", wgInterfaceController.GetInterfacesData)
	//interfaceGroup.GET("/wg-interface/:id", wgInterfaceController.GetWgInterfaceByID)
}

func setupPeerRoutes(router *echo.Group, jwtConfig echojwt.Config, peerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator) {
	wgPeerController := NewWgPeerController(peerService, peerConfigService, peerQrCodeService)

	peerGroup := router.Group("/peer")
	peerGroup.Use(echojwt.WithConfig(jwtConfig))

	peerGroup.GET("/keys", wgPeerController.GetPeerKeys)
	peerGroup.GET("", wgPeerController.GetPeers)
	peerGroup.POST("", wgPeerController.CreatePeer)
	peerGroup.POST("/:id/status", wgPeerController.UpdatePeerStatus)
	peerGroup.PATCH("/:id", wgPeerController.UpdatePeer)
	peerGroup.DELETE("/:id", wgPeerController.DeletePeer)
	//peerGroup.GET("/wg-peer/:id", wgPeerController.GetPeerByID)
	peerGroup.GET("/:id/config", wgPeerController.GetPeerConfig)
	peerGroup.GET("/:id/qrcode", wgPeerController.GetPeerQRCode)
	peerGroup.GET("/stats", wgPeerController.GetPeersData)
}

func setupDeviceInfoRoutes(router *echo.Group, jwtConfig echojwt.Config, deviceDataService *service.DeviceData) {
	deviceInfoController := NewDeviceDataController(deviceDataService)

	deviceGroup := router.Group("/device")
	deviceGroup.Use(echojwt.WithConfig(jwtConfig))

	deviceGroup.GET("/stats", deviceInfoController.GetDeviceInfo)
}
