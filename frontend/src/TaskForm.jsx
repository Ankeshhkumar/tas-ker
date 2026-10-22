import { useState } from 'react'

export default function TaskForm({ onCreate }) {
  const [title, setTitle] = useState('')
  const [priority, setPriority] = useState('medium')
  const [dueDate, setDueDate] = useState('')

  function submit(e) {
    e.preventDefault()
    if (!title.trim()) return
    onCreate({ title: title.trim(), priority, due_date: dueDate || null })
    setTitle('')
    setDueDate('')
  }

  return (
    <form className="task-form" onSubmit={submit}>
      <label className="grow">
        <span className="visually-hidden">Task title</span>
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="What needs doing?"
          maxLength={300}
        />
      </label>
      <label>
        <span className="visually-hidden">Priority</span>
        <select value={priority} onChange={(e) => setPriority(e.target.value)}>
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
        </select>
      </label>
      <label>
        <span className="visually-hidden">Due date</span>
        <input
          type="date"
          value={dueDate}
          onChange={(e) => setDueDate(e.target.value)}
        />
      </label>
      <button type="submit" className="primary">
        Add task
      </button>
    </form>
  )
}
