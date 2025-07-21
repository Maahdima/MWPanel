package http

import (
	"net/http"

	"github.com/maahdima/mwp/api/cmd/traffic-job"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/http/middleware"
	"github.com/maahdima/mwp/api/service"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func SetupMwpAPI(app *echo.Echo, mwpClients *common.MwpClients, authenticationService *service.Authentication, serverService *service.Server, interfaceService *service.WgInterface, peerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator, deviceDataService *service.DeviceData, trafficCalculator *traffic.Calculator, syncService *service.SyncService) {
	router := app.Group("/api")

	// TODO : read from config environment variables
	jwtConfig := echojwt.Config{
		SigningKey: []byte("access_secret"),
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		},
	}

	authController := NewAuthController(authenticationService)
	serverController := NewServerController(serverService)
	wgInterfaceController := NewWgInterfaceController(interfaceService)
	wgPeerController := NewWgPeerController(peerService, peerConfigService, peerQrCodeService, trafficCalculator)
	deviceInfoController := NewDeviceDataController(deviceDataService)
	syncController := NewSyncController(syncService)
	userController := NewUserController(peerService, peerConfigService, peerQrCodeService)

	setupAuthenticationRoutes(router, authController)
	setupServerRoutes(router, jwtConfig, serverController)
	setupInterfaceRoutes(router, mwpClients, jwtConfig, wgInterfaceController)
	setupPeerRoutes(router, mwpClients, jwtConfig, wgPeerController)
	setupDeviceInfoRoutes(router, mwpClients, jwtConfig, deviceInfoController)
	setupSyncRoutes(router, mwpClients, jwtConfig, syncController)
	setupUserRoutes(router, userController)
}

func setupAuthenticationRoutes(router *echo.Group, authController *AuthController) {
	authGroup := router.Group("/auth")

	authGroup.POST("/login", authController.Login)
	//router.POST("/refresh-token", authController.RefreshToken)
	//router.GET("/logout", authController.Logout)
}

func setupServerRoutes(router *echo.Group, jwtConfig echojwt.Config, serverController *ServerController) {
	serverGroup := router.Group("/server")
	serverGroup.Use(echojwt.WithConfig(jwtConfig))

	serverGroup.GET("", serverController.GetServers)
	serverGroup.POST("", serverController.CreateServer)
	//serverGroup.GET("/server/:id", serverController.GetServerByID)
	serverGroup.POST("/:id/status", serverController.UpdateServerStatus)
	serverGroup.PATCH("/:id", serverController.UpdateServer)
	serverGroup.DELETE("/:id", serverController.DeleteServer)
}

func setupInterfaceRoutes(router *echo.Group, mwpClients *common.MwpClients, jwtConfig echojwt.Config, wgInterfaceController *WgInterfaceController) {
	interfaceGroup := router.Group("/interface")
	interfaceGroup.Use(echojwt.WithConfig(jwtConfig), middleware.ClientConnectionMiddleware(mwpClients))

	interfaceGroup.GET("", wgInterfaceController.GetInterfaces)
	interfaceGroup.POST("", wgInterfaceController.CreateInterface)
	interfaceGroup.POST("/:id/status", wgInterfaceController.UpdateInterfaceStatus)
	interfaceGroup.PATCH("/:id", wgInterfaceController.UpdateInterface)
	interfaceGroup.DELETE("/:id", wgInterfaceController.DeleteInterface)
	//interfaceGroup.GET("/wg-interface/:id", wgInterfaceController.GetWgInterfaceByID)
}

func setupPeerRoutes(router *echo.Group, mwpClients *common.MwpClients, jwtConfig echojwt.Config, wgPeerController *WgPeerController) {
	peerGroup := router.Group("/peer")
	peerGroup.Use(echojwt.WithConfig(jwtConfig))

	peerGroup.GET("/keys", wgPeerController.GetPeerKeys)
	peerGroup.GET("", wgPeerController.GetPeers, middleware.ClientConnectionMiddleware(mwpClients))
	peerGroup.POST("", wgPeerController.CreatePeer, middleware.ClientConnectionMiddleware(mwpClients))
	peerGroup.GET("/:id/share", wgPeerController.GetPeerShareStatus)
	peerGroup.PATCH("/:id/share/status", wgPeerController.UpdatePeerShareStatus)
	peerGroup.PUT("/:id/share/expire", wgPeerController.UpdatePeerShareExpire)
	peerGroup.PATCH("/:id/status", wgPeerController.UpdatePeerStatus, middleware.ClientConnectionMiddleware(mwpClients))
	peerGroup.PATCH("/:id/reset-usage", wgPeerController.ResetPeerUsage, middleware.ClientConnectionMiddleware(mwpClients))
	peerGroup.PUT("/:id", wgPeerController.UpdatePeer, middleware.ClientConnectionMiddleware(mwpClients))
	peerGroup.DELETE("/:id", wgPeerController.DeletePeer, middleware.ClientConnectionMiddleware(mwpClients))
	peerGroup.GET("/:id/config", wgPeerController.GetPeerConfig)
	peerGroup.GET("/:id/qrcode", wgPeerController.GetPeerQRCode)
}

func setupDeviceInfoRoutes(router *echo.Group, mwpClients *common.MwpClients, jwtConfig echojwt.Config, deviceInfoController *DeviceDataController) {
	deviceGroup := router.Group("/device")
	deviceGroup.Use(echojwt.WithConfig(jwtConfig), middleware.ClientConnectionMiddleware(mwpClients))

	deviceGroup.GET("/stats", deviceInfoController.GetDeviceInfo)
}

func setupSyncRoutes(router *echo.Group, mwpClients *common.MwpClients, jwtConfig echojwt.Config, syncController *SyncController) {
	syncGroup := router.Group("/sync")
	syncGroup.Use(echojwt.WithConfig(jwtConfig), middleware.ClientConnectionMiddleware(mwpClients))

	syncGroup.GET("/peers", syncController.SyncPeers)
	syncGroup.GET("/interfaces", syncController.SyncInterfaces)
}

func setupUserRoutes(router *echo.Group, userController *UserController) {
	userGroup := router.Group("/user")

	userGroup.GET("/:uuid/config", userController.GetUserConfig)
	userGroup.GET("/:uuid/qrcode", userController.GetUserQRCode)
	userGroup.GET("/:uuid/details", userController.GetUserDetails)
}
