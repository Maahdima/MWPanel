package http

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"mikrotik-wg-go/http/schema"
	"mikrotik-wg-go/service"
	"net/http"
)

type DeviceDataController struct {
	deviceDataService *service.DeviceData
	logger            *zap.Logger
}

func NewDeviceDataController(deviceDataService *service.DeviceData) *DeviceDataController {
	return &DeviceDataController{
		deviceDataService: deviceDataService,
		logger:            zap.L().Named("DeviceDataController"),
	}
}

func (c *DeviceDataController) GetDeviceInfo(ctx echo.Context) error {
	info, err := c.deviceDataService.GetDeviceData()
	if err != nil {
		c.logger.Error("failed to fetch device info", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to fetch device info: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.DeviceStatsResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *info,
	})
}
