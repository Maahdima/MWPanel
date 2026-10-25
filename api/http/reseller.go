package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"
	"gorm.io/gorm"
)

type ResellerController struct{ service *service.Resellers }

func NewResellerController(resellers *service.Resellers) *ResellerController {
	return &ResellerController{service: resellers}
}

func resellerActor(ctx echo.Context) service.Principal {
	actor, _ := ctx.Get("principal").(service.Principal)
	return actor
}
func resellerID(ctx echo.Context, param string) (uint, error) {
	id, err := strconv.ParseUint(ctx.Param(param), 10, 32)
	if err != nil || id == 0 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid ID")
	}
	return uint(id), nil
}
func resellerError(ctx echo.Context, err error) error {
	status, message := http.StatusInternalServerError, "Unable to complete reseller operation"
	switch {
	case errors.Is(err, service.ErrRoleForbidden):
		status, message = http.StatusForbidden, "Forbidden"
	case errors.Is(err, gorm.ErrRecordNotFound):
		status, message = http.StatusNotFound, "Resource not found"
	case errors.Is(err, service.ErrResellerSettings):
		status, message = http.StatusBadRequest, err.Error()
	case errors.Is(err, service.ErrResourceForbidden):
		status, message = http.StatusForbidden, err.Error()
	case errors.Is(err, service.ErrResellerLimit), errors.Is(err, service.ErrRouterMigrationRequired), errors.Is(err, service.ErrPendingPeerOperation):
		status, message = http.StatusConflict, err.Error()
	case strings.Contains(strings.ToLower(err.Error()), "unique constraint"), strings.Contains(strings.ToLower(err.Error()), "duplicate key"):
		status, message = http.StatusConflict, "Username or resource assignment already exists"
	}
	return ctx.JSON(status, schema.ErrorResponse{StatusCode: status, Status: "error", Message: message})
}
func (c *ResellerController) List(ctx echo.Context) error {
	rows, err := c.service.List(resellerActor(ctx))
	if err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]service.ResellerView]{BasicResponse: schema.OkBasicResponse, Data: rows})
}
func (c *ResellerController) Get(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")
	if err != nil {
		return err
	}
	row, err := c.service.Get(resellerActor(ctx), id)
	if err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[service.ResellerView]{BasicResponse: schema.OkBasicResponse, Data: row})
}
func (c *ResellerController) Create(ctx echo.Context) error {
	var input service.CreateReseller
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	row, err := c.service.Create(resellerActor(ctx), input)
	if err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusCreated, schema.BasicResponseData[service.ResellerView]{BasicResponse: schema.OkBasicResponse, Data: row})
}
func (c *ResellerController) Update(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")
	if err != nil {
		return err
	}
	var input service.ResellerSettings
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	row, err := c.service.Update(resellerActor(ctx), id, input)
	if err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[service.ResellerView]{BasicResponse: schema.OkBasicResponse, Data: row})
}
func (c *ResellerController) Peers(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")
	if err != nil {
		return err
	}
	peers, err := c.service.PeerList(resellerActor(ctx), id)
	if err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.PeerResponse]{BasicResponse: schema.OkBasicResponse, Data: *peers})
}
func (c *ResellerController) Transfer(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")
	if err != nil {
		return err
	}
	var input struct {
		ResellerID *uint  `json:"reseller_id"`
		Reason     string `json:"reason"`
	}
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	if err := c.service.TransferPeer(resellerActor(ctx), id, input.ResellerID, input.Reason); err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}
