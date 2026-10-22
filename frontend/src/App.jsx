import { useCallback, useEffect, useState } from 'react'
import * as api from './api.js'
import Board from './Board.jsx'
import TaskForm from './TaskForm.jsx'

export default function App() {
  const [projects, setProjects] = useState([])
  const [activeId, setActiveId] = useState(null)
  const [tasks, setTasks] = useState([])
  const [query, setQuery] = useState('')
  const [newProject, setNewProject] = useState('')
  const [error, setError] = useState('')

  const fail = (e) => setError(e.message)

  useEffect(() => {
    api
      .getProjects()
      .then((ps) => {
        setProjects(ps)
        if (ps.length) setActiveId((id) => id ?? ps[0].id)
      })
      .catch(fail)
  }, [])

  const loadTasks = useCallback(() => {
    if (activeId == null) return
    const q = query.trim()
    const req = q ? api.searchTasks(activeId, q) : api.getTasks(activeId)
    req
      .then((ts) => {
        setTasks(ts)
        setError('')
      })
      .catch(fail)
  }, [activeId, query])

  useEffect(() => {
    loadTasks()
  }, [loadTasks])

  const active = projects.find((p) => p.id === activeId)

  const addTask = (fields) =>
    api
      .createTask({ ...fields, project_id: activeId })
      .then(loadTasks)
      .catch(fail)

  const moveTask = (id, status) =>
    api.updateTask(id, { status }).then(loadTasks).catch(fail)

  const addProject = (e) => {
    e.preventDefault()
    if (!newProject.trim()) return
    api
      .createProject(newProject.trim())
      .then((p) => {
        setProjects((ps) => [...ps, p])
        setActiveId(p.id)
        setNewProject('')
      })
      .catch(fail)
  }

  return (
    <div className="app">
      <aside className="sidebar">
        <h1 className="brand">Tas-ker</h1>
        <nav aria-label="Projects">
          <ul>
            {projects.map((p) => (
              <li key={p.id}>
                <button
                  className={p.id === activeId ? 'project active' : 'project'}
                  onClick={() => setActiveId(p.id)}
                  aria-current={p.id === activeId ? 'true' : undefined}
                >
                  {p.name}
                </button>
              </li>
            ))}
          </ul>
        </nav>
        <form onSubmit={addProject} className="project-form">
          <label>
            <span className="visually-hidden">New project name</span>
            <input
              value={newProject}
              onChange={(e) => setNewProject(e.target.value)}
              placeholder="New project"
              maxLength={200}
            />
          </label>
          <button type="submit">Add</button>
        </form>
      </aside>

      <main className="main">
        {error && (
          <p role="alert" className="error">
            {error}
          </p>
        )}
        {active ? (
          <>
            <header className="main-head">
              <div>
                <h2 className="project-name">{active.name}</h2>
                {active.description && (
                  <p className="project-desc">{active.description}</p>
                )}
              </div>
              <label className="search">
                <span className="visually-hidden">Search tasks</span>
                <input
                  type="search"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search tasks"
                />
              </label>
            </header>
            <TaskForm onCreate={addTask} />
            <Board tasks={tasks} onMove={moveTask} />
          </>
        ) : (
          <p className="empty">Create a project to start adding tasks.</p>
        )}
      </main>
    </div>
  )
}
