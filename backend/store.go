package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Project is a collection of tasks.
type Project struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     *int      `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// Task is one card on the board.
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ProjectID   int       `json:"project_id"`
	AssigneeID  *int      `json:"assignee_id"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	DueDate     *string   `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaskPatch holds the fields a client may change. nil means "leave as is".
type TaskPatch struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	DueDate     *string `json:"due_date"`
}

var errNotFound = errors.New("not found")

// Store is everything the HTTP handlers need from a database.
// Handlers depend on this interface, so tests can use a fake instead of Postgres.
type Store interface {
	Ping(ctx context.Context) error
	ListProjects(ctx context.Context) ([]Project, error)
	CreateProject(ctx context.Context, name, description string) (Project, error)
	ListTasks(ctx context.Context, projectID int) ([]Task, error)
	CreateTask(ctx context.Context, t Task) (Task, error)
	UpdateTask(ctx context.Context, id int, p TaskPatch) (Task, error)
	SearchTasks(ctx context.Context, projectID int, q string) ([]Task, error)
}

type pgStore struct{ db *sql.DB }

func newPGStore(db *sql.DB) *pgStore { return &pgStore{db: db} }

func (s *pgStore) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

func (s *pgStore) ListProjects(ctx context.Context) ([]Project, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, COALESCE(description,''), owner_id, created_at FROM projects ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Project{}
	for rows.Next() {
		var p Project
		var owner sql.NullInt64
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &owner, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.OwnerID = nullInt(owner)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *pgStore) CreateProject(ctx context.Context, name, description string) (Project, error) {
	var p Project
	var owner sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO projects (name, description) VALUES ($1, $2)
		 RETURNING id, name, COALESCE(description,''), owner_id, created_at`,
		name, description).Scan(&p.ID, &p.Name, &p.Description, &owner, &p.CreatedAt)
	p.OwnerID = nullInt(owner)
	return p, err
}

const taskCols = `id, title, COALESCE(description,''), project_id, assignee_id, status, priority,
	to_char(due_date, 'YYYY-MM-DD'), created_at, updated_at`

type scanner interface{ Scan(dest ...any) error }

func scanTask(r scanner) (Task, error) {
	var t Task
	var assignee sql.NullInt64
	var due sql.NullString
	err := r.Scan(&t.ID, &t.Title, &t.Description, &t.ProjectID, &assignee,
		&t.Status, &t.Priority, &due, &t.CreatedAt, &t.UpdatedAt)
	t.AssigneeID = nullInt(assignee)
	if due.Valid {
		t.DueDate = &due.String
	}
	return t, err
}

func scanTasks(rows *sql.Rows) ([]Task, error) {
	defer rows.Close()
	out := []Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *pgStore) ListTasks(ctx context.Context, projectID int) ([]Task, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+taskCols+` FROM tasks WHERE project_id = $1 ORDER BY id`, projectID)
	if err != nil {
		return nil, err
	}
	return scanTasks(rows)
}

func (s *pgStore) CreateTask(ctx context.Context, t Task) (Task, error) {
	row := s.db.QueryRowContext(ctx,
		`INSERT INTO tasks (title, description, project_id, status, priority, due_date)
		 VALUES ($1, $2, $3, $4, $5, $6::date)
		 RETURNING `+taskCols,
		t.Title, t.Description, t.ProjectID, t.Status, t.Priority, t.DueDate)
	return scanTask(row)
}

func (s *pgStore) UpdateTask(ctx context.Context, id int, p TaskPatch) (Task, error) {
	row := s.db.QueryRowContext(ctx,
		`UPDATE tasks SET
		   title       = COALESCE($2, title),
		   description = COALESCE($3, description),
		   status      = COALESCE($4, status),
		   priority    = COALESCE($5, priority),
		   due_date    = COALESCE($6::date, due_date)
		 WHERE id = $1
		 RETURNING `+taskCols,
		id, p.Title, p.Description, p.Status, p.Priority, p.DueDate)
	t, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, errNotFound
	}
	return t, err
}

func (s *pgStore) SearchTasks(ctx context.Context, projectID int, q string) ([]Task, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+taskCols+` FROM tasks
		 WHERE project_id = $1 AND title ILIKE '%' || $2 || '%' ORDER BY id`, projectID, q)
	if err != nil {
		return nil, err
	}
	return scanTasks(rows)
}

func nullInt(n sql.NullInt64) *int {
	if !n.Valid {
		return nil
	}
	v := int(n.Int64)
	return &v
}
