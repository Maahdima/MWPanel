package api

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"mikrotik-wg-go/api/schema"
	"mikrotik-wg-go/service"
	"net/http"
	"strconv"
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

func (c *WgInterfaceController) CreateInterface(ctx echo.Context) error {
	var req schema.CreateInterfaceRequest
	if err := ctx.Bind(&req); err != nil {
		c.logger.Error("failed to bind request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Status:     "error",
			Message:    "invalid request data: " + err.Error(),
		})
	}

	iface, err := c.interfaceService.CreateInterface(&req)
	if err != nil {
		c.logger.Error("failed to create wireguard interface", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to create wireguard interface: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusCreated, schema.BasicResponseData[schema.InterfaceResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *iface,
	})
}

func (c *WgInterfaceController) UpdateInterfaceStatus(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("interface ID is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	interfaceId, err := strconv.Atoi(id)
	if err != nil {
		c.logger.Error("Invalid interface ID", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	err = c.interfaceService.ToggleInterfaceStatus(uint(interfaceId))
	if err != nil {
		c.logger.Error("failed to update wireguard interface status", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to update wireguard interface status: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}

func (c *WgInterfaceController) UpdateInterface(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("interface ID is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	interfaceId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.logger.Error("Invalid interface ID", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	var req schema.UpdateInterfaceRequest
	if err := ctx.Bind(&req); err != nil {
		c.logger.Error("failed to bind request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := ctx.Validate(req); err != nil {
		c.logger.Warn("failed to validate request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	iface, err := c.interfaceService.UpdateInterface(uint(interfaceId), &req)
	if err != nil {
		c.logger.Error("failed to update wireguard interface", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to update wireguard interface: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.InterfaceResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *iface,
	})
}

func (c *WgInterfaceController) DeleteInterface(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("Interface ID is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	interfaceId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.logger.Error("Invalid interface ID", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	err = c.interfaceService.DeleteInterface(uint(interfaceId))
	if err != nil {
		c.logger.Error("failed to delete wireguard interface", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to delete wireguard interface: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusNoContent, schema.BasicResponse{
		StatusCode: http.StatusNoContent,
		Status:     "success",
	})
}
