package controllers

import (
	"net/http"
	"order-service/src/lib"
	"order-service/src/models"
	"order-service/src/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CustomerController struct {
	service services.CustomerService
}

func NewCustomerController(service services.CustomerService) *CustomerController {
	return &CustomerController{service: service}
}

type CreateCustomerInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type UpdateCustomerInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

func (c *CustomerController) Create(ctx *gin.Context) {
	var input CreateCustomerInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	customer, err := c.service.CreateCustomer(models.Customer{
		Name:  input.Name,
		Email: input.Email,
	})
	if err != nil {
		lib.HandleServiceError(ctx, err)
		return
	}

	lib.RespondData(ctx, http.StatusCreated, customer)
}

func (c *CustomerController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	q := ctx.Query("q")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	customers, total, err := c.service.ListCustomers(page, limit, q)
	if err != nil {
		lib.HandleServiceError(ctx, err)
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	ctx.JSON(http.StatusOK, gin.H{
		"data": customers,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (c *CustomerController) Show(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "invalid_input", "invalid customer id")
		return
	}

	customer, err := c.service.GetCustomer(uint(id))
	if err != nil {
		lib.HandleServiceError(ctx, err)
		return
	}

	lib.RespondData(ctx, http.StatusOK, customer)
}

func (c *CustomerController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "invalid_input", "invalid customer id")
		return
	}

	var input UpdateCustomerInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	customer, err := c.service.UpdateCustomer(uint(id), models.Customer{
		Name:  input.Name,
		Email: input.Email,
	})
	if err != nil {
		lib.HandleServiceError(ctx, err)
		return
	}

	lib.RespondData(ctx, http.StatusOK, customer)
}
