package store

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

//Schema SQL, se ejecuta una vez al conectar, es idempotente (IF NOT EXISTS)

const schema = `
CREATE TABLE IF NOT EXISTS projects (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name TEXT NOT NULL UNIQUE,
	repo_url TEXT NOT NULL,
	namespace TEXT NOT NULL,
	branch TEXT DEFAULT 'main',
	webhook_secret TEXT NOT NULL,
	created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS deploys(
	id	UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	project_id	UUID REFERENCES projects(id),
	commit_sha	TEXT NOT NULL,
	image_tag	TEXT,
	status	TEXT DEFAULT 'pending', -- pending|building|deploying|succes|failed
	logs	TEXT,
	created_at	TIMESTAMPTZ DEFAULT NOW(),
	finished_at	TIMESTAMPTZ
	);
`
//ErrNotFound es devuelto cuando una consulta esperaba una fila y no pudo encontrar ninguna.
var ErrNotFound = errors.New("store: not found")
type Project struct {
	ID	string
	Name	string
	RepoURL	string
	Namespace	string
	Branch	string
	WebhookSecret	string
}

type Deploy struct {
	ID	string
	ProjectID	string
	CommitSHA	string
	ImageTag	string
	Status	string
}

type Store struct{
	db *sql.DB
}

//La función Connect abre la conexión a PostgreSQL y aplica el schema, solo se llama una vez.
//Cuando arranca cmd/server o cmd/worker.
func Connect(dsn string) (*sql.DB, error){
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	return db, nil
}

func New(db *sql.DB) *Store{
	return &Store{db: db}
}

//La función CreateProject registra un proyecto nuevo, luego devuelve su ID generado
func (s *Store) CreateProject(ctx context.Context, name, repoURL, namespace, branch, webhookSecret string) (string, error){
	var id string
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO projects (name, repo_url, namespace, branch, webhook_secret)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		name, repoURL, namespace, branch, webhookSecret).Scan(&id)
	return id, err
}

//La función GetProject busca un proyecto por su ID, luego lo usa el handler de webhook
//para recuperar el namespace y el webhook_secret de validar la firma
func (s *Store) GetProject(ctx context.Context, id string) (*Project, error){
	var p Project
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, repo_url, namespace, branch, webhook_secret
		 FROM projects WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.RepoURL, &p.Namespace, &p.Branch, &p.WebhookSecret)
	if errors.Is(err, sql.ErrNoRows){
		return nil, ErrNotFound
	}
	return &p, err
}

//La función CreateDeploy registra un nuevo intento de deploy para un proyecto, en estado "pending"
func (s *Store) CreateDeploy(ctx context.Context, projectID, sha string) (string, error){
	var id string
	err := s.db.QueryRowContext(ctx,
		"INSERT INTO deploys (project_id,commit_sha) VALUES ($1,$2) RETURNING id",
		projectID, sha).Scan(&id)
	return id, err
}

//la función UpdateDeployStatus lo llama el worker en cada transición del pipeline
// (building a deploying a success/failed)
func (s *Store) UpdateDeployStatus(ctx context.Context, id, status, logs string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE deploys SET status=$1, logs=$2, finished_at=NOW() WHERE id=$3",
		status, logs, id)
	return err
}

//La función GetPreviousSuccessfulDeploy aunque con un nombre largo, busca el deploy ANTERIOR que fue exitoso al más recientemente
//registrado para el proyecto, este es el que usa el rollback para saber a qué image_tag volver
func (s *Store) GetPreviousSuccessfulDeploy(ctx context.Context, projectID string) (*Deploy, error){
	var d Deploy
	err := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, commit_sha, image_tag, status
		FROM deploys
		WHERE project_id = $1
			AND status = 'success'
			AND id != (SELECT id FROM deploys WHERE project_id = $1 ORDER BY created_at DESC LIMIT 1)
		ORDER BY created_at DESC
		LIMIT 1`, projectID, projectID).
		Scan(&d.ID, &d.ProjectID, &d.CommitSHA, &d.ImageTag, &d.Status)
	if errors.Is(err, sql.ErrNoRows){
		return nil, ErrNotFound
	}
	return &d, err
}