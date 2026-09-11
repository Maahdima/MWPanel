package http

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"
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

	admin, err := a.authService.Login(req.Username, req.Password)
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
		Data:          *admin,
	})
}

func (a *AuthController) Verify2FA(ctx echo.Context) error {
	var req schema.Verify2FARequest
	if err := ctx.Bind(&req); err != nil {
		a.logger.Error("failed to bind 2FA verify request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	if err := ctx.Validate(&req); err != nil {
		a.logger.Error("validation failed for 2FA verify request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	admin, err := a.authService.Verify2FA(req.TempToken, req.Code)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		a.logger.Error("failed to verify 2FA", zap.Error(err))
		return ctx.JSON(status, schema.ErrorResponse{
			StatusCode: status,
			Status:     "error",
			Message:    "failed to verify two-factor code: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.LoginResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *admin,
	})
}

func (a *AuthController) GetTotpStatus(ctx echo.Context) error {
	username, err := a.usernameFromJWT(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, schema.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Status:     "error",
			Message:    "unauthorized",
		})
	}

	status, err := a.authService.GetTotpStatus(username)
	if err != nil {
		a.logger.Error("failed to get TOTP status", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to get two-factor status: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.TotpStatusResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *status,
	})
}

func (a *AuthController) StartTotpSetup(ctx echo.Context) error {
	username, err := a.usernameFromJWT(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, schema.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Status:     "error",
			Message:    "unauthorized",
		})
	}

	setup, err := a.authService.StartTotpSetup(username)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrTotpAlreadyEnabled) {
			status = http.StatusConflict
		}
		a.logger.Error("failed to start TOTP setup", zap.Error(err))
		return ctx.JSON(status, schema.ErrorResponse{
			StatusCode: status,
			Status:     "error",
			Message:    "failed to start two-factor setup: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.TotpSetupResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *setup,
	})
}

func (a *AuthController) ConfirmTotpSetup(ctx echo.Context) error {
	username, err := a.usernameFromJWT(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, schema.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Status:     "error",
			Message:    "unauthorized",
		})
	}

	var req schema.TotpConfirmRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	result, err := a.authService.ConfirmTotpSetup(username, req.Code)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrInvalidTotpCode) {
			status = http.StatusUnauthorized
		}
		a.logger.Error("failed to confirm TOTP setup", zap.Error(err))
		return ctx.JSON(status, schema.ErrorResponse{
			StatusCode: status,
			Status:     "error",
			Message:    "failed to confirm two-factor setup: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.TotpConfirmResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *result,
	})
}

func (a *AuthController) DisableTotp(ctx echo.Context) error {
	username, err := a.usernameFromJWT(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, schema.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
			Status:     "error",
			Message:    "unauthorized",
		})
	}

	var req schema.TotpDisableRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := a.authService.DisableTotp(username, req.Password, req.Code); err != nil {
		status := http.StatusBadRequest
		switch {
		case errors.Is(err, service.ErrInvalidTotpCode), err.Error() == "password mismatch":
			status = http.StatusUnauthorized
		case errors.Is(err, service.ErrTotpNotEnabled):
			status = http.StatusConflict
		}
		a.logger.Error("failed to disable TOTP", zap.Error(err))
		return ctx.JSON(status, schema.ErrorResponse{
			StatusCode: status,
			Status:     "error",
			Message:    "failed to disable two-factor authentication: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
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

	err := a.authService.UpdateProfile(req.OldUsername, req.OldPassword, req.NewUsername, req.NewPassword)
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

func (a *AuthController) usernameFromJWT(ctx echo.Context) (string, error) {
	token, ok := ctx.Get("user").(*jwt.Token)
	if !ok || token == nil {
		return "", errors.New("missing token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	username, _ := claims["sub"].(string)
	if username == "" {
		return "", errors.New("missing subject")
	}
	return username, nil
}
