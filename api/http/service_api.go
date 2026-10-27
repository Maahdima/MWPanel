package http

import (
	"crypto/subtle"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"
	"github.com/maahdima/mwp/api/shopclient"
)

type ServiceController struct {
	peers  *service.WgPeer
	files  *service.ConfigGenerator
	qr     *service.QRCodeGenerator
	logger *zap.Logger
}

func NewServiceController(peers *service.WgPeer, files *service.ConfigGenerator, qr *service.QRCodeGenerator) *ServiceController {
	return &ServiceController{
		peers:  peers,
		files:  files,
		qr:     qr,
		logger: zap.L().Named("ServiceController"),
	}
}

type servicePeer struct {
	ID            uint    `json:"id"`
	UUID          string  `json:"uuid"`
	Name          string  `json:"name"`
	Disabled      bool    `json:"disabled"`
	DownloadUsage int64   `json:"download_usage"`
	UploadUsage   int64   `json:"upload_usage"`
	TrafficLimit  *int64  `json:"traffic_limit"`
	ExpireTime    *string `json:"expire_time"`
	Online        bool    `json:"online"`
}

type serviceInterface struct {
	ID      uint `json:"id"`
	HasPool bool `json:"has_pool"`
}

func (c *ServiceController) Interface(ctx echo.Context) error {
	id, _ := strconv.ParseUint(ctx.QueryParam("id"), 10, 64)
	iface, err := c.peers.ShopInterface(uint(id))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Status:     "error",
			Message:    err.Error(),
		})
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[serviceInterface]{
		BasicResponse: schema.OkBasicResponse,
		Data: serviceInterface{
			ID:      iface.ID,
			HasPool: iface.IPPool != nil,
		},
	})
}

func (c *ServiceController) CreatePeer(ctx echo.Context) error {
	var req struct {
		Name         string `json:"name"`
		Endpoint     string `json:"endpoint"`
		Comment      string `json:"comment"`
		InterfaceID  uint   `json:"interface_id"`
		TrafficGB    int    `json:"traffic_gb"`
		DurationDays int    `json:"duration_days"`
	}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	peer, err := c.peers.CreateShopPeer(req.Name, req.Endpoint, req.Comment, req.InterfaceID, req.TrafficGB, req.DurationDays)
	if err != nil {
		c.logger.Error("service peer create failed", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Status:     "error",
			Message:    err.Error(),
		})
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[servicePeer]{
		BasicResponse: schema.OkBasicResponse,
		Data:          c.snapshot(peer),
	})
}

func (c *ServiceController) ChargePeer(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	var req struct {
		AddBytes int64 `json:"add_bytes"`
		AddDays  int   `json:"add_days"`
	}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	peer, err := c.peers.ApplyCharge(uint(id), req.AddBytes, req.AddDays)
	if err != nil {
		c.logger.Error("service peer charge failed", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Status:     "error",
			Message:    err.Error(),
		})
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[servicePeer]{
		BasicResponse: schema.OkBasicResponse,
		Data:          c.snapshot(peer),
	})
}

func (c *ServiceController) BonusPeer(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	var req struct {
		AddBytes int64 `json:"add_bytes"`
	}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	peer, err := c.peers.ApplyBonus(uint(id), req.AddBytes)
	if err != nil {
		c.logger.Error("service peer bonus failed", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Status:     "error",
			Message:    err.Error(),
		})
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[servicePeer]{
		BasicResponse: schema.OkBasicResponse,
		Data:          c.snapshot(peer),
	})
}

func (c *ServiceController) PeerByID(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	peer, err := c.peers.FindPeerByID(uint(id))
	if err != nil {
		return servicePeerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[servicePeer]{
		BasicResponse: schema.OkBasicResponse,
		Data:          c.snapshot(peer),
	})
}

func (c *ServiceController) PeerByUUID(ctx echo.Context) error {
	peer, err := c.peers.FindPeerByUUID(ctx.Param("uuid"))
	if err != nil {
		return servicePeerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[servicePeer]{
		BasicResponse: schema.OkBasicResponse,
		Data:          c.snapshot(peer),
	})
}

func (c *ServiceController) NameTaken(ctx echo.Context) error {
	taken, err := c.peers.PeerNameTaken(ctx.QueryParam("name"))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, schema.InternalServerErrorResponse)
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[map[string]bool]{
		BasicResponse: schema.OkBasicResponse,
		Data:          map[string]bool{"taken": taken},
	})
}

func (c *ServiceController) Config(ctx echo.Context) error {
	path, err := c.files.GetConfigByUUID(ctx.Param("uuid"))
	if err != nil {
		return servicePeerError(ctx, err)
	}
	return ctx.File(path)
}

func (c *ServiceController) QR(ctx echo.Context) error {
	path, err := c.qr.GetQRCodeByUUID(ctx.Param("uuid"))
	if err != nil {
		return servicePeerError(ctx, err)
	}
	return ctx.File(path)
}

func (c *ServiceController) snapshot(peer model.Peer) servicePeer {
	limit := peer.TrafficLimit
	return servicePeer{
		ID:            peer.ID,
		UUID:          peer.UUID,
		Name:          peer.Name,
		Disabled:      peer.Disabled,
		DownloadUsage: peer.DownloadUsage,
		UploadUsage:   peer.UploadUsage,
		TrafficLimit:  limit,
		ExpireTime:    peer.ExpireTime,
		Online:        c.peers.PeerOnline(peer),
	}
}

func servicePeerError(ctx echo.Context, err error) error {
	status := http.StatusInternalServerError
	if err == gorm.ErrRecordNotFound || os.IsNotExist(err) {
		status = http.StatusNotFound
	}
	return ctx.JSON(status, schema.ErrorResponse{
		StatusCode: status,
		Status:     "error",
		Message:    err.Error(),
	})
}

func ServiceTokenMiddleware(token string) echo.MiddlewareFunc {
	want := strings.TrimSpace(token)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if want == "" {
				return c.JSON(http.StatusUnauthorized, schema.ErrorResponse{
					StatusCode: http.StatusUnauthorized,
					Status:     "error",
					Message:    "service token is not configured",
				})
			}
			got := strings.TrimSpace(strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer "))
			if len(got) != len(want) || subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
				return c.JSON(http.StatusUnauthorized, schema.ErrorResponse{
					StatusCode: http.StatusUnauthorized,
					Status:     "error",
					Message:    "unauthorized",
				})
			}
			return next(c)
		}
	}
}

type ShopProxy struct {
	client *shopclient.Client
	logger *zap.Logger
}

func NewShopProxy(client *shopclient.Client) *ShopProxy {
	return &ShopProxy{client: client, logger: zap.L().Named("ShopProxy")}
}

func (p *ShopProxy) ShopStatus(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.ShopStatus]{
		BasicResponse: schema.OkBasicResponse,
		Data:          schema.ShopStatus{Enabled: config.GetShopConfig().Enabled},
	})
}

func (p *ShopProxy) TelegramStatus(ctx echo.Context) error {
	if !config.GetShopConfig().Enabled {
		return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.TelegramStatusResponse]{
			BasicResponse: schema.OkBasicResponse,
			Data:          schema.TelegramStatusResponse{},
		})
	}

	return p.Forward(ctx)
}

func (p *ShopProxy) Forward(ctx echo.Context) error {
	if p.client == nil || !p.client.Enabled() {
		return ctx.JSON(http.StatusServiceUnavailable, schema.ErrorResponse{
			StatusCode: http.StatusServiceUnavailable,
			Status:     "error",
			Message:    "shop service is not configured",
		})
	}
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	resp, err := p.client.Forward(ctx.Request(), body)
	if err != nil {
		p.logger.Error("shop proxy failed", zap.Error(err))
		return ctx.JSON(http.StatusBadGateway, schema.ErrorResponse{
			StatusCode: http.StatusBadGateway,
			Status:     "error",
			Message:    "shop service is unavailable",
		})
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx.JSON(http.StatusBadGateway, schema.InternalServerErrorResponse)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	return ctx.Blob(resp.StatusCode, contentType, payload)
}
