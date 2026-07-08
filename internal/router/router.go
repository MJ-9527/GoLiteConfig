package router

import (
	"GoLiteConfig/internal/handler"
	"GoLiteConfig/internal/logging"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRouter(configHandler *handler.ConfigHandler, logger *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors())
	r.Use(logging.RequestLogger(logger))

	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/console", "./web/index.html")
	r.GET("/health", handler.Health)

	api := r.Group("/api")
	{
		api.POST("/config", configHandler.PublishConfig)
		api.GET("/config", configHandler.GetConfig)
		api.GET("/config/diff", configHandler.DiffVersions)
		api.GET("/config/versions", configHandler.ListVersions)
		api.POST("/config/versions/delete", configHandler.DeleteVersions)
		api.POST("/config/rollback", configHandler.Rollback)
		api.GET("/watch", configHandler.Watch)
	}

	return r
}
