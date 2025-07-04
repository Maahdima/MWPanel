package api

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"mikrotik-wg-go/api/schema"
	"mikrotik-wg-go/service"
	"net/http"
)

type WgInterfaceController struct {
	interfaceService *service.WgInterface
	logger           *zap.Logger
}

func NewWgInterfaceController(interfaceService *service.WgInterface) *WgInterfaceController {
	return &WgInterfaceController{
		interfaceService: interfaceService,
		logger:           zap.L().Named("WgInterfaceController"),
	}
}

func (c *WgInterfaceController) GetInterfaces(ctx echo.Context) error {
	interfaces, err := c.interfaceService.GetInterfaces()
	if err != nil {
		c.logger.Error("failed to get wireguard interfaces", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to retrieve wireguard interfaces: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.InterfaceResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *interfaces,
	})
}

func (c *WgInterfaceController) GetInterfacesData(ctx echo.Context) error {
	data, err := c.interfaceService.GetInterfacesData()
	if err != nil {
		c.logger.Error("failed to get wireguard interfaces data", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to fetch interface info: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.InterfacesDataResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *data,
	})
}
