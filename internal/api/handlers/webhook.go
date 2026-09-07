package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jairocodes/CloudDeploy.git/internal/queue"
)

// GithubPushPayload es quien extra solo los campos del payload de github que
// realmente necesitamos del evento "push"
type GithubPushPayload struct {
	After       string `json:"after"`
	Respository struct {
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
}

// Webhook recibe el push de github, valida su firma HMAC, y enconla el deploy,
func (h *Handler) Webhook(c *gin.Context) {
	projectID := c.Param("project_id")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad body"})
		return
	}

	project, err := h.store.GetProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	sig := c.GetHeader("X-Hub-Signature-256")
	if !validateHMAC(body, project.WebhookSecret, sig) {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "invalid signature"})
		return
	}

	if c.GetHeader("X-GitHub-Event") != "push" {
		c.JSON(http.StatusOK, gin.H{"msg": "ignored"})
		return
	}

	var payload GithubPushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "invalid payload"})
		return
	}

	imageTag := fmt.Sprintf("localhost:5000/%s:%s", project.Name, payload.After[:7])

	deployID, err := h.store.CreateDeploy(c.Request.Context(), projectID, payload.After, imageTag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = h.queue.Enqueue(c.Request.Context(), queue.DeployJob{
		DeployID:  deployID,
		ProjectID: projectID,
		CommitSHA: payload.After,
		RepoURL:   project.RepoURL,
		Namespace: project.Namespace,
		ImageTag:  imageTag,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"deploy_id": deployID, "sha": payload.After[:7]})
}

// validateHMAC recalcula la firma esperada yla compara en tiempo constante
// contra la que se envia por GitHub, hmac.Equal, nunca es ==, esto evita timing attacks
func validateHMAC(body []byte, secret, sig string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}
