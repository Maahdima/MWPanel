package api

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"mikrotik-wg-go/service"
	"net/http"
)

func SetupMwpAPI(app *echo.Echo, authenticationService *service.Authentication, peerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator, deviceDataService *service.DeviceData, interfaceService *service.WgInterface) {
	router := app.Group("/api")

	// TODO : read from config environment variables
	jwtConfig := echojwt.Config{
		SigningKey: []byte("access_secret"),
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		},
	}

	setupAuthenticationRoutes(router, authenticationService)
	setupWgInterfaceRoutes(router, jwtConfig, interfaceService)
	setupWgPeerRoutes(router, jwtConfig, peerService, peerConfigService, peerQrCodeService)
	setupDeviceInfoRoutes(router, jwtConfig, deviceDataService)
}

func setupAuthenticationRoutes(router *echo.Group, authService *service.Authentication) {
	authController := NewAuthController(authService)

	authGroup := router.Group("/auth")

	authGroup.POST("/login", authController.Login)
	//router.POST("/refresh-token", authController.RefreshToken)
	//router.GET("/logout", authController.Logout)
}

func setupWgInterfaceRoutes(router *echo.Group, jwtConfig echojwt.Config, interfaceService *service.WgInterface) {
	wgInterfaceController := NewWgInterfaceController(interfaceService)

	interfaceGroup := router.Group("/interface")
	interfaceGroup.Use(echojwt.WithConfig(jwtConfig))

	interfaceGroup.GET("", wgInterfaceController.GetInterfaces)
	//router.POST("/wg-interfaces", wgInterfaceController.CreateInterface)
	//router.DELETE("/wg-interfaces/:name", wgInterfaceController.DeleteInterface)
	interfaceGroup.GET("/stats", wgInterfaceController.GetInterfacesData)
	//router.GET("/wg-interface/:id", wgInterfaceController.GetWgInterfaceByID)
	//router.PUT("/wg-interface/:id", wgInterfaceController.UpdateWgInterface)
}

func setupWgPeerRoutes(router *echo.Group, jwtConfig echojwt.Config, peerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator) {
	wgPeerController := NewWgPeerController(peerService, peerConfigService, peerQrCodeService)

	// TODO : read from config environment variables
	peerGroup := router.Group("/peer")
	peerGroup.Use(echojwt.WithConfig(jwtConfig))

	peerGroup.GET("/keys", wgPeerController.GetPeerKeys)
	peerGroup.GET("", wgPeerController.GetPeers)
	peerGroup.POST("", wgPeerController.CreatePeer)
	peerGroup.POST("/:id/status", wgPeerController.UpdatePeerStatus)
	peerGroup.PATCH("/:id", wgPeerController.UpdatePeer)
	peerGroup.DELETE("/:id", wgPeerController.DeletePeer)
	//router.GET("/wg-peer/:id", wgPeerController.GetPeerByID)
	peerGroup.GET("/:id/config", wgPeerController.GetPeerConfig)
	peerGroup.GET("/:id/qrcode", wgPeerController.GetPeerQRCode)
	peerGroup.GET("/stats", wgPeerController.GetPeersData)
}

func setupDeviceInfoRoutes(router *echo.Group, jwtConfig echojwt.Config, deviceDataService *service.DeviceData) {
	deviceInfoController := NewDeviceDataController(deviceDataService)

	// TODO : read from config environment variables
	deviceGroup := router.Group("/device")
	deviceGroup.Use(echojwt.WithConfig(jwtConfig))

	deviceGroup.GET("/stats", deviceInfoController.GetDeviceInfo)
}
