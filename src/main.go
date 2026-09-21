package main

import (
	"fmt"
	"order-service/src/config"
	"order-service/src/lib"
	"order-service/src/routes"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(lib.LoggerMiddleware())

	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	fmt.Println()
	fmt.Printf("Server Running: http://localhost:%s\n", port)
	fmt.Println()

	r.Run(":" + port)
}
