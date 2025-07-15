package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/http/schema"
)

func ClientConnectionMiddleware(mwpClients *common.MwpClients) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// TODO : multiple server support
			//serverName := c.Request().Header.Get("X-Server-Name")
			//if serverName == "" {
			//	serverName = c.QueryParam("server")
			//}
			//serverName = strings.TrimSpace(serverName)

			if !mwpClients.IsConnected(nil) {
				return c.JSON(http.StatusServiceUnavailable, schema.ErrorResponse{
					StatusCode: http.StatusServiceUnavailable,
					Status:     "error",
					Message:    "Client is not connected to the server",
				})
			}

			return next(c)
		}
	}
}
