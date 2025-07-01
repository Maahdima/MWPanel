package api

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"mikrotik-wg-go/api/schema"
	"mikrotik-wg-go/service"
	"net/http"
	"strconv"
)

type WgPeerController struct {
	peerService       *service.WgPeer
	peerConfigService *service.ConfigGenerator
	peerQrCodeService *service.QRCodeGenerator
	logger            *zap.Logger
}

func NewWgPeerController(PeerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator) *WgPeerController {
	return &WgPeerController{
		peerService:       PeerService,
		peerConfigService: peerConfigService,
		peerQrCodeService: peerQrCodeService,
		logger:            zap.L().Named("WgPeerController"),
	}
}

func (c *WgPeerController) GetPeers(ctx echo.Context) error {
	peers, err := c.peerService.GetPeers()
	if err != nil {
		c.logger.Error("failed to get wireguard peers", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to retrieve wireguard peers: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.WgPeerResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *peers,
	})
}

func (c *WgPeerController) GetPeersData(ctx echo.Context) error {
	recentOnlinePeers, totalPeers, onlinePeers, err := c.peerService.GetPeersData()
	if err != nil {
		c.logger.Error("failed to get recent online peers", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to retrieve recent online peers: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.PeersDataResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data: schema.PeersDataResponse{
			RecentOnlinePeers: recentOnlinePeers,
			TotalPeers:        totalPeers,
			OnlinePeers:       onlinePeers,
		},
	})
}

// TODO: implement
func (c *WgPeerController) GetPeerByID(ctx echo.Context) error {
	return nil
}

func (c *WgPeerController) CreatePeer(ctx echo.Context) error {
	var req schema.WgPeerRequest

	if err := ctx.Bind(&req); err != nil {
		c.logger.Warn("failed to bind request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := ctx.Validate(&req); err != nil {
		c.logger.Warn("failed to validate request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	peer, err := c.peerService.CreatePeer(&req)
	if err != nil {
		c.logger.Error("failed to create WireGuard peer", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to create wireguard peer: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusCreated, schema.BasicResponseData[schema.WgPeerResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *peer,
	})
}

// TODO: implement
func (c *WgPeerController) UpdatePeer(ctx echo.Context) error {
	return nil
}

// TODO: implement
func (c *WgPeerController) DeletePeer(ctx echo.Context) error {
	return nil
}

func (c *WgPeerController) GetPeerConfig(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("Peer ID is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	peerId, err := strconv.Atoi(id)
	if err != nil {
		c.logger.Error("Invalid peer ID", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	config, err := c.peerConfigService.GetPeerConfig(int64(peerId))
	if err != nil {
		c.logger.Error("failed to get peer config", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to retrieve peer config: " + err.Error(),
		})
	}

	return ctx.File(config)
}

func (c *WgPeerController) GetPeerQRCode(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("Peer ID is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	peerId, err := strconv.Atoi(id)
	if err != nil {
		c.logger.Error("Invalid peer ID", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	qrCode, err := c.peerQrCodeService.GetPeerQRCode(int64(peerId))
	if err != nil {
		c.logger.Error("failed to get peer QR code", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to retrieve peer QR code: " + err.Error(),
		})
	}

	return ctx.File(qrCode)
}
