package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fakeStore is an in-memory Store so handler tests need no database.
type fakeStore struct {
	projects []Project
	tasks    []Task
	pingErr  error
}

func (f *fakeStore) Ping(context.Context) error { return f.pingErr }
func (f *fakeStore) ListProjects(context.Context) ([]Project, error) {
	return f.projects, nil
}
func (f *fakeStore) CreateProject(_ context.Context, name, desc string) (Project, error) {
	p := Project{ID: len(f.projects) + 1, Name: name, Description: desc, CreatedAt: time.Now()}
	f.projects = append(f.projects, p)
	return p, nil
}
func (f *fakeStore) ListTasks(_ context.Context, pid int) ([]Task, error) {
	out := []Task{}
	for _, t := range f.tasks {
		if t.ProjectID == pid {
			out = append(out, t)
		}
	}
	return out, nil
}
func (f *fakeStore) CreateTask(_ context.Context, t Task) (Task, error) {
	t.ID = len(f.tasks) + 1
	f.tasks = append(f.tasks, t)
	return t, nil
}
func (f *fakeStore) UpdateTask(_ context.Context, id int, p TaskPatch) (Task, error) {
	for i := range f.tasks {
		if f.tasks[i].ID == id {
			if p.Status != nil {
				f.tasks[i].Status = *p.Status
			}
			return f.tasks[i], nil
		}
	}
	return Task{}, errNotFound
}
func (f *fakeStore) SearchTasks(_ context.Context, pid int, q string) ([]Task, error) {
	out := []Task{}
	for _, t := range f.tasks {
		if t.ProjectID == pid && strings.Contains(strings.ToLower(t.Title), strings.ToLower(q)) {
			out = append(out, t)
		}
	}
	return out, nil
}

func do(t *testing.T, s *fakeStore, method, url, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	rec := httptest.NewRecorder()
	(&Server{store: s}).Routes().ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	if rec := do(t, &fakeStore{}, "GET", "/health", ""); rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	down := &fakeStore{pingErr: context.DeadlineExceeded}
	if rec := do(t, down, "GET", "/health", ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503 when db is down, got %d", rec.Code)
	}
}

func TestCreateAndListProjects(t *testing.T) {
	s := &fakeStore{}
	if rec := do(t, s, "POST", "/projects", `{"name":"Launch"}`); rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", rec.Code)
	}
	if rec := do(t, s, "POST", "/projects", `{"name":"  "}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("blank name: want 400, got %d", rec.Code)
	}
	rec := do(t, s, "GET", "/projects", "")
	var got []Project
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || len(got) != 1 {
		t.Fatalf("want 1 project, got %v (err %v)", got, err)
	}
}

func TestCreateTaskValidation(t *testing.T) {
	cases := []struct {
		name, body string
		want       int
	}{
		{"ok with defaults", `{"title":"Write docs","project_id":1}`, http.StatusCreated},
		{"missing title", `{"project_id":1}`, http.StatusBadRequest},
		{"missing project", `{"title":"x"}`, http.StatusBadRequest},
		{"bad status", `{"title":"x","project_id":1,"status":"nope"}`, http.StatusBadRequest},
		{"bad date", `{"title":"x","project_id":1,"due_date":"tomorrow"}`, http.StatusBadRequest},
		{"bad json", `{`, http.StatusBadRequest},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if rec := do(t, &fakeStore{}, "POST", "/tasks", c.body); rec.Code != c.want {
				t.Fatalf("want %d, got %d (%s)", c.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestListTasksNeedsProjectID(t *testing.T) {
	if rec := do(t, &fakeStore{}, "GET", "/tasks", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	s := &fakeStore{tasks: []Task{{ID: 1, Title: "A", ProjectID: 1, Status: "todo"}}}
	if rec := do(t, s, "PATCH", "/tasks/1", `{"status":"done"}`); rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if s.tasks[0].Status != "done" {
		t.Fatalf("status not updated: %s", s.tasks[0].Status)
	}
	if rec := do(t, s, "PATCH", "/tasks/1", `{"status":"nope"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad status: want 400, got %d", rec.Code)
	}
	if rec := do(t, s, "PATCH", "/tasks/99", `{"status":"done"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("missing task: want 404, got %d", rec.Code)
	}
}

func TestSearch(t *testing.T) {
	s := &fakeStore{tasks: []Task{
		{ID: 1, Title: "Fix login bug", ProjectID: 1},
		{ID: 2, Title: "Write docs", ProjectID: 1},
	}}
	rec := do(t, s, "GET", "/search?project_id=1&q=LOGIN", "")
	var got []Task
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("unexpected search result: %v", got)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	s := &fakeStore{}
	do(t, s, "GET", "/health", "") // generate one counted request
	rec := do(t, s, "GET", "/metrics", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "http_requests_total") {
		t.Fatal("metrics output is missing http_requests_total")
	}
}
