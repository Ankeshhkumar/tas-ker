# Tas-ker

![CI](https://github.com/Ankeshhkumar/tas-ker/actions/workflows/ci.yml/badge.svg)

A task manager (projects, Kanban board, search) built to practise a complete DevOps workflow: containers, a DevSecOps CI pipeline, Kubernetes, GitOps and ChatOps. It runs entirely on a laptop at zero cost.

<img width="1205" height="877" alt="image" src="https://github.com/user-attachments/assets/afbf5002-a71b-449e-bbce-44dafa701369" />


**Stack:** React + Vite, Go (`net/http`), PostgreSQL · Docker, GitHub Actions, Docker Hub · kind, Traefik, Argo CD, Sealed Secrets · Slack

![App](docs/screenshots/app.png)

## Architecture

```
Browser -> Traefik (Ingress) -> frontend (Nginx) -> backend (Go) -> PostgreSQL
                                   /api/* proxied
```

![Delivery pipeline](docs/screenshots/pipeline.png)

```
git push -> CI checks + scans -> multi-arch image to Docker Hub (SHA tag)
         -> CI commits new tag to git-ops/ -> Argo CD syncs the cluster -> Slack alert
```

## Run locally (Docker Compose)

```bash
make up       # creates .env, builds images, starts everything
make smoke    # health checks
make test     # go test + eslint + vitest
```

Open http://localhost:3000. More commands: `make help`.

## CI pipeline

Runs on every push and pull request (docs and `git-ops/` changes are skipped).
<img width="1470" height="956" alt="Screenshot 2026-10-25 at 5 03 09 PM" src="https://github.com/user-attachments/assets/dead46ce-beea-4e7d-8e47-2ce5130110fb" />


| Stage | Tool |
|---|---|
| Lint and tests | `go vet`, `go test`, ESLint, Vitest |
| Secret scan | gitleaks |
| Dependency scan | govulncheck, `npm audit` |
| Dockerfile lint | Hadolint |
| Image scan | Trivy (fails on fixable CRITICAL) |
| Publish | Multi-arch (amd64 + arm64) images to Docker Hub, tagged with the commit SHA only |
| Deploy | Bot commits the new tag to `git-ops/`; Argo CD rolls it out |

## Kubernetes (kind + Argo CD)

```bash
kind create cluster --config infra/kind-config.yaml
helm install traefik traefik/traefik -n traefik --create-namespace \
  --set service.type=NodePort --set ports.web.nodePort=30080
# install metrics-server, Argo CD and Sealed Secrets, then:
kubectl apply -f git-ops/argocd/tas-ker-app.yml
```

App: http://localhost:8000. Manifests live in `git-ops/tas-ker-gitops/`.

**What's in the cluster**

- Postgres StatefulSet with a persistent volume; backend and frontend Deployments
- Liveness and readiness probes, resource requests and limits, non-root containers
- HPA on both services (2 to 5 pods at 70% CPU)
- Argo CD with auto-sync, prune and self-heal
- Sealed Secrets: the database password is stored encrypted in Git
- RBAC read-only role; API token not mounted into pods
- Slack notifications for deploys, failed syncs and degraded health

![Argo CD](docs/screenshots/argocd.png)
<img width="1464" height="956" alt="Screenshot 2026-10-25 at 4 29 30 PM" src="https://github.com/user-attachments/assets/9429563b-a839-4776-a907-7afe64a0e69b" />

![Autoscaling](docs/screenshots/hpa.png)
<img width="1464" height="956" alt="Screenshot 2026-10-25 at 4 37 36 PM" src="https://github.com/user-attachments/assets/62155a11-d4be-4c47-81d0-20b1d238353d" />

![Slack](docs/screenshots/slack.png)
<img width="1470" height="956" alt="Screenshot 2026-10-25 at 9 00 33 PM" src="https://github.com/user-attachments/assets/25177373-cf48-4f5b-ba8a-a6ae4dc15dc8" />


## Repository layout

```
backend/            Go API, tests, Dockerfile
frontend/           React app, tests, Nginx config, Dockerfile
init/postgres/      schema and seed data
infra/              kind cluster config
git-ops/            Kubernetes manifests and Argo CD config
.github/workflows/  CI pipeline
docker-compose.yml, Makefile
```

## Design decisions and known limits

- **Traefik, not ingress-nginx:** the community ingress-nginx project was retired in March 2026.
- **SHA tags only:** every running pod traces back to one commit, and rollback is deploying an older SHA.
- **Multi-arch images:** needed because CI builds on amd64 and the dev machine is arm64.
- **NetworkPolicy** is written but not enforced on kind's default network plugin.
- **Sealed secrets** are encrypted for one cluster's key; recreating the cluster means re-sealing.

## Roadmap

- [ ] Prometheus and Grafana dashboards, `/metrics` endpoint
- [ ] OWASP ZAP scan in CI
- [ ] JWT auth service, AI service
