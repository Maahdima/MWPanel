package middleware

import (
	"errors"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/service"
	"net/http"
)

// JWTConfig extends the existing JWT middleware with current database roles.
func JWTConfig(authenticationService *service.Authentication) echojwt.Config {
	return echojwt.Config{
		SigningKey:    authenticationService.AccessSecret,
		SigningMethod: "HS256",
		ParseTokenFunc: func(c echo.Context, raw string) (interface{}, error) {
			token, principal, err := authenticationService.ParseAccessToken(raw)

			if err != nil {
				return nil, err
			}

			// Legacy administrative routes remain admin-only. Resellers use the
			// explicit scoped routes below, with ownership enforced by services.
			if principal.Role != model.RoleSuperadmin {
				switch c.Path() {
				case "/api/reseller/me/peers/:id/mutations", "/api/reseller/me/resources", "/api/reseller/me/peers/:id/purchases", "/api/reseller/me/peers", "/api/reseller/me/peers/:id/config", "/api/reseller/me/peers/:id/qrcode", "/api/reseller/me/peers/:id/sessions", "/api/reseller/me/peers/export":

				case "/api/billing/me/:kind/export", "/api/billing/me", "/api/billing/me/receipts", "/api/billing/me/receipts/:id/image":

				case "/api/billing/instructions":
					if c.Request().Method != http.MethodGet {
						return nil, service.ErrRoleForbidden
					}

				case "/api/auth/me", "/api/auth/profile", "/api/auth/totp", "/api/auth/totp/setup", "/api/auth/totp/confirm", "/api/auth/totp/disable":

				default:
					return nil, service.ErrRoleForbidden
				}
			}

			c.Set("principal", principal)

			return token, nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			if errors.Is(err, service.ErrRoleForbidden) {
				return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
			}

			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		},
	}

}
