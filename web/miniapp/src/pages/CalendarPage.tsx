import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { api } from '@/api/client'
import type { Task } from '@/api/types'
import { Header } from '@/components/layout/Header'
import { CreateEventSheet } from '@/components/calendar/CreateEventSheet'
import { TaskCard } from '@/components/tasks/TaskCard'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
import { cn } from '@/lib/cn'
import { hapticError, hapticLight, hapticSuccess } from '@/lib/telegram'

const RANGE_DAYS = 14

export function CalendarPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [selectedDay, setSelectedDay] = useState(() => toDateKey(new Date()))
  const [createOpen, setCreateOpen] = useState(false)
  const [eventOpen, setEventOpen] = useState(false)
  const [title, setTitle] = useState('')
  const [kind, setKind] = useState<'task' | 'reminder' | 'meeting'>('task')
  const [dueDate, setDueDate] = useState(() => toDateKey(new Date()))
  const [formError, setFormError] = useState<string | null>(null)
  // TASK-011 п.4: filter the agenda by sphere (multi-select chips).
  const [selectedSpheres, setSelectedSpheres] = useState<string[]>([])

  const range = useMemo(() => {
    const from = new Date()
    const to = new Date()
    to.setDate(to.getDate() + RANGE_DAYS - 1)
    return { from: toDateKey(from), to: toDateKey(to) }
  }, [])

  const days = useMemo(() => {
    const list: string[] = []
    const d = new Date()
    for (let i = 0; i < RANGE_DAYS; i++) {
      list.push(toDateKey(d))
      d.setDate(d.getDate() + 1)
    }
    return list
  }, [])

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['tasks', 'calendar', range.from, range.to, selectedSpheres],
    queryFn: async () => {
      const res = await api.tasksDueBetween(range.from, range.to)
      const all = Array.isArray(res.tasks) ? res.tasks : []
      if (selectedSpheres.length === 0) return all
      return all.filter((t) => (t.sphere_ids ?? []).some((s) => selectedSpheres.includes(s)))
    },
  })

  // Sphere list for the filter chips (TASK-011 п.4).
  const { data: spheresData } = useQuery({
    queryKey: ['settings', 'spheres'],
    queryFn: async () => {
      const res = await api.spheres()
      return Array.isArray(res.spheres) ? res.spheres : []
    },
  })
  const spheres = spheresData ?? []

  const toggleSphere = (id: string) => {
    hapticLight()
    setSelectedSpheres((prev) =>
      prev.includes(id) ? prev.filter((s) => s !== id) : [...prev, id],
    )
  }

  // MA-B5: calendar events for the selected day (user-local TZ resolved on the server).
  const { data: eventsData } = useQuery({
    queryKey: ['calendar', 'day', selectedDay],
    queryFn: async () => {
      const res = await api.calendarToday(selectedDay)
      return Array.isArray(res.events) ? res.events : []
    },
  })

  const complete = useMutation({
    mutationFn: (id: string) => api.completeTask(id),
    onSuccess: () => {
      hapticSuccess()
      void queryClient.invalidateQueries({ queryKey: ['tasks'] })
    },
    onError: () => hapticError(),
  })

  const create = useMutation({
    mutationFn: () =>
      api.createTask({
        title: title.trim(),
        kind,
        due_date: dueDate,
      }),
    onSuccess: () => {
      hapticSuccess()
      void queryClient.invalidateQueries({ queryKey: ['tasks'] })
      setCreateOpen(false)
      setTitle('')
      setKind('task')
      setDueDate(selectedDay)
      setFormError(null)
    },
    onError: (err) => {
      hapticError()
      setFormError(ruApiError(err, 'Не удалось создать задачу'))
    },
  })

  const tasksByDay = useMemo(() => {
    const map = new Map<string, Task[]>()
    for (const day of days) map.set(day, [])
    for (const t of data ?? []) {
      if (!t.due_date || t.status === 'cancelled') continue
      const list = map.get(t.due_date)
      if (list) list.push(t)
    }
    return map
  }, [data, days])

  const selectedTasks = (tasksByDay.get(selectedDay) ?? []).filter(
    (t) => t.status !== 'cancelled',
  )
  const totalCount = (data ?? []).filter((t) => t.status !== 'cancelled' && t.due_date).length

  const openCreate = () => {
    setFormError(null)
    setTitle('')
    setKind('task')
    setDueDate(selectedDay)
    setCreateOpen(true)
  }

  return (
    <>
      <Header title="Календарь" subtitle="Задачи по датам на 2 недели" />
      <div className="space-y-4 px-4 pb-4">
        <div className="flex items-center justify-end gap-2">
          <Button size="sm" variant="secondary" onClick={() => setEventOpen(true)}>
            <Plus size={16} className="mr-1" />
            Событие
          </Button>
          <Button size="sm" onClick={openCreate}>
            <Plus size={16} className="mr-1" />
            Задача
          </Button>
        </div>

        {spheres.length > 0 && (
          <div className="-mx-1 flex gap-2 overflow-x-auto pb-1">
            <button
              type="button"
              onClick={() => {
                hapticLight()
                setSelectedSpheres([])
              }}
              className={cn(
                'shrink-0 rounded-full px-3 py-1.5 text-sm',
                selectedSpheres.length === 0
                  ? 'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]'
                  : 'bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-[var(--tg-theme-hint-color,#94a3b8)]',
              )}
            >
              Все сферы
            </button>
            {spheres.map((s) => {
              const active = selectedSpheres.includes(s.id)
              return (
                <button
                  key={s.id}
                  type="button"
                  onClick={() => toggleSphere(s.id)}
                  className={cn(
                    'shrink-0 rounded-full px-3 py-1.5 text-sm',
                    active
                      ? 'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]'
                      : 'bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-[var(--tg-theme-hint-color,#94a3b8)]',
                  )}
                >
                  {s.name}
                </button>
              )
            })}
          </div>
        )}

        <div className="-mx-1 flex gap-2 overflow-x-auto pb-1">
          {days.map((day) => {
            const count = (tasksByDay.get(day) ?? []).filter(
              (t) => t.status !== 'done' && t.status !== 'cancelled',
            ).length
            const active = day === selectedDay
            return (
              <button
                key={day}
                type="button"
                onClick={() => {
                  hapticLight()
                  setSelectedDay(day)
                }}
                className={cn(
                  'shrink-0 rounded-2xl px-3 py-2 text-center text-sm',
                  active
                    ? 'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]'
                    : 'bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-[var(--tg-theme-hint-color,#94a3b8)]',
                )}
              >
                <div className="font-medium">{formatDayShort(day)}</div>
                {count > 0 && (
                  <div className={cn('text-xs', active ? 'opacity-90' : 'text-[var(--tg-theme-link-color,#22c55e)]')}>
                    {count}
                  </div>
                )}
              </button>
            )
          })}
        </div>

        <h3 className="text-sm font-medium text-[var(--tg-theme-hint-color,#94a3b8)]">
          {formatDayLong(selectedDay)}
        </h3>

        {(eventsData?.length ?? 0) > 0 && (
          <div className="space-y-2">
            <p className="text-xs font-medium uppercase tracking-wide text-[var(--tg-theme-hint-color,#64748b)]">
              События
            </p>
            {eventsData!.map((ev) => (
              <div
                key={ev.id}
                className="flex items-center gap-3 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3"
              >
                <span className="shrink-0 text-sm font-medium text-[var(--tg-theme-link-color,#22c55e)]">
                  {formatTime(ev.starts_at)}
                </span>
                <span className="min-w-0 flex-1 truncate">{ev.title}</span>
              </div>
            ))}
          </div>
        )}

        {isLoading && (
          <div className="space-y-2">
            <Skeleton className="h-14 w-full" />
            <Skeleton className="h-14 w-full" />
          </div>
        )}

        {isError && (
          <QueryError message="Не удалось загрузить задачи" onRetry={() => void refetch()} />
        )}

        {!isLoading && !isError && selectedTasks.length === 0 && (eventsData?.length ?? 0) === 0 && totalCount === 0 && (
          <EmptyState
            title="Нет задач с датой"
            description="Добавь задачу, встречу или напоминание"
            actionLabel="Создать"
            onAction={openCreate}
          />
        )}

        {!isLoading && !isError && selectedTasks.length === 0 && (eventsData?.length ?? 0) === 0 && totalCount > 0 && (
          <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">На этот день пусто</p>
        )}

        <div className="space-y-2">
          {selectedTasks.map((t) => (
            <TaskCard
              key={t.id}
              title={t.title}
              priority={t.priority}
              kind={t.kind}
              done={t.status === 'done'}
              detail={t.address}
              onComplete={
                t.status === 'done'
                  ? undefined
                  : () => {
                      hapticLight()
                      complete.mutate(t.id)
                    }
              }
              onOpen={() => navigate(`/tasks/${t.id}`)}
            />
          ))}
        </div>
      </div>

      <Sheet
        open={createOpen}
        onClose={() => {
          setCreateOpen(false)
          setFormError(null)
        }}
        title="Новая задача"
      >
        <input
          value={title}
          onChange={(e) => {
            setTitle(e.target.value)
            if (formError) setFormError(null)
          }}
          placeholder="Название"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          autoFocus
        />
        <p className="mb-2 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Тип</p>
        <div className="mb-3 flex flex-wrap gap-2">
          {(['task', 'reminder', 'meeting'] as const).map((k) => (
            <button
              key={k}
              type="button"
              onClick={() => setKind(k)}
              className={cn(
                'rounded-full px-3 py-1.5 text-sm',
                kind === k
                  ? 'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]'
                  : 'bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-[var(--tg-theme-hint-color,#94a3b8)]',
              )}
            >
              {kindLabel(k)}
            </button>
          ))}
        </div>
        <label className="mb-3 block text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
          Дата
          <input
            type="date"
            value={dueDate}
            onChange={(e) => setDueDate(e.target.value)}
            className="mt-1 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          />
        </label>
        {formError && (
          <p className="mb-3 text-sm text-rose-400" role="alert">
            {formError}
          </p>
        )}
        <Button
          className="w-full"
          disabled={!title.trim() || !dueDate || create.isPending}
          onClick={() => create.mutate()}
        >
          Создать
        </Button>
      </Sheet>

      <CreateEventSheet open={eventOpen} onClose={() => setEventOpen(false)} day={selectedDay} />
    </>
  )
}

function toDateKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

/** Format an RFC3339 instant as HH:mm in the user's browser locale. */
function formatTime(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}

function formatDayShort(iso: string): string {
  const d = new Date(`${iso}T12:00:00`)
  if (Number.isNaN(d.getTime())) return iso
  const today = toDateKey(new Date())
  if (iso === today) return 'Сегодня'
  return d.toLocaleDateString('ru-RU', { weekday: 'short', day: 'numeric' })
}

function formatDayLong(iso: string): string {
  const d = new Date(`${iso}T12:00:00`)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleDateString('ru-RU', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  })
}

function kindLabel(k: 'task' | 'reminder' | 'meeting') {
  switch (k) {
    case 'reminder':
      return 'Напоминание'
    case 'meeting':
      return 'Встреча'
    default:
      return 'Задача'
  }
}
