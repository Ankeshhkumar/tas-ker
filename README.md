# Tas-ker

A task tracking app (projects, Kanban board, search).

**Stack:** React + Vite, Go (`net/http`), PostgreSQL, Docker Compose.
Coming next: GitHub Actions DevSecOps pipeline, Kubernetes (kind), Argo CD, Prometheus and Grafana.

## Run it

Requires Docker with Compose.

```bash
make up        # creates .env, builds images, starts everything
make smoke     # checks the API and the frontend proxy
```

Open http://localhost:3000. Other commands: `make help`.

## Layout

```
backend/         Go API (handlers, Postgres store, tests, Dockerfile)
frontend/        React app (Board, TaskForm, tests, Dockerfile, nginx.conf)
init/postgres/   schema and seed data, loaded on first start
docker-compose.yml
Makefile
docs/            system design and architecture diagrams
```

## API

| Method | Path | Purpose |
|---|---|---|
| GET | `/projects` | List projects |
| POST | `/projects` | Create a project |
| GET | `/tasks?project_id=N` | List tasks in a project |
| POST | `/tasks` | Create a task |
| PATCH | `/tasks/{id}` | Update a task (e.g. change status) |
| GET | `/search?project_id=N&q=text` | Search tasks by title |
| GET | `/health` | Health check (also pings the database) |

Task statuses: `todo`, `in_progress`, `blocked`, `done`. Priorities: `low`, `medium`, `high`.

## Develop without Docker

```bash
# terminal 1: a Postgres you can reach, then
cd backend && POSTGRES_URL='postgres://tasker:change-me@localhost:5432/tasker?sslmode=disable' go run .
# terminal 2
cd frontend && npm install && npm run dev      # http://localhost:5173
```

## Tests

```bash
make test      # go test + eslint + vitest
```
