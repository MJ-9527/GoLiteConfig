package router

import (
	"GoLiteConfig/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/health", handler.Health)
	}

	return r
}
