package lib

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		log.Printf("[%s] %s %d %s", ctx.Request.Method, ctx.Request.URL.Path, ctx.Writer.Status(), time.Since(start))
	}
}
