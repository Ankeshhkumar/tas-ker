package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	validStatus   = map[string]bool{"todo": true, "in_progress": true, "blocked": true, "done": true}
	validPriority = map[string]bool{"low": true, "medium": true, "high": true}
)

// Server holds the dependencies the handlers use.
type Server struct{ store Store }

// Routes wires URLs to handlers. Go 1.22 lets us put the method in the pattern.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /projects", s.listProjects)
	mux.HandleFunc("POST /projects", s.createProject)
	mux.HandleFunc("GET /tasks", s.listTasks)
	mux.HandleFunc("POST /tasks", s.createTask)
	mux.HandleFunc("PATCH /tasks/{id}", s.updateTask)
	mux.HandleFunc("GET /search", s.searchTasks)
	mux.Handle("GET /metrics", metricsHandler())
	return logRequests(metricsMiddleware(mux))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	ps, err := s.store.ListProjects(r.Context())
	respond(w, http.StatusOK, ps, err)
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len(in.Name) > 200 {
		writeError(w, http.StatusBadRequest, "name is required (max 200 characters)")
		return
	}
	p, err := s.store.CreateProject(r.Context(), in.Name, in.Description)
	respond(w, http.StatusCreated, p, err)
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	pid, ok := projectID(w, r)
	if !ok {
		return
	}
	ts, err := s.store.ListTasks(r.Context(), pid)
	respond(w, http.StatusOK, ts, err)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		ProjectID   int     `json:"project_id"`
		Status      string  `json:"status"`
		Priority    string  `json:"priority"`
		DueDate     *string `json:"due_date"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" || len(in.Title) > 300 {
		writeError(w, http.StatusBadRequest, "title is required (max 300 characters)")
		return
	}
	if in.ProjectID <= 0 {
		writeError(w, http.StatusBadRequest, "project_id is required")
		return
	}
	if in.Status == "" {
		in.Status = "todo"
	}
	if in.Priority == "" {
		in.Priority = "medium"
	}
	if !validStatus[in.Status] || !validPriority[in.Priority] {
		writeError(w, http.StatusBadRequest, "invalid status or priority")
		return
	}
	if !validDate(in.DueDate) {
		writeError(w, http.StatusBadRequest, "due_date must be YYYY-MM-DD")
		return
	}
	t, err := s.store.CreateTask(r.Context(), Task{
		Title: in.Title, Description: in.Description, ProjectID: in.ProjectID,
		Status: in.Status, Priority: in.Priority, DueDate: in.DueDate,
	})
	respond(w, http.StatusCreated, t, err)
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	var p TaskPatch
	if !decode(w, r, &p) {
		return
	}
	if p.Title != nil && (strings.TrimSpace(*p.Title) == "" || len(*p.Title) > 300) {
		writeError(w, http.StatusBadRequest, "title must be 1-300 characters")
		return
	}
	if p.Status != nil && !validStatus[*p.Status] {
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}
	if p.Priority != nil && !validPriority[*p.Priority] {
		writeError(w, http.StatusBadRequest, "invalid priority")
		return
	}
	if !validDate(p.DueDate) {
		writeError(w, http.StatusBadRequest, "due_date must be YYYY-MM-DD")
		return
	}
	t, err := s.store.UpdateTask(r.Context(), id, p)
	respond(w, http.StatusOK, t, err)
}

func (s *Server) searchTasks(w http.ResponseWriter, r *http.Request) {
	pid, ok := projectID(w, r)
	if !ok {
		return
	}
	ts, err := s.store.SearchTasks(r.Context(), pid, r.URL.Query().Get("q"))
	respond(w, http.StatusOK, ts, err)
}

// ---- helpers ----

func projectID(w http.ResponseWriter, r *http.Request) (int, bool) {
	pid, err := strconv.Atoi(r.URL.Query().Get("project_id"))
	if err != nil || pid <= 0 {
		writeError(w, http.StatusBadRequest, "project_id query parameter is required")
		return 0, false
	}
	return pid, true
}

func validDate(d *string) bool {
	if d == nil || *d == "" {
		return true
	}
	_, err := time.Parse("2006-01-02", *d)
	return err == nil
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

func respond(w http.ResponseWriter, okStatus int, v any, err error) {
	switch {
	case errors.Is(err, errNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, okStatus, v)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
