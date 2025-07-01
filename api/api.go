package api

import (
	"github.com/labstack/echo/v4"
	"mikrotik-wg-go/service"
)

func SetupMwpAPI(app *echo.Echo, peerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator, deviceDataService *service.DeviceData) {
	router := app.Group("/api")

	//setupWgInterfaceRoutes(router, interfaceService)
	setupWgPeerRoutes(router, peerService, peerConfigService, peerQrCodeService)
	setupDeviceInfoRoutes(router, deviceDataService)
}

func setupWgInterfaceRoutes(router *echo.Group, interfaceService *service.WgInterface) {}

func setupWgPeerRoutes(router *echo.Group, peerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator) {
	wgPeerController := NewWgPeerController(peerService, peerConfigService, peerQrCodeService)

	router.GET("/wg-peer", wgPeerController.GetPeers)
	router.POST("/wg-peer", wgPeerController.CreatePeer)
	//router.PUT("/wg-peer/:id", wgPeerController.UpdatePeer)
	//router.DELETE("/wg-peer/:id", wgPeerController.DeletePeer)
	//router.GET("/wg-peer/:id", wgPeerController.GetPeerByID)
	router.GET("/wg-config/:id", wgPeerController.GetPeerConfig)
	router.GET("/wg-qrcode/:id", wgPeerController.GetPeerQRCode)
	router.GET("/peers-data", wgPeerController.GetPeersData)
}

func setupDeviceInfoRoutes(router *echo.Group, deviceDataService *service.DeviceData) {
	deviceInfoController := NewDeviceDataController(deviceDataService)

	router.GET("/device-info", deviceInfoController.GetDeviceInfo)
	//router.GET("/wg-interfaces", deviceInfoController.GetWgInterfaces)
}
