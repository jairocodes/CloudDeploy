package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jairocodes/CloudDeploy.git/internal/api/handlers"
)

// NewRouter registra todas las rutas de la API sobre un handler ya construido.
func NewRouter(h *handlers.Handler) *gin.Engine {
	r := gin.Default()

	r.POST("/projects", h.CreateProject)
	r.POST("/webhook/:project_id", h.Webhook)
	r.GET("/projects/:id/status", h.Status)

	return r
}
