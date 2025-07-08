package http

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"mikrotik-wg-go/http/schema"
	"mikrotik-wg-go/service"
	"net/http"
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
