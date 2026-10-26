package http

import (
	"net/http"
	"strconv"
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
	syncer, err := c.forRequest(ctx)

	if err != nil {
		return err
	}

	ifaces, err := syncer.GetSyncInterfaces()

	if err != nil {
		c.logger.Error("failed to fetch sync interfaces", zap.Error(err))

		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{StatusCode: http.StatusInternalServerError, Status: "error", Message: "failed to fetch interfaces for sync: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.SyncInterfacePreviewResponse]{BasicResponse: schema.OkBasicResponse, Data: ifaces})
}

func (c *SyncController) GetSyncPeers(ctx echo.Context) error {
	syncer, err := c.forRequest(ctx)

	if err != nil {
		return err
	}

	interfaceName := strings.TrimSpace(ctx.QueryParam("interface"))

	peers, err := syncer.GetSyncPeers(interfaceName)

	if err != nil {
		c.logger.Error("failed to fetch sync peers", zap.Error(err))

		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{StatusCode: http.StatusInternalServerError, Status: "error", Message: "failed to fetch peers for sync: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.SyncPeerPreviewResponse]{BasicResponse: schema.OkBasicResponse, Data: peers})
}

func (c *SyncController) SyncSelectedInterfaces(ctx echo.Context) error {
	syncer, err := c.forRequest(ctx)

	if err != nil {
		return err
	}

	var req schema.SyncInterfacesRequest

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := syncer.SyncSelectedInterfaces(req.InterfaceIDs); err != nil {
		c.logger.Error("failed to sync selected interfaces", zap.Error(err))

		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{StatusCode: http.StatusInternalServerError, Status: "error", Message: "failed to sync selected interfaces: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}

func (c *SyncController) SyncSelectedPeers(ctx echo.Context) error {
	syncer, err := c.forRequest(ctx)

	if err != nil {
		return err
	}

	var req schema.SyncPeersRequest

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := syncer.SyncSelectedPeers(req.PeerIDs); err != nil {
		c.logger.Error("failed to sync selected peers", zap.Error(err))

		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{StatusCode: http.StatusInternalServerError, Status: "error", Message: "failed to sync selected peers: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}

func (c *SyncController) SyncPeers(ctx echo.Context) error {
	syncer, err := c.forRequest(ctx)

	if err != nil {
		return err
	}

	if err := syncer.SyncPeers(); err != nil {
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
	syncer, err := c.forRequest(ctx)

	if err != nil {
		return err
	}

	if err := syncer.SyncInterfaces(); err != nil {
		c.logger.Error("failed to sync interfaces", zap.Error(err))

		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to sync interfaces: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}

func (c *SyncController) forRequest(ctx echo.Context) (*service.SyncService, error) {
	raw := ctx.QueryParam("server_id")

	if raw == "" {
		return c.syncService, nil
	}

	id, err := strconv.ParseUint(raw, 10, 32)

	if err != nil || id == 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid server_id")
	}

	selected, err := c.syncService.ForServer(uint(id))

	if err != nil {
		return nil, echo.NewHTTPError(http.StatusServiceUnavailable, "selected server is unavailable")
	}

	return selected, nil
}
