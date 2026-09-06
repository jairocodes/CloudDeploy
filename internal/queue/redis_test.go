package queue

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestQueueSmoke(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	q := New(rdb)
	ctx := context.Background()

	job := DeployJob{
		DeployID:  "test-deploy-1",
		ProjectID: "test-project-1",
		CommitSHA: "abc123",
		RepoURL:   "https://github.com/x/demo",
		Namespace: "demo-ns",
		ImageTag:  "localhost:5000/demo:abc1234",
	}

	if err := q.Enqueue(ctx, job); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	dequeueCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	got, err := q.Dequeue(dequeueCtx)
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}

	if got.DeployID != job.DeployID {
		t.Fatalf("expected deploy id %s, got %s", job.DeployID, got.DeployID)
	}

	t.Logf("OK - dequeued job %+v", got)
}
