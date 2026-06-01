package http

import (
	"net/http"
	"strings"

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

func (c *SyncController) GetSyncInterfaces(ctx echo.Context) error {
	ifaces, err := c.syncService.GetSyncInterfaces()
	if err != nil {
		c.logger.Error("failed to fetch sync interfaces", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{StatusCode: http.StatusInternalServerError, Status: "error", Message: "failed to fetch interfaces for sync: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.SyncInterfacePreviewResponse]{BasicResponse: schema.OkBasicResponse, Data: ifaces})
}

func (c *SyncController) GetSyncPeers(ctx echo.Context) error {
	interfaceName := strings.TrimSpace(ctx.QueryParam("interface"))
	peers, err := c.syncService.GetSyncPeers(interfaceName)
	if err != nil {
		c.logger.Error("failed to fetch sync peers", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{StatusCode: http.StatusInternalServerError, Status: "error", Message: "failed to fetch peers for sync: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.SyncPeerPreviewResponse]{BasicResponse: schema.OkBasicResponse, Data: peers})
}

func (c *SyncController) SyncSelectedInterfaces(ctx echo.Context) error {
	var req schema.SyncInterfacesRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	if err := c.syncService.SyncSelectedInterfaces(req.InterfaceIDs); err != nil {
		c.logger.Error("failed to sync selected interfaces", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{StatusCode: http.StatusInternalServerError, Status: "error", Message: "failed to sync selected interfaces: " + err.Error()})
	}
	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}

func (c *SyncController) SyncSelectedPeers(ctx echo.Context) error {
	var req schema.SyncPeersRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	if err := c.syncService.SyncSelectedPeers(req.PeerIDs); err != nil {
		c.logger.Error("failed to sync selected peers", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{StatusCode: http.StatusInternalServerError, Status: "error", Message: "failed to sync selected peers: " + err.Error()})
	}
	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
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
