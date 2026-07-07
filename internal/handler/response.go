package handler

import (
	"GoLiteConfig/internal/model"

	"github.com/gin-gonic/gin"
)

func respondSuccess(c *gin.Context, data any) {
	c.JSON(200, model.APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, model.APIResponse{
		Code:    status,
		Message: message,
		Data:    nil,
	})
}
