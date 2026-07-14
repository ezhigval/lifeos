import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { api } from '@/api/client'
import type { Reminder } from '@/api/types'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
import { formatShortDateTime } from '@/lib/datetime'
import { confirmAction, hapticError, hapticSuccess, hapticWarning } from '@/lib/telegram'

export function RemindersPage() {
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [message, setMessage] = useState('')
  const [when, setWhen] = useState(() => defaultFireLocal())
  const [formError, setFormError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['reminders'],
    queryFn: async () => {
      const res = await api.reminders()
      return Array.isArray(res.reminders) ? res.reminders : []
    },
  })

  const reminders = data ?? []

  const create = useMutation({
    mutationFn: () => {
      const fireAt = localDateTimeToISO(when)
      if (!fireAt) throw new Error('invalid fire_at, use RFC3339')
      if (new Date(fireAt).getTime() <= Date.now()) {
        throw new Error('fire_at must be in the future')
      }
      return api.createReminder(message.trim(), fireAt)
    },
    onSuccess: (reminder) => {
      hapticSuccess()
      // POST /reminders returns full Reminder (id/status) — insert before refetch.
      if (reminder?.id) {
        queryClient.setQueryData<Reminder[]>(['reminders'], (old) => {
          const list = old ?? []
          if (list.some((r) => r.id === reminder.id)) return list
          return [reminder, ...list]
        })
      }
      void queryClient.invalidateQueries({ queryKey: ['reminders'] })
      setCreateOpen(false)
      setMessage('')
      setWhen(defaultFireLocal())
      setFormError(null)
    },
    onError: (err) => {
      hapticError()
      setFormError(ruApiError(err, 'Не удалось создать напоминание'))
    },
  })

  const cancel = useMutation({
    mutationFn: (id: string) => api.cancelReminder(id),
    onMutate: async (id) => {
      setActionError(null)
      await queryClient.cancelQueries({ queryKey: ['reminders'] })
      const prev = queryClient.getQueryData<Reminder[]>(['reminders'])
      queryClient.setQueryData<Reminder[]>(['reminders'], (old) =>
        (old ?? []).filter((r) => r.id !== id),
      )
      return { prev }
    },
    onSuccess: () => hapticSuccess(),
    onError: (err, _id, ctx) => {
      hapticError()
      if (ctx?.prev) queryClient.setQueryData(['reminders'], ctx.prev)
      setActionError(ruApiError(err, 'Не удалось отменить'))
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: ['reminders'] })
    },
  })

  const openCreate = () => {
    setFormError(null)
    setMessage('')
    setWhen(defaultFireLocal())
    setCreateOpen(true)
  }

  const cancellingId = cancel.isPending ? cancel.variables : null

  return (
    <>
      <Header title="Напоминания" subtitle="Отложенные пуши" />
      <div className="space-y-4 px-4 pb-6">
        {reminders.length > 0 && (
          <div className="flex justify-end">
            <Button size="sm" onClick={openCreate}>
              <Plus size={16} className="mr-1" />
              Напоминание
            </Button>
          </div>
        )}

        {actionError && (
          <p className="text-sm text-rose-400" role="alert">
            {actionError}
          </p>
        )}

        {isLoading && (
          <div className="space-y-2">
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-20 w-full" />
          </div>
        )}

        {isError && (
          <QueryError message="Не удалось загрузить напоминания" onRetry={() => void refetch()} />
        )}

        {!isLoading && !isError && reminders.length === 0 && (
          <EmptyState
            title="Нет активных напоминаний"
            description="Задай время — бот пришлёт пуш"
            actionLabel="Создать"
            onAction={openCreate}
          />
        )}

        {!isLoading && !isError && reminders.length > 0 && (
          <div className="space-y-2">
            {reminders.map((r) => (
              <div
                key={r.id}
                className="flex items-start gap-2 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4"
              >
                <div className="min-w-0 flex-1">
                  <p className="font-medium">{r.message}</p>
                  <p className="mt-1 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
                    {formatShortDateTime(r.fire_at)} · {ruReminderStatus(r.status)}
                  </p>
                </div>
                <button
                  type="button"
                  className="rounded-full p-2 text-rose-400 disabled:opacity-50"
                  aria-label="Отменить"
                  disabled={cancellingId === r.id}
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
        )}
      </div>

      <Sheet
        open={createOpen}
        onClose={() => {
          setCreateOpen(false)
          setFormError(null)
        }}
        title="Новое напоминание"
      >
        <input
          value={message}
          onChange={(e) => {
            setMessage(e.target.value)
            if (formError) setFormError(null)
          }}
          placeholder="Текст"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          autoFocus
        />
        <label className="mb-1 block text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
          Когда
        </label>
        <input
          type="datetime-local"
          value={when}
          onChange={(e) => {
            setWhen(e.target.value)
            if (formError) setFormError(null)
          }}
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        {formError && (
          <p className="mb-3 text-sm text-rose-400" role="alert">
            {formError}
          </p>
        )}
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
  return toLocalInput(d)
}

/** datetime-local value → RFC3339 Instant (browser local wall clock). */
function localDateTimeToISO(value: string): string | null {
  if (!value) return null
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return null
  return d.toISOString()
}

function toLocalInput(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function ruReminderStatus(status: string): string {
  const s = (status || '').toLowerCase()
  if (s === 'pending' || s === 'scheduled') return 'ожидает'
  if (s === 'cancelled' || s === 'canceled') return 'отменено'
  if (s === 'done' || s === 'sent' || s === 'completed') return 'отправлено'
  return status || '—'
}
