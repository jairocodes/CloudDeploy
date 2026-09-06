package queue

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// QueueKey es el nombre que se le da a la lista de Redis donde vivirán los jobs pendientes
const QueueKey = "clouddeploy:jobs"

// DeployJob es el mensaje que viajara por la cola, el API server lo arma y luego
// al recibir un webhook, el worker lo consume para ejecutar el pipeline
type DeployJob struct {
	DeployID  string `json:"deploy_id"`
	ProjectID string `json:"project_id"`
	CommitSHA string `json:"commit_sha"`
	RepoURL   string `json:"repo_url"`
	Namespace string `json:"namespace"`
	ImageTag  string `json:"image_tag"`
}

type Queue struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Queue {
	return &Queue{rdb: rdb}
}

// Enqueue se encarga de serializar el job a JSON y lo empuja al frente de lista con LPUSH
func (q *Queue) Enqueue(ctx context.Context, job DeployJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.rdb.LPush(ctx, QueueKey, data).Err()
}

// Dequeue usa BRPop que es bloqueante, es decir: el worker espera ahí, sin consumir nada de CPU,
// todo hasta que exista un job en la cola
func (q *Queue) Dequeue(ctx context.Context) (*DeployJob, error) {
	result, err := q.rdb.BRPop(ctx, 0, QueueKey).Result()
	if err != nil {
		return nil, err
	}
	var job DeployJob
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, err
	}
	return &job, nil
}
