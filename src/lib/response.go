package lib

import "github.com/gin-gonic/gin"

func RespondError(ctx *gin.Context, status int, code string, detail any) {
	ctx.JSON(status, gin.H{
		"error":  code,
		"detail": detail,
	})
}

func RespondData(ctx *gin.Context, status int, data any) {
	ctx.JSON(status, gin.H{
		"data": data,
	})
}
