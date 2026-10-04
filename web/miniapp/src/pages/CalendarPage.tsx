import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { api } from '@/api/client'
import type { AgendaItem } from '@/api/types'
import { Header } from '@/components/layout/Header'
import { TaskCard } from '@/components/tasks/TaskCard'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
import { cn } from '@/lib/cn'
import { hapticError, hapticLight, hapticSuccess } from '@/lib/telegram'

type CalendarView = 'day' | 'week' | 'month'

const VIEWS: Array<{ id: CalendarView; label: string }> = [
  { id: 'day', label: 'День' },
  { id: 'week', label: 'Неделя' },
  { id: 'month', label: 'Месяц' },
]

const TYPE_FILTERS: Array<{ id: AgendaItem['type']; label: string }> = [
  { id: 'task', label: 'Задачи' },
  { id: 'event', label: 'События' },
  { id: 'reminder', label: 'Напоминания' },
  { id: 'note', label: 'Заметки' },
]

export function CalendarPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [view, setView] = useState<CalendarView>('week')
  const [anchor, setAnchor] = useState(() => toDateKey(new Date()))
  const [selectedDay, setSelectedDay] = useState(() => toDateKey(new Date()))
  const [typesOn, setTypesOn] = useState<Set<AgendaItem['type']]>(
    () => new Set(['task', 'event', 'reminder', 'note']),
  )
  const [createOpen, setCreateOpen] = useState(false)
  const [title, setTitle] = useState('')
  const [kind, setKind] = useState<'task' | 'reminder' | 'meeting'>('task')
  const [dueDate, setDueDate] = useState(() => toDateKey(new Date()))
  const [formError, setFormError] = useState<string | null>(null)

  const window = useMemo(() => computeWindow(view, anchor), [view, anchor])

  const typesParam = useMemo(
    () =>
      TYPE_FILTERS.filter((t) => typesOn.has(t.id))
        .map((t) => t.id)
        .join(','),
    [typesOn],
  )

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['calendar', 'agenda', view, window.from, window.to, typesParam],
    queryFn: () => api.calendarAgenda({ view, from: window.from, to: window.to, types: typesParam }),
  })

  const itemsByDay = useMemo(() => {
    const map = new Map<string, AgendaItem[]>()
    for (const item of data?.items ?? []) {
      const day = dayKeyOf(item.starts_at)
      if (!map.has(day)) map.set(day, [])
      map.get(day)!.push(item)
    }
    return map
  }, [data])

  const shift = (dir: number) => {
    hapticLight()
    const d = parseDay(anchor)
    if (view === 'day') d.setDate(d.getDate() + dir)
    else if (view === 'week') d.setDate(d.getDate() + dir * 7)
    else d.setMonth(d.getMonth() + dir)
    const next = toDateKey(d)
    setAnchor(next)
    setSelectedDay(next)
  }

  const completeTask = useMutation({
    mutationFn: (id: string) => api.completeTask(id),
    onSuccess: () => {
      hapticSuccess()
      void queryClient.invalidateQueries({ queryKey: ['calendar'] })
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
      void queryClient.invalidateQueries({ queryKey: ['calendar'] })
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

  const selectedItems = itemsByDay.get(selectedDay) ?? []
  const totalItems = data?.items?.length ?? 0

  const openCreate = () => {
    setFormError(null)
    setTitle('')
    setKind('task')
    setDueDate(selectedDay)
    setCreateOpen(true)
  }

  const toggleType = (id: AgendaItem['type']) => {
    hapticLight()
    setTypesOn((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      if (next.size === 0) next.add(id) // keep at least one type on
      return next
    })
  }

  return (
    <>
      <Header title="Календарь" subtitle={window.title} />
      <div className="space-y-4 px-4 pb-4">
        {/* View switcher — iOS-like */}
        <div className="flex rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-1">
          {VIEWS.map((v) => (
            <button
              key={v.id}
              type="button"
              onClick={() => {
                hapticLight()
                setView(v.id)
              }}
              className={cn(
                'flex-1 rounded-xl py-2 text-sm font-medium',
                view === v.id
                  ? 'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]'
                  : 'text-[var(--tg-theme-hint-color,#94a3b8)]',
              )}
            >
              {v.label}
            </button>
          ))}
        </div>

        {/* Window navigation */}
        <div className="flex items-center justify-between">
          <Button size="sm" variant="ghost" onClick={() => shift(-1)}>
            ‹
          </Button>
          <button
            type="button"
            className="text-sm text-[var(--tg-theme-link-color,#22c55e)]"
            onClick={() => {
              const today = toDateKey(new Date())
              setAnchor(today)
              setSelectedDay(today)
            }}
          >
            Сегодня
          </button>
          <Button size="sm" variant="ghost" onClick={() => shift(1)}>
            ›
          </Button>
        </div>

        {/* Type filters */}
        <div className="flex flex-wrap gap-2">
          {TYPE_FILTERS.map((t) => (
            <button
              key={t.id}
              type="button"
              onClick={() => toggleType(t.id)}
              className={cn(
                'rounded-full px-3 py-1.5 text-xs',
                typesOn.has(t.id)
                  ? 'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]'
                  : 'bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-[var(--tg-theme-hint-color,#94a3b8)]',
              )}
            >
              {t.label}
            </button>
          ))}
        </div>

        {/* Day strip (week/month) or single day header (day view) */}
        {view !== 'day' && (
          <div className="-mx-1 flex gap-2 overflow-x-auto pb-1">
            {daysInRange(window.from, window.to).map((day) => {
              const count = (itemsByDay.get(day) ?? []).filter((i) => !i.done).length
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
                    <div
                      className={cn(
                        'text-xs',
                        active ? 'opacity-90' : 'text-[var(--tg-theme-link-color,#22c55e)]',
                      )}
                    >
                      {count}
                    </div>
                  )}
                </button>
              )
            })}
          </div>
        )}

        <div className="flex items-center justify-between">
          <h3 className="text-sm font-medium text-[var(--tg-theme-hint-color,#94a3b8)]">
            {formatDayLong(selectedDay)}
          </h3>
          <Button size="sm" onClick={openCreate}>
            <Plus size={16} className="mr-1" />
            Задача
          </Button>
        </div>

        {isLoading && (
          <div className="space-y-2">
            <Skeleton className="h-14 w-full" />
            <Skeleton className="h-14 w-full" />
          </div>
        )}

        {isError && (
          <QueryError message="Не удалось загрузить календарь" onRetry={() => void refetch()} />
        )}

        {!isLoading && !isError && selectedItems.length === 0 && totalItems === 0 && (
          <EmptyState
            title="Пусто в этом периоде"
            description="Добавь задачу, встречу или напоминание"
            actionLabel="Создать"
            onAction={openCreate}
          />
        )}

        {!isLoading && !isError && selectedItems.length === 0 && totalItems > 0 && (
          <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">На этот день пусто</p>
        )}

        <div className="space-y-2">
          {selectedItems.map((item) => (
            <AgendaRow
              key={`${item.type}-${item.id}`}
              item={item}
              onCompleteTask={() => completeTask.mutate(item.id)}
              onOpenTask={() => navigate(`/tasks/${item.id}`)}
              onOpenNote={() => navigate('/notes')}
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
    </>
  )
}

function AgendaRow({
  item,
  onCompleteTask,
  onOpenTask,
  onOpenNote,
}: {
  item: AgendaItem
  onCompleteTask: () => void
  onOpenTask: () => void
  onOpenNote: () => void
}) {
  if (item.type === 'task') {
    return (
      <TaskCard
        title={item.title}
        priority="medium"
        kind="task"
        done={!!item.done}
        detail={timeLabel(item.starts_at)}
        onComplete={item.done ? undefined : onCompleteTask}
        onOpen={onOpenTask}
      />
    )
  }
  const badge =
    item.type === 'event' ? 'Событие' : item.type === 'reminder' ? 'Напоминание' : 'Заметка'
  return (
    <button
      type="button"
      onClick={item.type === 'note' ? onOpenNote : onOpenTask}
      className="flex w-full items-center gap-3 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 text-left"
    >
      <span className="rounded-full bg-[var(--tg-theme-link-color,#22c55e)]/15 px-2 py-0.5 text-xs text-[var(--tg-theme-link-color,#22c55e)]">
        {badge}
      </span>
      <span className="min-w-0 flex-1 truncate text-sm">{item.title}</span>
      <span className="shrink-0 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
        {timeLabel(item.starts_at)}
      </span>
    </button>
  )
}

function computeWindow(view: CalendarView, anchorISO: string): { from: string; to: string; title: string } {
  const base = parseDay(anchorISO)
  if (view === 'day') {
    return {
      from: anchorISO,
      to: anchorISO,
      title: base.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' }),
    }
  }
  if (view === 'week') {
    const wd = (base.getDay() + 6) % 7 // Monday = 0
    const from = new Date(base)
    from.setDate(from.getDate() - wd)
    const to = new Date(from)
    to.setDate(to.getDate() + 6)
    return {
      from: toDateKey(from),
      to: toDateKey(to),
      title: `${formatDayShort(toDateKey(from))} — ${formatDayShort(toDateKey(to))}`,
    }
  }
  const from = new Date(base.getFullYear(), base.getMonth(), 1)
  const to = new Date(base.getFullYear(), base.getMonth() + 1, 0)
  return {
    from: toDateKey(from),
    to: toDateKey(to),
    title: base.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' }),
  }
}

function daysInRange(fromISO: string, toISO: string): string[] {
  const list: string[] = []
  const d = parseDay(fromISO)
  const end = parseDay(toISO)
  while (d <= end) {
    list.push(toDateKey(d))
    d.setDate(d.getDate() + 1)
  }
  return list
}

function parseDay(iso: string): Date {
  const [y, m, d] = iso.split('-').map(Number)
  return new Date(y, (m ?? 1) - 1, d ?? 1)
}

function toDateKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function dayKeyOf(rfc3339: string): string {
  const d = new Date(rfc3339)
  if (Number.isNaN(d.getTime())) return toDateKey(new Date())
  return toDateKey(d)
}

function timeLabel(rfc3339: string): string {
  const d = new Date(rfc3339)
  if (Number.isNaN(d.getTime())) return ''
  if (d.getHours() === 0 && d.getMinutes() === 0) return 'весь день'
  return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}

function formatDayShort(iso: string): string {
  const d = parseDay(iso)
  const today = toDateKey(new Date())
  if (iso === today) return 'Сегодня'
  return d.toLocaleDateString('ru-RU', { weekday: 'short', day: 'numeric' })
}

function formatDayLong(iso: string): string {
  const d = parseDay(iso)
  return d.toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long' })
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
