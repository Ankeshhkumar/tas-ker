export const COLUMNS = [
  { status: 'todo', label: 'To do' },
  { status: 'in_progress', label: 'In progress' },
  { status: 'blocked', label: 'Blocked' },
  { status: 'done', label: 'Done' },
]

function TaskCard({ task, onMove }) {
  return (
    <li className={`card priority-${task.priority}`}>
      <p className="card-title">{task.title}</p>
      {task.description && <p className="card-desc">{task.description}</p>}
      <div className="card-meta">
        <span className="chip">{task.priority}</span>
        {task.due_date && <span className="due">Due {task.due_date}</span>}
      </div>
      <label className="move">
        <span className="visually-hidden">Move “{task.title}” to</span>
        <select
          value={task.status}
          onChange={(e) => onMove(task.id, e.target.value)}
        >
          {COLUMNS.map((c) => (
            <option key={c.status} value={c.status}>
              {c.label}
            </option>
          ))}
        </select>
      </label>
    </li>
  )
}

export default function Board({ tasks, onMove }) {
  return (
    <div className="board">
      {COLUMNS.map((col) => {
        const items = tasks.filter((t) => t.status === col.status)
        return (
          <section key={col.status} className={`column status-${col.status}`}>
            <h2>
              {col.label} <span className="count">{items.length}</span>
            </h2>
            {items.length === 0 ? (
              <p className="empty">Nothing here yet.</p>
            ) : (
              <ul>
                {items.map((t) => (
                  <TaskCard key={t.id} task={t} onMove={onMove} />
                ))}
              </ul>
            )}
          </section>
        )
      })}
    </div>
  )
}
