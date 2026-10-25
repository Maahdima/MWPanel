package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"
)

type BillingController struct {
	billing  *service.Billing
	payments *service.Payments
}

func NewBillingController(billing *service.Billing, payments *service.Payments) *BillingController {
	return &BillingController{billing, payments}
}
func billingTarget(ctx echo.Context) (uint, error) {
	actor := resellerActor(ctx)
	if ctx.Param("reseller_id") != "" {
		return resellerID(ctx, "reseller_id")
	}
	if actor.ResellerID == nil || *actor.ResellerID == 0 {
		return 0, echo.NewHTTPError(http.StatusForbidden, "A reseller account is required")
	}
	return *actor.ResellerID, nil
}
func (c *BillingController) Book(ctx echo.Context) error {
	id, err := billingTarget(ctx)
	if err != nil {
		return err
	}
	book, err := c.billing.Book(resellerActor(ctx), id)
	if err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[service.BillingBook]{BasicResponse: schema.OkBasicResponse, Data: book})
}
func (c *BillingController) Instructions(ctx echo.Context) error {
	input, err := c.billing.Instructions(resellerActor(ctx))
	if err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[service.PaymentInstructions]{BasicResponse: schema.OkBasicResponse, Data: input})
}
func (c *BillingController) UpdateInstructions(ctx echo.Context) error {
	var input service.PaymentInstructions
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	if err := c.billing.UpdateInstructions(resellerActor(ctx), input); err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}
func (c *BillingController) Receipts(ctx echo.Context) error {
	id, err := billingTarget(ctx)
	if err != nil {
		return err
	}
	rows, err := c.payments.History(resellerActor(ctx), id)
	if err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]service.ReceiptView]{BasicResponse: schema.OkBasicResponse, Data: rows})
}
func (c *BillingController) Submit(ctx echo.Context) error {
	id, err := billingTarget(ctx)
	if err != nil {
		return err
	}
	ctx.Request().Body = http.MaxBytesReader(ctx.Response(), ctx.Request().Body, service.MaxReceiptBytes+(1<<20))
	if err := ctx.Request().ParseMultipartForm(1 << 20); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or oversized receipt upload")
	}
	if ctx.Request().MultipartForm != nil {
		defer ctx.Request().MultipartForm.RemoveAll()
	}
	form := ctx.Request().MultipartForm
	if form == nil || len(form.File) != 1 || len(form.File["image"]) != 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "One receipt image is required")
	}
	image, err := form.File["image"][0].Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Unable to read receipt image")
	}
	defer image.Close()
	input := service.ReceiptInput{IdempotencyKey: ctx.Request().Header.Get("Idempotency-Key"), AmountRial: ctx.FormValue("amount_rial"), PaidAt: ctx.FormValue("paid_at"), TransactionReference: ctx.FormValue("transaction_reference")}
	receipt, err := c.payments.Submit(resellerActor(ctx), id, input, image)
	if err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusCreated, schema.BasicResponseData[service.ReceiptView]{BasicResponse: schema.OkBasicResponse, Data: receipt})
}
func (c *BillingController) Image(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")
	if err != nil {
		return err
	}
	file, mime, err := c.payments.OpenReceipt(resellerActor(ctx), id)
	if err != nil {
		return resellerError(ctx, err)
	}
	defer file.Close()
	ctx.Response().Header().Set("Cache-Control", "no-store")
	ctx.Response().Header().Set("X-Content-Type-Options", "nosniff")
	ctx.Response().Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=receipt-%d", id))
	return ctx.Stream(http.StatusOK, mime, file)
}
func (c *BillingController) Approve(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")
	if err != nil {
		return err
	}
	if err := c.payments.Approve(resellerActor(ctx), id); err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}
func (c *BillingController) Reject(ctx echo.Context) error {
	id, err := resellerID(ctx, "id")
	if err != nil {
		return err
	}
	var input struct {
		Reason string `json:"reason"`
	}
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	if err := c.payments.Reject(resellerActor(ctx), id, input.Reason); err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}
func (c *BillingController) Adjust(ctx echo.Context) error {
	id, err := billingTarget(ctx)
	if err != nil {
		return err
	}
	var input struct {
		DeltaRial string `json:"delta_rial"`
		Reason    string `json:"reason"`
	}
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	delta, err := strconv.ParseInt(input.DeltaRial, 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	if err := c.billing.Adjust(resellerActor(ctx), id, ctx.Request().Header.Get("Idempotency-Key"), delta, input.Reason); err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}

func (c *BillingController) ExtendGrace(ctx echo.Context) error {
	id, err := billingTarget(ctx)
	if err != nil {
		return err
	}
	var input struct {
		Until  int64  `json:"until"`
		Reason string `json:"reason"`
	}
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	if err := c.billing.ExtendGrace(resellerActor(ctx), id, input.Until, input.Reason); err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}
func (c *BillingController) Override(ctx echo.Context) error {
	id, err := billingTarget(ctx)
	if err != nil {
		return err
	}
	var input struct {
		Until  int64  `json:"until"`
		Reason string `json:"reason"`
	}
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}
	if err := c.billing.OverrideBilling(resellerActor(ctx), id, input.Until, input.Reason); err != nil {
		return resellerError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, schema.OkBasicResponse)
}
