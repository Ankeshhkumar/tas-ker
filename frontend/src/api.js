// Thin wrapper around fetch. Every call goes to /api/..., which the proxy
// (Vite in dev, Nginx in Docker, Ingress in Kubernetes) sends to the Go backend.
const BASE = '/api'

async function request(path, options = {}) {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const body = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(body.error || `Request failed (${res.status})`)
  return body
}

export const getProjects = () => request('/projects')

export const createProject = (name) =>
  request('/projects', { method: 'POST', body: JSON.stringify({ name }) })

export const getTasks = (projectId) => request(`/tasks?project_id=${projectId}`)

export const searchTasks = (projectId, q) =>
  request(`/search?project_id=${projectId}&q=${encodeURIComponent(q)}`)

export const createTask = (task) =>
  request('/tasks', { method: 'POST', body: JSON.stringify(task) })

export const updateTask = (id, patch) =>
  request(`/tasks/${id}`, { method: 'PATCH', body: JSON.stringify(patch) })
