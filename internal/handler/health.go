package handler

import (
	"net/http"

	"GoLiteConfig/internal/model"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"status": "ok",
		},
	})
}
