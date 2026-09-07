package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateProjectRequest es quien define la forma esperada del JSON de entrada
// Los tags `binding:"required"` hacen que Gin rechace la peticion con un 400
// automaticamente si falta alguno de esos campos.
type createProjectRequest struct {
	Name          string `json:"name" binding:"required"`
	RepoURL       string `json:"repo_url" binding:"required"`
	Namespace     string `json:"namespace" binding:"required"`
	Branch        string `json:"branch"`
	WebhookSecret string `json:"webhook_secret" binding:"required"`
}

// CreateProject es quien registra un proyecto nuevo, es el paso previo indispensable
// para poder recibir webhooks de él (necesita namespace y webhook_secret)
func (h *Handler) CreateProject(c *gin.Context) {
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branch := req.Branch
	if branch == "" {
		branch = "main"
	}

	id, err := h.store.CreateProject(c.Request.Context(), req.Name, req.RepoURL, req.Namespace, branch, req.WebhookSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}
