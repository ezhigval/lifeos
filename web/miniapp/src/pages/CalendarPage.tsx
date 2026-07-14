import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Clock, Plus } from 'lucide-react'
import { api } from '@/api/client'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { ListRow } from '@/components/ui/ListRow'
import { hapticError, hapticSuccess } from '@/lib/telegram'

export function CalendarPage() {
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [title, setTitle] = useState('')
  const [time, setTime] = useState('12:00')

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['calendar', 'today'],
    queryFn: async () => (await api.calendarToday()).events,
  })

  const create = useMutation({
    mutationFn: () => {
      const startsAt = localTimeToRFC3339(time)
      return api.createCalendarEvent(title.trim(), startsAt)
    },
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['calendar'] })
      setCreateOpen(false)
      setTitle('')
    },
    onError: () => hapticError(),
  })

  const events = useMemo(
    () => [...(data ?? [])].sort((a, b) => a.starts_at.localeCompare(b.starts_at)),
    [data],
  )

  return (
    <>
      <Header title="Календарь" subtitle="События на сегодня" />
      <div className="space-y-4 px-4 pb-4">
        <div className="flex justify-end">
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus size={16} className="mr-1" />
            Событие
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

        {!isLoading && !isError && events.length === 0 && (
          <EmptyState
            title="На сегодня пусто"
            description="Добавь встречу или напоминание"
            actionLabel="Создать"
            onAction={() => setCreateOpen(true)}
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

      <Sheet open={createOpen} onClose={() => setCreateOpen(false)} title="Новое событие">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Название"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          autoFocus
        />
        <label className="mb-4 block text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
          Время
          <input
            type="time"
            value={time}
            onChange={(e) => setTime(e.target.value)}
            className="mt-1 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          />
        </label>
        <Button
          className="w-full"
          disabled={!title.trim() || create.isPending}
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

function localTimeToRFC3339(hhmm: string): string {
  const [h, m] = hhmm.split(':').map(Number)
  const d = new Date()
  d.setHours(h || 0, m || 0, 0, 0)
  return d.toISOString()
}
