package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Health 健康检查
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "health",
		"version":   "0.0.1",
	})
}
