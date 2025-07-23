package http

import (
	"net/http"

	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AuthController struct {
	authService *service.Authentication
	logger      *zap.Logger
}

func NewAuthController(authService *service.Authentication) *AuthController {
	return &AuthController{
		authService: authService,
		logger:      zap.L().Named("AuthController"),
	}
}

func (a *AuthController) Login(ctx echo.Context) error {
	var req schema.LoginRequest

	if err := ctx.Bind(&req); err != nil {
		a.logger.Error("failed to bind login request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := ctx.Validate(&req); err != nil {
		a.logger.Error("validation failed for login request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	accessToken, refreshToken, expiresIn, err := a.authService.Login(req.Username, req.Password)
	if err != nil {
		a.logger.Error("failed to login", zap.Error(err))
		return ctx.JSON(http.StatusNotFound, schema.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Status:     "error",
			Message:    "failed to login: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.LoginResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data: schema.LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    expiresIn,
		},
	})
}

func (a *AuthController) UpdateProfile(ctx echo.Context) error {
	var req schema.UpdateProfileRequest
	if err := ctx.Bind(&req); err != nil {
		a.logger.Warn("failed to bind request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := ctx.Validate(&req); err != nil {
		a.logger.Warn("failed to validate request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	err := a.authService.UpdateProfile(req.OldUsername, req.NewUsername, req.OldPassword, req.NewPassword)
	if err != nil {
		a.logger.Error("failed to update profile", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to update profile: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}
