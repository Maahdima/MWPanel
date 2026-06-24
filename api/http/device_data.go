package http

import (
	"net/http"

	"github.com/maahdima/mwp/api/cmd/jobs"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type DeviceDataController struct {
	deviceDataService *service.DeviceData
	trafficCalculator *traffic.Calculator
	logger            *zap.Logger
}

func NewDeviceDataController(deviceDataService *service.DeviceData, trafficCalculator *traffic.Calculator) *DeviceDataController {
	return &DeviceDataController{
		deviceDataService: deviceDataService,
		trafficCalculator: trafficCalculator,
		logger:            zap.L().Named("DeviceDataController"),
	}
}

func (c *DeviceDataController) GetDailyTrafficUsage(ctx echo.Context) error {
	rangeParam := ctx.QueryParam("range")
	if rangeParam == "" {
		c.logger.Error("failed to parse range query parameter")
		return ctx.JSON(http.StatusBadRequest, schema.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Status:     "error",
			Message:    "invalid range query parameter",
		})
	}

	trafficData, err := c.deviceDataService.GetDailyTrafficUsage(rangeParam)
	if err != nil {
		c.logger.Error("failed to fetch daily traffic usage", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to fetch daily traffic usage: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.DailyTrafficUsageResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *trafficData,
	})
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

func (c *DeviceDataController) ResetTotalTrafficUsage(ctx echo.Context) error {
	if err := c.trafficCalculator.ResetTotalTrafficUsage(); err != nil {
		c.logger.Error("failed to reset total traffic usage", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to reset total traffic usage: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}
