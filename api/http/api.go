package http

import (
	"errors"
	"github.com/maahdima/mwp/api/dataservice/model"
	"net/http"

	traffic "github.com/maahdima/mwp/api/cmd/jobs"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/http/middleware"
	"github.com/maahdima/mwp/api/service"
	"github.com/maahdima/mwp/api/shopclient"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func SetupMwpAPI(
	app *echo.Echo,
	mwpClients *common.MwpClient,
	authenticationService *service.Authentication,
	serverService *service.Server,
	interfaceService *service.WgInterface,
	ipPoolService *service.IPPool,
	peerService *service.WgPeer,
	configGeneratorService *service.ConfigGenerator,
	qrCodeGeneratorService *service.QRCodeGenerator,
	excelGeneratorService *service.ExcelGenerator,
	deviceDataService *service.DeviceData,
	trafficCalculator *traffic.Calculator,
	syncService *service.SyncService,
	shopAPI *shopclient.Client,
	resellers *service.Resellers,
) {
	router := app.Group("/api")

	jwtConfig := echojwt.Config{
		SigningKey:    authenticationService.AccessSecret,
		SigningMethod: "HS256",
		ParseTokenFunc: func(c echo.Context, raw string) (interface{}, error) {
			token, principal, err := authenticationService.ParseAccessToken(raw)
			if err != nil {
				return nil, err
			}
			// Existing administrative endpoints are not tenant-scoped yet. Keep them
			// fail-closed for resellers until each dedicated route is implemented.
			if principal.Role != model.RoleSuperadmin {
				switch c.Path() {
				case "/api/auth/me", "/api/auth/profile", "/api/auth/totp", "/api/auth/totp/setup", "/api/auth/totp/confirm", "/api/auth/totp/disable":
				default:
					return nil, service.ErrRoleForbidden
				}
			}
			c.Set("principal", principal)
			return token, nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			if errors.Is(err, service.ErrRoleForbidden) {
				return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
			}
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		},
	}

	authController := NewAuthController(authenticationService)
	serverController := NewServerController(serverService)
	wgInterfaceController := NewWgInterfaceController(interfaceService)
	ipPoolController := NewIPPoolController(ipPoolService)
	wgPeerController := NewWgPeerController(
		peerService,
		configGeneratorService,
		qrCodeGeneratorService,
		excelGeneratorService,
		trafficCalculator,
	)
	deviceInfoController := NewDeviceDataController(deviceDataService, trafficCalculator)
	syncController := NewSyncController(syncService)
	userController := NewUserController(peerService, configGeneratorService, qrCodeGeneratorService)
	shopProxy := NewShopProxy(shopAPI)
	serviceController := NewServiceController(peerService, configGeneratorService, qrCodeGeneratorService)

	setupAuthenticationRoutes(router, jwtConfig, authController)
	resellerController := NewResellerController(resellers)
	resellerGroup := router.Group("/reseller", echojwt.WithConfig(jwtConfig))
	resellerGroup.GET("", resellerController.List)
	resellerGroup.POST("", resellerController.Create)
	resellerGroup.GET("/:id", resellerController.Get)
	resellerGroup.PUT("/:id", resellerController.Update)
	resellerGroup.GET("/:id/peers", resellerController.Peers)
	router.PUT("/peer/:id/owner", resellerController.Transfer, echojwt.WithConfig(jwtConfig))

	setupServerRoutes(router, jwtConfig, serverController)
	setupInterfaceRoutes(router, mwpClients, jwtConfig, wgInterfaceController)
	setupIPPoolRoutes(router, jwtConfig, ipPoolController)
	setupPeerRoutes(router, mwpClients, jwtConfig, wgPeerController)
	setupDeviceInfoRoutes(router, mwpClients, jwtConfig, deviceInfoController)
	setupSyncRoutes(router, mwpClients, jwtConfig, syncController)
	setupUserRoutes(router, userController)
	setupShopProxyRoutes(router, jwtConfig, shopProxy)
	setupServiceRoutes(router, serviceController, shopAPI)
}

func setupAuthenticationRoutes(router *echo.Group, jwtConfig echojwt.Config, authController *AuthController) {
	authGroup := router.Group("/auth")
	authGroup.POST("/login", authController.Login)
	authGroup.POST("/totp/verify", authController.Verify2FA)

	authProtected := authGroup.Group("")
	authProtected.Use(echojwt.WithConfig(jwtConfig))
	authProtected.GET("/me", authController.Me)
	authProtected.PUT("/profile", authController.UpdateProfile)
	authProtected.GET("/totp", authController.GetTotpStatus)
	authProtected.POST("/totp/setup", authController.StartTotpSetup)
	authProtected.POST("/totp/confirm", authController.ConfirmTotpSetup)
	authProtected.POST("/totp/disable", authController.DisableTotp)
}

func setupServerRoutes(router *echo.Group, jwtConfig echojwt.Config, serverController *ServerController) {
	serverGroup := router.Group("/server")
	serverGroup.Use(echojwt.WithConfig(jwtConfig))

	serverGroup.GET("", serverController.GetServers)
	serverGroup.POST("", serverController.CreateServer)
	serverGroup.PATCH("/:id/status", serverController.UpdateServerStatus)
	serverGroup.PUT("/:id", serverController.UpdateServer)
	serverGroup.DELETE("/:id", serverController.DeleteServer)
}

func setupInterfaceRoutes(router *echo.Group, mwpClients *common.MwpClient, jwtConfig echojwt.Config, wgInterfaceController *WgInterfaceController) {
	interfaceGroup := router.Group("/interface")
	interfaceGroup.Use(echojwt.WithConfig(jwtConfig))

	secured := interfaceGroup.Group("")
	secured.Use(middleware.ClientConnectionMiddleware(mwpClients))

	secured.GET("", wgInterfaceController.GetInterfaces)
	secured.POST("", wgInterfaceController.CreateInterface)
	secured.PATCH("/:id/status", wgInterfaceController.UpdateInterfaceStatus)
	secured.PUT("/:id", wgInterfaceController.UpdateInterface)
	secured.DELETE("/:id", wgInterfaceController.DeleteInterface)
}

func setupIPPoolRoutes(router *echo.Group, jwtConfig echojwt.Config, ipPpolController *IPPoolController) {
	ipPoolGroup := router.Group("/ip-pool")
	ipPoolGroup.Use(echojwt.WithConfig(jwtConfig))

	ipPoolGroup.GET("", ipPpolController.GetIPPools)
	ipPoolGroup.POST("", ipPpolController.CreateIPPool)
	ipPoolGroup.PUT("/:id", ipPpolController.UpdateIPPool)
	ipPoolGroup.DELETE("/:id", ipPpolController.DeleteIPPool)
}

func setupPeerRoutes(router *echo.Group, mwpClients *common.MwpClient, jwtConfig echojwt.Config, wgPeerController *WgPeerController) {
	peerGroup := router.Group("/peer")
	peerGroup.Use(echojwt.WithConfig(jwtConfig))

	peerGroup.POST("/allowed-address", wgPeerController.GetNewPeerAllowedAddress)
	peerGroup.GET("/credentials", wgPeerController.GetPeerCredentials)
	peerGroup.GET("/:id/share", wgPeerController.GetPeerShareStatus)
	peerGroup.PATCH("/:id/share/status", wgPeerController.UpdatePeerShareStatus)
	peerGroup.PATCH("/:id/share/expire", wgPeerController.UpdatePeerShareExpire)
	peerGroup.GET("/:id/config", wgPeerController.GetPeerConfig)
	peerGroup.GET("/:id/qrcode", wgPeerController.GetPeerQRCode)
	peerGroup.GET("/:id/sessions", wgPeerController.GetPeerSessions)

	peerSecured := peerGroup.Group("")
	peerSecured.Use(middleware.ClientConnectionMiddleware(mwpClients))

	peerSecured.GET("", wgPeerController.GetPeers)
	peerSecured.POST("", wgPeerController.CreatePeer)
	peerSecured.PATCH("/:id/status", wgPeerController.UpdatePeerStatus)
	peerSecured.PATCH("/:id/reset-usage", wgPeerController.ResetPeerUsage)
	peerSecured.PATCH("/reset-usage", wgPeerController.ResetPeerUsages)
	peerSecured.PUT("/:id", wgPeerController.UpdatePeer)
	peerSecured.DELETE("/:id", wgPeerController.DeletePeer)
	peerSecured.POST("/traffic/export", wgPeerController.ExportPeersTrafficData)
}

func setupDeviceInfoRoutes(router *echo.Group, mwpClients *common.MwpClient, jwtConfig echojwt.Config, deviceInfoController *DeviceDataController) {
	deviceGroup := router.Group("/device")
	deviceGroup.Use(echojwt.WithConfig(jwtConfig))

	deviceSecured := deviceGroup.Group("")
	deviceSecured.Use(middleware.ClientConnectionMiddleware(mwpClients))

	deviceSecured.GET("/stats", deviceInfoController.GetDeviceInfo)
	deviceSecured.GET("/traffic", deviceInfoController.GetDailyTrafficUsage)
	deviceSecured.PATCH("/traffic/reset", deviceInfoController.ResetTotalTrafficUsage)
}

func setupSyncRoutes(router *echo.Group, mwpClients *common.MwpClient, jwtConfig echojwt.Config, syncController *SyncController) {
	syncGroup := router.Group("/sync")
	syncGroup.Use(echojwt.WithConfig(jwtConfig))

	syncSecured := syncGroup.Group("")
	syncSecured.Use(middleware.ClientConnectionMiddleware(mwpClients))

	syncSecured.GET("/interfaces", syncController.GetSyncInterfaces)
	syncSecured.GET("/peers", syncController.GetSyncPeers)
	syncSecured.POST("/interfaces/selected", syncController.SyncSelectedInterfaces)
	syncSecured.POST("/peers/selected", syncController.SyncSelectedPeers)
	syncSecured.POST("/peers", syncController.SyncPeers)
	syncSecured.POST("/interfaces", syncController.SyncInterfaces)
}

func setupUserRoutes(router *echo.Group, userController *UserController) {
	userGroup := router.Group("/user")

	userGroup.GET("/:uuid/config", userController.GetUserConfig)
	userGroup.GET("/:uuid/qrcode", userController.GetUserQRCode)
	userGroup.GET("/:uuid/details", userController.GetUserDetails)
	userGroup.GET("/:uuid/telegram", userController.GetUserTelegram)
}

func setupShopProxyRoutes(router *echo.Group, jwtConfig echojwt.Config, proxy *ShopProxy) {
	router.GET("/telegram/status", proxy.TelegramStatus)
	router.POST("/bale/webhook", proxy.Forward)

	protected := router.Group("")
	protected.Use(echojwt.WithConfig(jwtConfig))
	protected.GET("/orders", proxy.Forward)
	protected.POST("/orders/:id/accept", proxy.Forward)
	protected.GET("/plans", proxy.Forward)
	protected.POST("/plans", proxy.Forward)
	protected.PUT("/plans/:id", proxy.Forward)
	protected.DELETE("/plans/:id", proxy.Forward)
	protected.GET("/shop/status", proxy.ShopStatus)
	protected.GET("/shop/sales", proxy.Forward)
	protected.PUT("/shop/sales", proxy.Forward)
}

func setupServiceRoutes(router *echo.Group, controller *ServiceController, shopAPI *shopclient.Client) {
	token := ""
	if shopAPI != nil {
		token = shopAPI.Token()
	}
	serviceGroup := router.Group("/service")
	serviceGroup.Use(ServiceTokenMiddleware(token))
	serviceGroup.GET("/interface", controller.Interface)
	serviceGroup.POST("/peers", controller.CreatePeer)
	serviceGroup.POST("/peers/:id/charge", controller.ChargePeer)
	serviceGroup.GET("/peers/name-taken", controller.NameTaken)
	serviceGroup.GET("/peers/id/:id", controller.PeerByID)
	serviceGroup.GET("/peers/:uuid", controller.PeerByUUID)
	serviceGroup.GET("/peers/:uuid/config", controller.Config)
	serviceGroup.GET("/peers/:uuid/qr", controller.QR)
}
