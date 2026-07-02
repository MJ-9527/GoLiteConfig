package router

import (
	"GoLiteConfig/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(configHandler *handler.ConfigHandler) *gin.Engine {
	r := gin.Default()

	// /health联通测试接口
	r.GET("/health", handler.Health)

	api := r.Group("/api")
	{
		api.POST("/config", configHandler.PublishConfig)
		api.GET("/config", configHandler.GetConfig)
	}
	return r
}
