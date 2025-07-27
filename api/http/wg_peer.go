package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/maahdima/mwp/api/cmd/traffic-job"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WgPeerController struct {
	peerService       *service.WgPeer
	peerConfigService *service.ConfigGenerator
	peerQrCodeService *service.QRCodeGenerator
	trafficCalculator *traffic.Calculator
	logger            *zap.Logger
}

func NewWgPeerController(PeerService *service.WgPeer, peerConfigService *service.ConfigGenerator, peerQrCodeService *service.QRCodeGenerator, trafficCalculator *traffic.Calculator) *WgPeerController {
	return &WgPeerController{
		peerService:       PeerService,
		peerConfigService: peerConfigService,
		peerQrCodeService: peerQrCodeService,
		trafficCalculator: trafficCalculator,
		logger:            zap.L().Named("WgPeerController"),
	}
}

func (c *WgPeerController) GetPeerCredentials(ctx echo.Context) error {
	credentials, err := c.peerService.GetPeerCredentials()
	if err != nil {
		c.logger.Error("failed to get wireguard peer credentials", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to retrieve wireguard peer credentials: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.PeerCredentialsResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *credentials,
	})
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

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.PeerResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *peers,
	})
}

// TODO: implement
func (c *WgPeerController) GetPeerByID(ctx echo.Context) error {
	return nil
}

func (c *WgPeerController) CreatePeer(ctx echo.Context) error {
	var req schema.CreatePeerRequest

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
		c.logger.Error("failed to create wireguard peer", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to create wireguard peer: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusCreated, schema.BasicResponseData[schema.PeerResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *peer,
	})
}

func (c *WgPeerController) UpdatePeerStatus(ctx echo.Context) error {
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

	err = c.peerService.TogglePeerStatus(uint(peerId))
	if err != nil {
		c.logger.Error("failed to update wireguard peer status", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to update wireguard peer status: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}

func (c *WgPeerController) UpdatePeer(ctx echo.Context) error {
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

	var req schema.UpdatePeerRequest
	if err := ctx.Bind(&req); err != nil {
		c.logger.Warn("failed to bind request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := ctx.Validate(&req); err != nil {
		c.logger.Warn("failed to validate request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	peer, err := c.peerService.UpdatePeer(uint(peerId), &req)
	if err != nil {
		c.logger.Error("failed to update wireguard peer", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to update wireguard peer: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.PeerResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *peer,
	})
}

func (c *WgPeerController) DeletePeer(ctx echo.Context) error {
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

	err = c.peerService.DeletePeer(uint(peerId))
	if err != nil {
		c.logger.Error("failed to delete wireguard peer", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to delete wireguard peer: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusNoContent, schema.BasicResponse{
		StatusCode: http.StatusNoContent,
		Status:     "success",
	})
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

	config, err := c.peerConfigService.GetPeerConfig(uint(peerId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, schema.ErrorResponse{
				StatusCode: http.StatusNotFound,
				Status:     "error",
				Message:    "peer not found",
			})
		}

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

	qrCode, err := c.peerQrCodeService.GetPeerQRCode(uint(peerId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, schema.ErrorResponse{
				StatusCode: http.StatusNotFound,
				Status:     "error",
				Message:    "peer not found",
			})
		}

		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to retrieve peer QR code: " + err.Error(),
		})
	}

	return ctx.File(qrCode)
}

func (c *WgPeerController) GetPeerShareStatus(ctx echo.Context) error {
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

	status, err := c.peerService.GetPeerShareStatus(uint(peerId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, schema.ErrorResponse{
				StatusCode: http.StatusNotFound,
				Status:     "error",
				Message:    "peer not found",
			})
		}

		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to retrieve peer share status: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.PeerShareStatusResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *status,
	})
}

func (c *WgPeerController) UpdatePeerShareStatus(ctx echo.Context) error {
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

	err = c.peerService.TogglePeerShareStatus(uint(peerId))
	if err != nil {
		c.logger.Error("failed to toggle wireguard peer share status", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to toggle wireguard peer share status: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}

func (c *WgPeerController) UpdatePeerShareExpire(ctx echo.Context) error {
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

	var req schema.UpdatePeerShareExpireRequest
	if err := ctx.Bind(&req); err != nil {
		c.logger.Warn("failed to bind request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := ctx.Validate(&req); err != nil {
		c.logger.Warn("failed to validate request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	err = c.peerService.UpdatePeerShareExpireTime(uint(peerId), req.ExpireTime)
	if err != nil {
		c.logger.Error("failed to update peer share expire time", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to update peer share expire time: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}

func (c *WgPeerController) ResetPeerUsage(ctx echo.Context) error {
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

	err = c.trafficCalculator.ResetPeerUsage(uint(peerId))
	if err != nil {
		c.logger.Error("failed to reset wireguard peer usage", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to reset wireguard peer usage: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}
