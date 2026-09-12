package worker

import (
	"context"
	"log"

	"github.com/jairocodes/CloudDeploy.git/internal/queue"
	"github.com/jairocodes/CloudDeploy.git/internal/store"
)

type Worker struct {
	queue *queue.Queue
	store *store.Store
}

func New(q *queue.Queue, s *store.Store) *Worker {
	return &Worker{queue: q, store: s}
}

// Este Run es el loop principal: este espera bloado por jobs y los procesa en paralelo.
func (w *Worker) Run(ctx context.Context) {
	log.Println("worker started - waiting for jobs")
	for {
		job, err := w.queue.Dequeue(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("dequeue error: %v", err)
			continue
		}
		go w.proccessJob(ctx, *job)
	}
}

// processJob tiene la responsabilidad de simular el pipeline por el momento - solo mueve el status hacia
// adelante. El build y el deploy reales serán agregados pronto en proximas fases.
func (w *Worker) proccessJob(ctx context.Context, job queue.DeployJob) {
	log.Printf("processing deploy %s (sha: %s)", job.DeployID, job.CommitSHA[:7])

	if err := w.store.UpdateDeployStatus(ctx, job.DeployID, "building", ""); err != nil {
		log.Printf("update status building: %v", err)
		return
	}

	//Todo lo del (paso 6): reemplazar por el build real con Kaniko
	if err := w.store.UpdateDeployStatus(ctx, job.DeployID, "deploying", ""); err != nil {
		log.Printf("update status deploying: %v", err)
		return
	}

	//Todo lo del (paso 7): reemplazar por el deploy real con client-go.
	if err := w.store.UpdateDeployStatus(ctx, job.DeployID, "success", ""); err != nil {
		log.Printf("update status success: %v", err)
		return
	}

	log.Printf("deploy %s completed (simulado)", job.DeployID)
}
