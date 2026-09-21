package controllers

import (
	"net/http"
	"order-service/src/lib"
	"order-service/src/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	service services.OrderService
}

func NewOrderController(service services.OrderService) *OrderController {
	return &OrderController{service: service}
}

type CreateOrderItemInput struct {
	ProductID uint `json:"product_id" binding:"required"`
	Qty       int  `json:"qty" binding:"required"`
}

type CreateOrderInput struct {
	CustomerID uint                   `json:"customer_id" binding:"required"`
	Items      []CreateOrderItemInput `json:"items" binding:"required,min=1,dive"`
}

type UpdateOrderStatusInput struct {
	Status string `json:"status" binding:"required"`
}

func (c *OrderController) Create(ctx *gin.Context) {
	var input CreateOrderInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	items := make([]services.OrderItemInput, len(input.Items))
	for i, item := range input.Items {
		items[i] = services.OrderItemInput{
			ProductID: item.ProductID,
			Qty:       item.Qty,
		}
	}

	order, err := c.service.CreateOrder(services.CreateOrderInput{
		CustomerID: input.CustomerID,
		Items:      items,
	})
	if err != nil {
		lib.HandleServiceError(ctx, err)
		return
	}

	lib.RespondData(ctx, http.StatusCreated, order)
}

func (c *OrderController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	var customerID *uint
	if rawID := ctx.Query("customer_id"); rawID != "" {
		if parsed, err := strconv.ParseUint(rawID, 10, 64); err == nil {
			id := uint(parsed)
			customerID = &id
		}
	}

	orders, total, err := c.service.ListOrders(page, limit, customerID)
	if err != nil {
		lib.HandleServiceError(ctx, err)
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	ctx.JSON(http.StatusOK, gin.H{
		"data": orders,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (c *OrderController) Show(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "invalid_input", "invalid order id")
		return
	}

	order, err := c.service.GetOrder(uint(id))
	if err != nil {
		lib.HandleServiceError(ctx, err)
		return
	}

	lib.RespondData(ctx, http.StatusOK, order)
}

func (c *OrderController) UpdateStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "invalid_input", "invalid order id")
		return
	}

	var input UpdateOrderStatusInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	order, err := c.service.UpdateStatus(uint(id), input.Status)
	if err != nil {
		lib.HandleServiceError(ctx, err)
		return
	}

	lib.RespondData(ctx, http.StatusOK, order)
}
