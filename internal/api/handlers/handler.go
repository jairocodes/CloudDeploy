package handlers

import (
	"github.com/jairocodes/CloudDeploy.git/internal/queue"
	"github.com/jairocodes/CloudDeploy.git/internal/store"
)

// Handler agrupa las dependecias que todos los endpoints necesitan:
// acceso a la base de datos(store) y a la cola de jobs (queue)
type Handler struct {
	store *store.Store
	queue *queue.Queue
}

func New(s *store.Store, q *queue.Queue) *Handler {
	return &Handler{store: s, queue: q}
}
