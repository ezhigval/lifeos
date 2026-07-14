import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Clock, Plus } from 'lucide-react'
import { api } from '@/api/client'
import type { CalendarEvent } from '@/api/types'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { ListRow } from '@/components/ui/ListRow'
import { ruApiError } from '@/lib/apiError'
import { hapticError, hapticSuccess } from '@/lib/telegram'

export function CalendarPage() {
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [title, setTitle] = useState('')
  const [time, setTime] = useState(() => defaultTimeLocal())
  const [formError, setFormError] = useState<string | null>(null)

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['calendar', 'today'],
    queryFn: async () => {
      const res = await api.calendarToday()
      return Array.isArray(res.events) ? res.events : []
    },
  })

  const create = useMutation({
    mutationFn: () => {
      const startsAt = localTimeToRFC3339(time)
      return api.createCalendarEvent(title.trim(), startsAt)
    },
    onSuccess: (ev) => {
      hapticSuccess()
      queryClient.setQueryData<CalendarEvent[]>(['calendar', 'today'], (old) => {
        const next = [...(old ?? []), ev]
        next.sort((a, b) => a.starts_at.localeCompare(b.starts_at))
        return next
      })
      void queryClient.invalidateQueries({ queryKey: ['calendar'] })
      setCreateOpen(false)
      setTitle('')
      setTime(defaultTimeLocal())
      setFormError(null)
    },
    onError: (err) => {
      hapticError()
      setFormError(ruApiError(err, 'Не удалось создать событие'))
    },
  })

  const events = useMemo(
    () => [...(data ?? [])].sort((a, b) => a.starts_at.localeCompare(b.starts_at)),
    [data],
  )

  const todayLabel = new Date().toLocaleDateString('ru-RU', {
    weekday: 'short',
    day: 'numeric',
    month: 'long',
  })

  const openCreate = () => {
    setFormError(null)
    setTitle('')
    setTime(defaultTimeLocal())
    setCreateOpen(true)
  }

  return (
    <>
      <Header title="Календарь" subtitle={todayLabel} />
      <div className="space-y-4 px-4 pb-4">
        {events.length > 0 && (
          <div className="flex justify-end">
            <Button size="sm" onClick={openCreate}>
              <Plus size={16} className="mr-1" />
              Событие
            </Button>
          </div>
        )}

        {isLoading && (
          <div className="space-y-2">
            <Skeleton className="h-14 w-full" />
            <Skeleton className="h-14 w-full" />
          </div>
        )}

        {isError && (
          <QueryError message="Не удалось загрузить календарь" onRetry={() => void refetch()} />
        )}

        {!isLoading && !isError && events.length === 0 && (
          <EmptyState
            title="На сегодня пусто"
            description="Добавь встречу или напоминание"
            actionLabel="Создать"
            onAction={openCreate}
          />
        )}

        <div className="space-y-2">
          {events.map((ev) => (
            <ListRow
              key={ev.id}
              title={ev.title}
              subtitle={formatEventTime(ev.starts_at)}
              icon={<Clock size={18} />}
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
        title="Новое событие"
      >
        <input
          value={title}
          onChange={(e) => {
            setTitle(e.target.value)
            if (formError) setFormError(null)
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && title.trim() && !create.isPending) {
              create.mutate()
            }
          }}
          placeholder="Название"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          autoFocus
        />
        <label className="mb-3 block text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
          Время
          <input
            type="time"
            value={time}
            onChange={(e) => setTime(e.target.value)}
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
          disabled={!title.trim() || !time || create.isPending}
          onClick={() => create.mutate()}
        >
          Создать
        </Button>
      </Sheet>
    </>
  )
}

function formatEventTime(iso: string): string {
  try {
    return new Date(iso).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
  } catch {
    return iso
  }
}

/** Local wall-clock today → RFC3339 (UTC) for POST /calendar/events. */
function localTimeToRFC3339(hhmm: string): string {
  const [h, m] = hhmm.split(':').map(Number)
  const d = new Date()
  d.setHours(Number.isFinite(h) ? h : 0, Number.isFinite(m) ? m : 0, 0, 0)
  return d.toISOString()
}

function defaultTimeLocal(): string {
  const d = new Date()
  d.setMinutes(0, 0, 0)
  d.setHours(d.getHours() + 1)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(d.getHours())}:${pad(d.getMinutes())}`
}
