package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"
)

type SyncController struct {
	syncService *service.SyncService
	logger      *zap.Logger
}

func NewSyncController(syncService *service.SyncService) *SyncController {
	return &SyncController{
		syncService: syncService,
		logger:      zap.L().Named("SyncController"),
	}
}

func (c *SyncController) SyncPeers(ctx echo.Context) error {
	if err := c.syncService.SyncPeers(); err != nil {
		c.logger.Error("failed to sync peers", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to sync peers: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}

func (c *SyncController) SyncInterfaces(ctx echo.Context) error {
	if err := c.syncService.SyncInterfaces(); err != nil {
		c.logger.Error("failed to sync interfaces", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to sync interfaces: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}
