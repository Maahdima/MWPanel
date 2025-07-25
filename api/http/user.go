package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"
)

type UserController struct {
	peerService       *service.WgPeer
	peerConfigService *service.ConfigGenerator
	peerQrCodeService *service.QRCodeGenerator
	logger            *zap.Logger
}

func NewUserController(peerService *service.WgPeer, peerConfigService *service.ConfigGenerator, qrCodeService *service.QRCodeGenerator) *UserController {
	return &UserController{
		peerService:       peerService,
		peerConfigService: peerConfigService,
		peerQrCodeService: qrCodeService,
		logger:            zap.L().Named("UserController"),
	}
}

func (u *UserController) GetUserDetails(ctx echo.Context) error {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		u.logger.Error("Peer uuid is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	stats, err := u.peerService.GetPeerDetails(uuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, common.ErrPeerNotShared) {
			return ctx.JSON(http.StatusNotFound, schema.ErrorResponse{
				StatusCode: http.StatusNotFound,
				Status:     "error",
				Message:    "peer not found",
			})
		}

		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to retrieve peer stats: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.PeerDetailsResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *stats,
	})
}

func (u *UserController) GetUserConfig(ctx echo.Context) error {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		u.logger.Error("Peer uuid is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	config, err := u.peerConfigService.GetUserConfig(uuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, common.ErrPeerNotShared) {
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

func (u *UserController) GetUserQRCode(ctx echo.Context) error {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		u.logger.Error("Peer uuid is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	qrCode, err := u.peerQrCodeService.GetUserQRCode(uuid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, common.ErrPeerNotShared) {
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
