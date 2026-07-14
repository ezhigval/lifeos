import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { api } from '@/api/client'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatShortDateTime } from '@/lib/datetime'
import { confirmAction, hapticError, hapticSuccess, hapticWarning } from '@/lib/telegram'

export function RemindersPage() {
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [message, setMessage] = useState('')
  const [when, setWhen] = useState(() => defaultFireLocal())

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['reminders'],
    queryFn: async () => (await api.reminders()).reminders,
  })

  const create = useMutation({
    mutationFn: () => {
      const fireAt = new Date(when).toISOString()
      return api.createReminder(message.trim(), fireAt)
    },
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['reminders'] })
      setCreateOpen(false)
      setMessage('')
      setWhen(defaultFireLocal())
    },
    onError: () => hapticError(),
  })

  const cancel = useMutation({
    mutationFn: (id: string) => api.cancelReminder(id),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['reminders'] })
    },
    onError: () => hapticError(),
  })

  return (
    <>
      <Header title="Напоминания" subtitle="Отложенные пуши" />
      <div className="space-y-4 px-4 pb-6">
        <div className="flex justify-end">
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus size={16} className="mr-1" />
            Напоминание
          </Button>
        </div>

        {isLoading && <Skeleton className="h-20 w-full" />}
        {isError && (
          <QueryError message="Не удалось загрузить напоминания" onRetry={() => void refetch()} />
        )}
        {!isLoading && !isError && (data ?? []).length === 0 && (
          <EmptyState title="Нет активных напоминаний" />
        )}

        <div className="space-y-2">
          {(data ?? []).map((r) => (
            <div
              key={r.id}
              className="flex items-start gap-2 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4"
            >
              <div className="min-w-0 flex-1">
                <p className="font-medium">{r.message}</p>
                <p className="mt-1 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
                  {formatShortDateTime(r.fire_at)} · {r.status}
                </p>
              </div>
              <button
                type="button"
                className="rounded-full p-2 text-rose-400"
                aria-label="Отменить"
                onClick={async () => {
                  hapticWarning()
                  if (await confirmAction('Отменить напоминание?')) cancel.mutate(r.id)
                }}
              >
                <Trash2 size={16} />
              </button>
            </div>
          ))}
        </div>
      </div>

      <Sheet open={createOpen} onClose={() => setCreateOpen(false)} title="Новое напоминание">
        <input
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="Текст"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <input
          type="datetime-local"
          value={when}
          onChange={(e) => setWhen(e.target.value)}
          className="mb-4 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <Button
          className="w-full"
          disabled={!message.trim() || !when || create.isPending}
          onClick={() => create.mutate()}
        >
          Создать
        </Button>
      </Sheet>
    </>
  )
}

function defaultFireLocal(): string {
  const d = new Date(Date.now() + 60 * 60 * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}
