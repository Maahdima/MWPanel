package http

import (
	"errors"
	"net/http"
	"os"
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

	case errors.Is(err, gorm.ErrRecordNotFound), errors.Is(err, os.ErrNotExist):
		status, message = http.StatusNotFound, "Resource not found"

	case errors.Is(err, service.ErrResellerSettings), errors.Is(err, service.ErrInvalidBillingQuantity), errors.Is(err, service.ErrReceiptImage):
		status, message = http.StatusBadRequest, err.Error()

	case errors.Is(err, service.ErrResourceForbidden):
		status, message = http.StatusForbidden, err.Error()

	case errors.Is(err, service.ErrInsufficientCredit):
		status, message = http.StatusPaymentRequired, err.Error()

	case errors.Is(err, service.ErrRouterSyncPending):
		status, message = http.StatusServiceUnavailable, "Router synchronization is pending"

	case errors.Is(err, service.ErrMigrationConflict), errors.Is(err, service.ErrPeerRestricted), errors.Is(err, service.ErrBilledOperationRequired), errors.Is(err, service.ErrBillingOverdue), errors.Is(err, service.ErrIdempotencyConflict), errors.Is(err, service.ErrReceiptState), errors.Is(err, service.ErrResellerLimit), errors.Is(err, service.ErrRouterMigrationRequired), errors.Is(err, service.ErrPendingPeerOperation):
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

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.ResellerView]{BasicResponse: schema.OkBasicResponse, Data: rows})
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

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.ResellerView]{BasicResponse: schema.OkBasicResponse, Data: row})
}

func (c *ResellerController) Create(ctx echo.Context) error {
	var input schema.CreateReseller

	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	row, err := c.service.Create(resellerActor(ctx), input)

	if err != nil {
		return resellerError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, schema.BasicResponseData[schema.ResellerView]{BasicResponse: schema.OkBasicResponse, Data: row})
}

func (c *ResellerController) Update(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")

	if err != nil {
		return err
	}

	var input schema.ResellerSettings

	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	row, err := c.service.Update(resellerActor(ctx), id, input)

	if err != nil {
		return resellerError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.ResellerView]{BasicResponse: schema.OkBasicResponse, Data: row})
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
		ResellerID   *uint   `json:"reseller_id"`
		Reason       string  `json:"reason"`
		DurationDays *string `json:"configured_duration_days"`
	}

	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	var duration *int64

	if input.DurationDays != nil && *input.DurationDays != "" {
		value, err := strconv.ParseInt(*input.DurationDays, 10, 64)

		if err != nil {
			return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
		}

		duration = &value
	}

	if err := c.service.TransferPeerPlan(resellerActor(ctx), id, input.ResellerID, input.Reason, duration); err != nil {
		return resellerError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}

func (c *ResellerController) MyPeers(ctx echo.Context) error {
	rows, err := c.service.MyPeers(resellerActor(ctx))

	if err != nil {
		return resellerError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.PeerResponse]{BasicResponse: schema.OkBasicResponse, Data: *rows})
}

func (c *ResellerController) MyConfig(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")

	if err != nil {
		return err
	}

	filename, name, err := c.service.MyConfig(resellerActor(ctx), id)

	if err != nil {
		return resellerError(ctx, err)
	}

	ctx.Response().Header().Set("Cache-Control", "no-store")

	return ctx.Attachment(filename, name)
}

func (c *ResellerController) MyQRCode(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")

	if err != nil {
		return err
	}

	filename, err := c.service.MyQRCode(resellerActor(ctx), id)

	if err != nil {
		return resellerError(ctx, err)
	}

	ctx.Response().Header().Set("Cache-Control", "no-store")

	return ctx.File(filename)
}

func (c *ResellerController) MySessions(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")

	if err != nil {
		return err
	}

	rows, err := c.service.MySessions(resellerActor(ctx), id)

	if err != nil {
		return resellerError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.PeerSessionsResponse]{BasicResponse: schema.OkBasicResponse, Data: *rows})
}

func (c *ResellerController) MyExport(ctx echo.Context) error {
	filename, err := c.service.MyExport(resellerActor(ctx))

	if err != nil {
		return resellerError(ctx, err)
	}

	defer os.Remove(filename)

	return ctx.Attachment(filename, "traffic-report.xlsx")
}

func (c *ResellerController) Suspension(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")

	if err != nil {
		return err
	}

	var input struct {
		Suspended *bool  `json:"suspended"`
		Reason    string `json:"reason"`
	}

	if err := ctx.Bind(&input); err != nil || input.Suspended == nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := c.service.SetSuspension(resellerActor(ctx), id, *input.Suspended, input.Reason); err != nil {
		return resellerError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}

func (c *ResellerController) Purchase(ctx echo.Context) error {
	actor := resellerActor(ctx)

	resellerID, peerID, err := purchaseIDs(ctx, actor)

	if err != nil {
		return err
	}

	if ctx.Request().Method == http.MethodGet {
		rows, err := c.service.PeerPurchaseHistory(actor, resellerID, peerID)

		if err != nil {
			return resellerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]service.PeerPurchaseView]{BasicResponse: schema.OkBasicResponse, Data: rows})
	}

	var req service.PeerPurchaseRequest

	if err = ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	view, err := c.service.PurchasePeer(actor, resellerID, peerID, ctx.Request().Header.Get("Idempotency-Key"), req)

	if err != nil {
		return resellerError(ctx, err)
	}

	status := http.StatusAccepted

	if view.State == "applied" {
		status = http.StatusOK
	}

	response := schema.OkBasicResponse

	response.StatusCode = status

	return ctx.JSON(status, schema.BasicResponseData[service.PeerPurchaseView]{BasicResponse: response, Data: view})
}

func purchaseIDs(ctx echo.Context, actor service.Principal) (uint, uint, error) {
	if ctx.Param("peer_id") != "" {
		reseller, err := resellerID(ctx, "id")

		if err != nil {
			return 0, 0, err
		}

		peer, err := resellerID(ctx, "peer_id")

		return reseller, peer, err
	}

	if actor.ResellerID == nil {
		return 0, 0, echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	peer, err := resellerID(ctx, "id")

	return *actor.ResellerID, peer, err
}

func (c *ResellerController) CreatePeer(ctx echo.Context) error {
	actor := resellerActor(ctx)

	var id uint

	if ctx.Param("id") != "" {
		value, err := resellerID(ctx, "id")

		if err != nil {
			return err
		}

		id = value
	} else {
		if actor.ResellerID == nil {
			return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
		}

		id = *actor.ResellerID
	}

	var req service.ResellerPeerCreate

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	view, err := c.service.CreatePeer(actor, id, ctx.Request().Header.Get("Idempotency-Key"), req)

	if err != nil {
		return resellerError(ctx, err)
	}

	status := http.StatusAccepted

	if view.State == "applied" {
		status = http.StatusOK
	}

	response := schema.OkBasicResponse

	response.StatusCode = status

	return ctx.JSON(status, schema.BasicResponseData[service.PeerPurchaseView]{BasicResponse: response, Data: view})
}

func (c *ResellerController) MyResources(ctx echo.Context) error {
	rows, err := c.service.MyResources(resellerActor(ctx))

	if err != nil {
		return resellerError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.ResellerResourceView]{BasicResponse: schema.OkBasicResponse, Data: rows})
}

func (c *ResellerController) Mutate(ctx echo.Context) error {
	actor := resellerActor(ctx)

	resellerID, peerID, err := purchaseIDs(ctx, actor)

	if err != nil {
		return err
	}

	var req service.PeerMutationRequest

	if err = ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	view, err := c.service.MutatePeer(actor, resellerID, peerID, ctx.Request().Header.Get("Idempotency-Key"), req)

	if err != nil {
		return resellerError(ctx, err)
	}

	status := http.StatusAccepted

	if view.State == "applied" {
		status = http.StatusOK
	}

	response := schema.OkBasicResponse

	response.StatusCode = status

	return ctx.JSON(status, schema.BasicResponseData[service.PeerPurchaseView]{BasicResponse: response, Data: view})
}

func (c *ResellerController) Migration(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")

	if err != nil {
		return err
	}

	actor := resellerActor(ctx)

	if ctx.Request().Method == http.MethodGet {
		rows, err := c.service.MigrationHistory(actor, id)

		if err != nil {
			return resellerError(ctx, err)
		}

		return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]service.RouterMigrationView]{BasicResponse: schema.OkBasicResponse, Data: rows})
	}

	var req service.RouterMigrationRequest

	if err = ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	view, err := c.service.BeginMigration(actor, id, ctx.Request().Header.Get("Idempotency-Key"), req)

	if err != nil {
		return resellerError(ctx, err)
	}

	status := http.StatusAccepted

	if view.State == "complete" {
		status = http.StatusOK
	}

	response := schema.OkBasicResponse

	response.StatusCode = status

	return ctx.JSON(status, schema.BasicResponseData[service.RouterMigrationView]{BasicResponse: response, Data: view})
}

func (c *ResellerController) CancelMigration(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")

	if err != nil {
		return err
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err = ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err = c.service.CancelMigration(resellerActor(ctx), id, ctx.Param("migration_id"), req.Reason); err != nil {
		return resellerError(ctx, err)
	}

	response := schema.OkBasicResponse

	response.StatusCode = http.StatusAccepted

	return ctx.JSON(http.StatusAccepted, schema.BasicResponseData[string]{BasicResponse: response, Data: "cancellation queued"})
}
