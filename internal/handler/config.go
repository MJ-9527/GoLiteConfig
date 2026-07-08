package handler

import (
	"GoLiteConfig/internal/model"
	"GoLiteConfig/internal/service"
	"net/http"
	"strconv"

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

// PublishConfig 推送Config响应结构
func (h *ConfigHandler) PublishConfig(c *gin.Context) {
	var req model.PublishConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Publish(c.Request.Context(), req)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	respondSuccess(c, resp)
}

// GetConfig 读取 Config 响应结构
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	app := c.Query("app")
	env := c.Query("env")
	group := c.Query("group")

	resp, err := h.service.GetCurrent(c.Request.Context(), app, env, group)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "config not found" {
			status = http.StatusNotFound
		}

		respondError(c, status, err.Error())
		return
	}

	respondSuccess(c, resp)
}

// ListVersions 列出历史版本
func (h *ConfigHandler) ListVersions(c *gin.Context) {
	app := c.Query("app")
	env := c.Query("env")

	resp, err := h.service.ListVersions(c.Request.Context(), app, env)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "config not found" {
			status = http.StatusNotFound
		}

		respondError(c, status, err.Error())
		return
	}

	respondSuccess(c, resp)
}

func (h *ConfigHandler) Rollback(c *gin.Context) {
	var req model.RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Rollback(c.Request.Context(), req)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "config not found" || err.Error() == "target version not found" {
			status = http.StatusNotFound
		}

		respondError(c, status, err.Error())
		return
	}

	respondSuccess(c, resp)
}

func (h *ConfigHandler) Watch(c *gin.Context) {
	app := c.Query("app")
	env := c.Query("env")
	lastRevisionStr := c.Query("last_revision")

	if lastRevisionStr == "" {
		respondError(c, http.StatusBadRequest, "last_revision is required")
		return
	}

	lastRevision, err := strconv.ParseInt(lastRevisionStr, 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "last_revision is invalid")
		return
	}

	resp, changed, err := h.service.Watch(c.Request.Context(), app, env, lastRevision)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "config not found" {
			status = http.StatusNotFound
		}

		respondError(c, status, err.Error())
		return
	}

	if !changed {
		c.Status(http.StatusNotModified)
		return
	}

	respondSuccess(c, resp)
}

func (h *ConfigHandler) DeleteVersions(c *gin.Context) {
	var req model.DeleteVersionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.DeleteVersions(c.Request.Context(), req)
	if err != nil {
		status := http.StatusBadRequest
		switch err.Error() {
		case "config not found", "target version not found":
			status = http.StatusNotFound
		case "cannot delete current version":
			status = http.StatusConflict
		}

		respondError(c, status, err.Error())
		return
	}

	respondSuccess(c, resp)
}

func (h *ConfigHandler) DiffVersions(c *gin.Context) {
	app := c.Query("app")
	env := c.Query("env")
	fromVersion := c.Query("from_version")
	toVersion := c.Query("to_version")

	resp, err := h.service.DiffVersions(c.Request.Context(), app, env, fromVersion, toVersion)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "target version not found" {
			status = http.StatusNotFound
		}
		respondError(c, status, err.Error())
		return
	}

	respondSuccess(c, resp)
}
