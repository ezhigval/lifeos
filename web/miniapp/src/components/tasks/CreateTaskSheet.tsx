import { useEffect, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Calendar } from 'lucide-react'
import { api } from '@/api/client'
import { Sheet } from '@/components/ui/Sheet'
import { Button } from '@/components/ui/Button'
import { ruApiError } from '@/lib/apiError'
import { hapticError, hapticSuccess } from '@/lib/telegram'

type Props = {
  open: boolean
  onClose: () => void
}

const PRIORITIES: Array<{ value: string; label: string }> = [
  { value: 'low', label: 'Низкий' },
  { value: 'medium', label: 'Средний' },
  { value: 'high', label: 'Высокий' },
  { value: 'urgent', label: 'Срочный' },
]

const KINDS: Array<{ value: 'task' | 'reminder' | 'meeting'; label: string }> = [
  { value: 'task', label: 'Задача' },
  { value: 'reminder', label: 'Напоминание' },
  { value: 'meeting', label: 'Встреча' },
]

/** Quick due-date presets per UX_UI_PLAN (MA-B1: title + optional priority/due). */
function isoDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function presetDue(offsetDays: number): string {
  const d = new Date()
  d.setDate(d.getDate() + offsetDays)
  return isoDate(d)
}

const DUE_PRESETS: Array<{ label: string; value: string }> = [
  { label: 'Сегодня', value: presetDue(0) },
  { label: 'Завтра', value: presetDue(1) },
  { label: 'Через неделю', value: presetDue(7) },
]

export function CreateTaskSheet({ open, onClose }: Props) {
  const queryClient = useQueryClient()
  const [title, setTitle] = useState('')
  const [priority, setPriority] = useState('medium')
  const [kind, setKind] = useState<'task' | 'reminder' | 'meeting'>('task')
  const [dueDate, setDueDate] = useState<string>('')
  const [error, setError] = useState<string | null>(null)

  // Reset form when the sheet closes.
  useEffect(() => {
    if (!open) {
      setTitle('')
      setPriority('medium')
      setKind('task')
      setDueDate('')
      setError(null)
    }
  }, [open])

  const create = useMutation({
    mutationFn: () =>
      api.createTask({
        title: title.trim(),
        priority,
        kind,
        due_date: dueDate || undefined,
      }),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      queryClient.invalidateQueries({ queryKey: ['priorities'] })
      queryClient.invalidateQueries({ queryKey: ['project-tasks'] })
      onClose()
    },
    onError: (err) => {
      hapticError()
      setError(ruApiError(err, 'Не удалось создать задачу'))
    },
  })

  const canSubmit = title.trim().length > 0 && !create.isPending

  function handleSubmit() {
    if (!canSubmit) return
    setError(null)
    create.mutate()
  }

  return (
    <Sheet open={open} onClose={onClose} title="Новая задача">
      <div className="space-y-4">
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Что нужно сделать?"
          autoFocus
          className="w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 text-base outline-none placeholder:text-[var(--tg-theme-hint-color,#94a3b8)]"
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleSubmit()
          }}
        />

        <div>
          <p className="mb-2 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Тип</p>
          <div className="flex gap-2">
            {KINDS.map((k) => (
              <button
                key={k.value}
                type="button"
                onClick={() => setKind(k.value)}
                className={
                  kind === k.value
                    ? 'rounded-full bg-[var(--tg-theme-button-color,#22c55e)] px-3 py-1.5 text-sm text-white'
                    : 'rounded-full bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-3 py-1.5 text-sm'
                }
              >
                {k.label}
              </button>
            ))}
          </div>
        </div>

        <div>
          <p className="mb-2 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Приоритет</p>
          <div className="flex flex-wrap gap-2">
            {PRIORITIES.map((p) => (
              <button
                key={p.value}
                type="button"
                onClick={() => setPriority(p.value)}
                className={
                  priority === p.value
                    ? 'rounded-full bg-[var(--tg-theme-button-color,#22c55e)] px-3 py-1.5 text-sm text-white'
                    : 'rounded-full bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-3 py-1.5 text-sm'
                }
              >
                {p.label}
              </button>
            ))}
          </div>
        </div>

        <div>
          <p className="mb-2 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Срок</p>
          <div className="mb-2 flex flex-wrap gap-2">
            {DUE_PRESETS.map((d) => (
              <button
                key={d.value}
                type="button"
                onClick={() => setDueDate(dueDate === d.value ? '' : d.value)}
                className={
                  dueDate === d.value
                    ? 'rounded-full bg-[var(--tg-theme-button-color,#22c55e)] px-3 py-1.5 text-sm text-white'
                    : 'rounded-full bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-3 py-1.5 text-sm'
                }
              >
                {d.label}
              </button>
            ))}
          </div>
          <label className="flex items-center gap-2 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3">
            <Calendar size={16} className="shrink-0 text-[var(--tg-theme-hint-color,#94a3b8)]" />
            <input
              type="date"
              value={dueDate}
              min={isoDate(new Date())}
              onChange={(e) => setDueDate(e.target.value)}
              className="w-full bg-transparent text-sm outline-none"
            />
          </label>
        </div>

        {error && (
          <p className="text-sm text-rose-400" role="alert">
            {error}
          </p>
        )}

        <Button className="w-full" disabled={!canSubmit} onClick={handleSubmit}>
          {create.isPending ? 'Создаём…' : 'Добавить'}
        </Button>
      </div>
    </Sheet>
  )
}
