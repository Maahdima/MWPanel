package api

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"mikrotik-wg-go/api/schema"
	"mikrotik-wg-go/service"
	"net/http"
	"strconv"
)

type ServerController struct {
	serverService *service.Server
	logger        *zap.Logger
}

func NewServerController(serverService *service.Server) *ServerController {
	return &ServerController{
		serverService: serverService,
		logger:        zap.L().Named("ServerController"),
	}
}

func (c *ServerController) GetServers(ctx echo.Context) error {
	servers, err := c.serverService.GetServers()
	if err != nil {
		c.logger.Error("failed to get servers", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to retrieve servers: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.ServerResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *servers,
	})
}

func (c *ServerController) CreateServer(ctx echo.Context) error {
	var req schema.CreateServerRequest

	if err := ctx.Bind(&req); err != nil {
		c.logger.Error("failed to bind request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := ctx.Validate(&req); err != nil {
		c.logger.Warn("failed to validate request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	server, err := c.serverService.CreateServer(&req)
	if err != nil {
		c.logger.Error("failed to create server", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to create server: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.ServerResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *server,
	})
}

func (c *ServerController) UpdateServerStatus(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("Server ID is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	serverId, err := strconv.Atoi(id)
	if err != nil {
		c.logger.Error("Invalid server ID", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	server, err := c.serverService.ToggleServerStatus(uint(serverId))
	if err != nil {
		c.logger.Error("failed to update server status", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to update server status: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.ServerResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *server,
	})
}

func (c *ServerController) UpdateServer(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("Server ID is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	serverId, err := strconv.Atoi(id)
	if err != nil {
		c.logger.Error("Invalid server ID", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	var req schema.UpdateServerRequest
	if err := ctx.Bind(&req); err != nil {
		c.logger.Error("failed to bind request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := ctx.Validate(&req); err != nil {
		c.logger.Warn("failed to validate request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	server, err := c.serverService.UpdateServer(uint(serverId), &req)
	if err != nil {
		c.logger.Error("failed to update server", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to update server: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.ServerResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *server,
	})
}

func (c *ServerController) DeleteServer(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("Server ID is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	serverId, err := strconv.Atoi(id)
	if err != nil {
		c.logger.Error("Invalid server ID", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := c.serverService.DeleteServer(uint(serverId)); err != nil {
		c.logger.Error("failed to delete server", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to delete server: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}
