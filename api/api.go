package api

import (
	"github.com/labstack/echo/v4"
	"mikrotik-wg-go/service"
)

func SetupMwpAPI(app *echo.Echo, peerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator, deviceDataService *service.DeviceData, interfaceService *service.WgInterface) {
	router := app.Group("/api")

	setupWgInterfaceRoutes(router, interfaceService)
	setupWgPeerRoutes(router, peerService, peerConfigService, peerQrCodeService)
	setupDeviceInfoRoutes(router, deviceDataService)
}

func setupWgInterfaceRoutes(router *echo.Group, interfaceService *service.WgInterface) {
	wgInterfaceController := NewWgInterfaceController(interfaceService)

	//router.GET("/wg-interfaces", wgInterfaceController.GetInterfaces)
	//router.POST("/wg-interfaces", wgInterfaceController.CreateInterface)
	//router.DELETE("/wg-interfaces/:name", wgInterfaceController.DeleteInterface)
	router.GET("/interfaces-data", wgInterfaceController.GetInterfacesData)
	//router.GET("/wg-interface/:id", wgInterfaceController.GetWgInterfaceByID)
	//router.PUT("/wg-interface/:id", wgInterfaceController.UpdateWgInterface)
}

func setupWgPeerRoutes(router *echo.Group, peerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator) {
	wgPeerController := NewWgPeerController(peerService, peerConfigService, peerQrCodeService)

	router.GET("/wg-peer", wgPeerController.GetPeers)
	router.POST("/wg-peer", wgPeerController.CreatePeer)
	router.PATCH("/wg-peer/:id", wgPeerController.UpdatePeer)
	//router.DELETE("/wg-peer/:id", wgPeerController.DeletePeer)
	//router.GET("/wg-peer/:id", wgPeerController.GetPeerByID)
	router.GET("/wg-config/:id", wgPeerController.GetPeerConfig)
	router.GET("/wg-qrcode/:id", wgPeerController.GetPeerQRCode)
	router.GET("/peers-data", wgPeerController.GetPeersData)
}

func setupDeviceInfoRoutes(router *echo.Group, deviceDataService *service.DeviceData) {
	deviceInfoController := NewDeviceDataController(deviceDataService)

	router.GET("/device-data", deviceInfoController.GetDeviceInfo)
}
