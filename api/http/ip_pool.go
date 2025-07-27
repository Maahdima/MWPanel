package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"
)

type IPPoolController struct {
	ipPoolService *service.IPPool
	logger        *zap.Logger
}

func NewIPPoolController(ipPoolService *service.IPPool) *IPPoolController {
	return &IPPoolController{
		ipPoolService: ipPoolService,
		logger:        zap.L().Named("IPPoolController"),
	}
}

func (c *IPPoolController) GetIPPools(ctx echo.Context) error {
	pools, err := c.ipPoolService.GetIPPools()
	if err != nil {
		c.logger.Error("failed to get IP pools", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to get IP pools: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[[]schema.IPPoolResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *pools,
	})
}

func (c *IPPoolController) CreateIPPool(ctx echo.Context) error {
	var req schema.CreateIPPoolRequest

	if err := ctx.Bind(&req); err != nil {
		c.logger.Warn("failed to bind request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := ctx.Validate(&req); err != nil {
		c.logger.Warn("failed to validate request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	ipPool, err := c.ipPoolService.CreateIPPool(&req)
	if err != nil {
		c.logger.Error("failed to create IP pool", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to create IP pool: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusCreated, schema.BasicResponseData[schema.IPPoolResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *ipPool,
	})
}

func (c *IPPoolController) UpdateIPPool(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("IP Pool ID is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	poolId, err := strconv.Atoi(id)
	if err != nil {
		c.logger.Error("Invalid IP Pool ID", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	var req schema.UpdateIPPoolRequest
	if err := ctx.Bind(&req); err != nil {
		c.logger.Warn("failed to bind request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := ctx.Validate(&req); err != nil {
		c.logger.Warn("failed to validate request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	ipPool, err := c.ipPoolService.UpdateIPPool(uint(poolId), &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Warn("IP Pool not found", zap.Uint("id", uint(poolId)))
			return ctx.JSON(http.StatusNotFound, schema.ErrorResponse{
				StatusCode: http.StatusNotFound,
				Status:     "error",
				Message:    "IP Pool not found",
			})
		}

		c.logger.Error("failed to update IP pool", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to update IP pool: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, schema.BasicResponseData[schema.IPPoolResponse]{
		BasicResponse: schema.OkBasicResponse,
		Data:          *ipPool,
	})
}

func (c *IPPoolController) DeleteIPPool(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		c.logger.Error("IP Pool ID is required")
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	poolId, err := strconv.Atoi(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.logger.Warn("IP Pool not found", zap.String("id", id))
			return ctx.JSON(http.StatusNotFound, schema.ErrorResponse{
				StatusCode: http.StatusNotFound,
				Status:     "error",
				Message:    "IP Pool not found",
			})
		}
		
		c.logger.Error("Invalid IP Pool ID", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, schema.BadParamsErrorResponse)
	}

	if err := c.ipPoolService.DeleteIPPool(uint(poolId)); err != nil {
		c.logger.Error("failed to delete IP pool", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, schema.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Status:     "error",
			Message:    "failed to delete IP pool: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusNoContent, schema.BasicResponse{
		StatusCode: http.StatusNoContent,
		Status:     "success",
	})
}
