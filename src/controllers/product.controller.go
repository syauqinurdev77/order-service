package controllers

import (
	"net/http"
	"order-service/src/lib"
	"order-service/src/models"
	"order-service/src/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	service services.ProductService
}

func NewProductController(service services.ProductService) *ProductController {
	return &ProductController{service: service}
}

type CreateProductInput struct {
	Sku   string  `json:"sku" binding:"required"`
	Name  string  `json:"name" binding:"required"`
	Price float64 `json:"price" binding:"required"`
	Stock *int    `json:"stock" binding:"required"`
}

type UpdateProductInput struct {
	Name  string  `json:"name" binding:"required"`
	Price float64 `json:"price" binding:"required"`
	Stock *int    `json:"stock" binding:"required"`
}

func (c *ProductController) Create(ctx *gin.Context) {
	var input CreateProductInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	product, err := c.service.CreateProduct(models.Product{
		Sku:   input.Sku,
		Name:  input.Name,
		Price: input.Price,
		Stock: *input.Stock,
	})
	if err != nil {
		lib.HandleServiceError(ctx, err)
		return
	}

	lib.RespondData(ctx, http.StatusCreated, product)
}

func (c *ProductController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	q := ctx.Query("q")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	products, total, err := c.service.ListProducts(page, limit, q)
	if err != nil {
		lib.HandleServiceError(ctx, err)
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	ctx.JSON(http.StatusOK, gin.H{
		"data": products,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (c *ProductController) Show(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "invalid_input", "invalid product id")
		return
	}

	product, err := c.service.GetProduct(uint(id))
	if err != nil {
		lib.HandleServiceError(ctx, err)
		return
	}

	lib.RespondData(ctx, http.StatusOK, product)
}

func (c *ProductController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "invalid_input", "invalid product id")
		return
	}

	var input UpdateProductInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		lib.RespondError(ctx, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	product, err := c.service.UpdateProduct(uint(id), models.Product{
		Name:  input.Name,
		Price: input.Price,
		Stock: *input.Stock,
	})
	if err != nil {
		lib.HandleServiceError(ctx, err)
		return
	}

	lib.RespondData(ctx, http.StatusOK, product)
}
