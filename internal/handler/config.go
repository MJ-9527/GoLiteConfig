package handler

import (
	"GoLiteConfig/internal/model"
	"GoLiteConfig/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	service *service.ConfigService
}

func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		service: configService,
	}
}

func (h *ConfigHandler) PublishConfig(c *gin.Context) {
	var req model.PublishConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Code:    40001,
			Message: "invalid request body",
			Data:    nil,
		})
		return
	}

	resp, err := h.service.Publish(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Code:    40001,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "success",
		Data:    resp,
	})

}

func (h *ConfigHandler) GetConfig(c *gin.Context) {
	app := c.Query("app")
	env := c.Query("env")

	resp, err := h.service.GetCurrent(c.Request.Context(), app, env)
	if err != nil {
		status := http.StatusBadGateway
		if err.Error() == "config not found" {
			status = http.StatusNotFound
		}

		c.JSON(status, model.APIResponse{
			Code:    40001,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "success",
		Data:    resp,
	})
}
