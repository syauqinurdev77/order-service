package routes

import (
	"order-service/src/config"
	"order-service/src/controllers"
	"order-service/src/repositories"
	"order-service/src/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	productRepo := repositories.NewProductRepository(config.DB)
	productService := services.NewProductService(productRepo)
	productController := controllers.NewProductController(productService)

	customerRepo := repositories.NewCustomerRepository(config.DB)
	customerService := services.NewCustomerService(customerRepo)
	customerController := controllers.NewCustomerController(customerService)

	orderRepo := repositories.NewOrderRepository(config.DB)
	orderService := services.NewOrderService(orderRepo)
	orderController := controllers.NewOrderController(orderService)

	registerProductRoutes(r.Group("/api"), productController)
	registerProductRoutes(r.Group("/"), productController)

	registerCustomerRoutes(r.Group("/api"), customerController)
	registerCustomerRoutes(r.Group("/"), customerController)

	registerOrderRoutes(r.Group("/api"), orderController)
	registerOrderRoutes(r.Group("/"), orderController)
}

func registerProductRoutes(group *gin.RouterGroup, controller *controllers.ProductController) {
	group.POST("/products", controller.Create)
	group.GET("/products", controller.List)
	group.GET("/products/:id", controller.Show)
	group.PUT("/products/:id", controller.Update)
}

func registerCustomerRoutes(group *gin.RouterGroup, controller *controllers.CustomerController) {
	group.POST("/customers", controller.Create)
	group.GET("/customers", controller.List)
	group.GET("/customers/:id", controller.Show)
	group.PUT("/customers/:id", controller.Update)
}

func registerOrderRoutes(group *gin.RouterGroup, controller *controllers.OrderController) {
	group.POST("/orders", controller.Create)
	group.GET("/orders", controller.List)
	group.GET("/orders/:id", controller.Show)
	group.PUT("/orders/:id/status", controller.UpdateStatus)
}
