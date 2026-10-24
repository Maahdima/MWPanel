package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/http/schema"
	"gorm.io/gorm"
)

// This availability middleware stays behind the existing JWT/account-role
// guards. It checks the actual target router instead of an unrelated default;
// it neither grants permissions nor changes tenant/ownership authorization.
func ClientConnectionMiddleware(clients *common.MwpClient) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			serverID, err := requestServerID(ctx, clients)
			if err != nil {
				var httpError *echo.HTTPError
				if errors.As(err, &httpError) {
					return httpError
				}
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return echo.NewHTTPError(http.StatusNotFound, "resource not found")
				}
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			connected := false
			if serverID != 0 {
				connected = clients.IsServerConnected(serverID)
			} else {
				connected = clients.AnyConnected()
			}
			if !connected {
				return ctx.JSON(http.StatusServiceUnavailable, schema.ErrorResponse{StatusCode: http.StatusServiceUnavailable, Status: "error", Message: "Client is not connected to the selected server"})
			}
			return next(ctx)
		}
	}
}

func requestServerID(ctx echo.Context, clients *common.MwpClient) (uint, error) {
	path := ctx.Path()
	if raw := ctx.Param("id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 32)
		if err != nil || id == 0 {
			return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid resource ID")
		}
		if strings.HasPrefix(path, "/api/peer/") {
			return clients.ResourceServerID("peer", uint(id))
		}
		if strings.HasPrefix(path, "/api/interface/") {
			return clients.ResourceServerID("interface", uint(id))
		}
	}
	if ctx.Request().Method == http.MethodPost && (path == "/api/peer" || path == "/api/interface") {
		const limit = 64 << 10
		data, err := io.ReadAll(io.LimitReader(ctx.Request().Body, limit+1))
		if err != nil {
			return 0, err
		}
		ctx.Request().Body = io.NopCloser(bytes.NewReader(data))
		if len(data) > limit {
			return 0, echo.NewHTTPError(http.StatusRequestEntityTooLarge, "request is too large")
		}
		var selection struct {
			ServerID    uint `json:"server_id"`
			InterfaceID uint `json:"interface_id"`
		}
		if err := json.Unmarshal(data, &selection); err != nil {
			return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid JSON")
		}
		if path == "/api/peer" {
			if selection.InterfaceID == 0 {
				return 0, echo.NewHTTPError(http.StatusBadRequest, "interface_id is required")
			}
			return clients.ResourceServerID("interface", selection.InterfaceID)
		}
		return selection.ServerID, nil
	}
	if raw := ctx.QueryParam("server_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 32)
		if err != nil || id == 0 {
			return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid server_id")
		}
		return uint(id), nil
	}
	return 0, nil
}
