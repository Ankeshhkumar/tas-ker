import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, within } from '@testing-library/react'
import Board from './Board.jsx'

const tasks = [
  { id: 1, title: 'Write docs', status: 'todo', priority: 'low', description: '', due_date: null },
  { id: 2, title: 'Fix login', status: 'in_progress', priority: 'high', description: 'Token bug', due_date: '2026-11-01' },
  { id: 3, title: 'Ship it', status: 'done', priority: 'medium', description: '', due_date: null },
]

describe('Board', () => {
  it('places each task in the column matching its status', () => {
    render(<Board tasks={tasks} onMove={() => {}} />)
    const inProgress = screen.getByRole('heading', { name: /in progress/i }).closest('section')
    expect(within(inProgress).getByText('Fix login')).toBeInTheDocument()
    expect(within(inProgress).queryByText('Write docs')).not.toBeInTheDocument()
  })

  it('shows an empty message for columns without tasks', () => {
    render(<Board tasks={tasks} onMove={() => {}} />)
    const blocked = screen.getByRole('heading', { name: /blocked/i }).closest('section')
    expect(within(blocked).getByText(/nothing here yet/i)).toBeInTheDocument()
  })

  it('calls onMove when a task is moved to another status', () => {
    const onMove = vi.fn()
    render(<Board tasks={tasks} onMove={onMove} />)
    fireEvent.change(screen.getByLabelText(/move “write docs” to/i), {
      target: { value: 'done' },
    })
    expect(onMove).toHaveBeenCalledWith(1, 'done')
  })
})
