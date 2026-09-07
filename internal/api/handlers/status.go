package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Status devuelve el historial de deploys de un proyecto. Por el momento es
// puramtne lo que sabemos de datos
func (h *Handler) Status(c *gin.Context) {
	projectID := c.Param("id")

	deploys, err := h.store.ListDeploys(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project_id": projectID, "deploys": deploys})
}
