package router

import (
	"GoLiteConfig/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/health", handler.Health)

	api := r.Group("/api")
	{
		api.GET("/health", handler.Health)
	}

	return r
}
