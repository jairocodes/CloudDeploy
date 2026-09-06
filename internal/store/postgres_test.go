package store

import (
	"context"
	"testing"
)

func TestStoreSmoke(t *testing.T){
	db, err := Connect("postgres://cd:secret@localhost:5432/clouddeploy?sslmode=disable")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	projectID, err := s.CreateProject(ctx, "demo", "https://github.com/x/demo", "demo-ns", "main", "s3cr3t")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	deployID, err := s.CreateDeploy(ctx, projectID, "abc123")
	if err != nil {
		t.Fatalf("create deploy: %v", err)
	}

	if err := s.UpdateDeployStatus(ctx, deployID, "success", ""); err != nil {
		t.Fatalf("update status: %v", err)
	}

	t.Logf("OK - project %s, deploy %s", projectID, deployID)
}