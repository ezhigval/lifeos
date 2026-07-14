import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Check, Plus } from 'lucide-react'
import { api } from '@/api/client'
import type { HabitDay } from '@/api/types'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
import { cn } from '@/lib/cn'
import { hapticError, hapticLight, hapticSuccess } from '@/lib/telegram'

export function HabitsPage() {
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [name, setName] = useState('')
  const [formError, setFormError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['habits', 'today'],
    queryFn: async () => {
      const res = await api.habitsToday()
      return Array.isArray(res.habits) ? res.habits : []
    },
  })

  const track = useMutation({
    mutationFn: (id: string) => api.trackHabit(id),
    onMutate: async (id) => {
      setActionError(null)
      await queryClient.cancelQueries({ queryKey: ['habits', 'today'] })
      const prev = queryClient.getQueryData<HabitDay[]>(['habits', 'today'])
      queryClient.setQueryData<HabitDay[]>(['habits', 'today'], (old) =>
        (old ?? []).map((h) =>
          h.id === id && !h.today_completed
            ? { ...h, today_completed: true, streak: h.streak + 1 }
            : h,
        ),
      )
      return { prev }
    },
    onSuccess: (res, id) => {
      hapticSuccess()
      queryClient.setQueryData<HabitDay[]>(['habits', 'today'], (old) =>
        (old ?? []).map((h) =>
          h.id === id ? { ...h, today_completed: true, streak: res.streak ?? h.streak } : h,
        ),
      )
    },
    onError: (err, _id, ctx) => {
      hapticError()
      if (ctx?.prev) queryClient.setQueryData(['habits', 'today'], ctx.prev)
      setActionError(ruApiError(err, 'Не удалось отметить привычку'))
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: ['habits'] })
    },
  })

  const create = useMutation({
    mutationFn: () => api.createHabit(name.trim()),
    onSuccess: () => {
      hapticSuccess()
      void queryClient.invalidateQueries({ queryKey: ['habits'] })
      setCreateOpen(false)
      setName('')
      setFormError(null)
    },
    onError: (err) => {
      hapticError()
      setFormError(ruApiError(err, 'Не удалось создать привычку'))
    },
  })

  const habits = data ?? []
  const doneCount = habits.filter((h) => h.today_completed).length
  const trackingId = track.isPending ? track.variables : null

  const openCreate = () => {
    setFormError(null)
    setName('')
    setCreateOpen(true)
  }

  return (
    <>
      <Header
        title="Привычки"
        subtitle={
          habits.length
            ? `Сегодня ${doneCount}/${habits.length}`
            : 'Трекер привычек'
        }
      />
      <div className="space-y-4 px-4 pb-4">
        {habits.length > 0 && (
          <div className="flex justify-end">
            <Button size="sm" onClick={openCreate}>
              <Plus size={16} className="mr-1" />
              Привычка
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
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
        )}

        {isError && (
          <QueryError message="Не удалось загрузить привычки" onRetry={() => void refetch()} />
        )}

        {!isLoading && !isError && habits.length === 0 && (
          <EmptyState
            title="Пока нет привычек"
            description="Добавь первую — один тап в день"
            actionLabel="Создать"
            onAction={openCreate}
          />
        )}

        <div className="space-y-2">
          {habits.map((h) => {
            const busy = trackingId === h.id
            return (
              <button
                key={h.id}
                type="button"
                disabled={h.today_completed || busy}
                onClick={() => {
                  hapticLight()
                  track.mutate(h.id)
                }}
                className={cn(
                  'flex w-full items-center gap-3 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4 text-left transition active:scale-[0.99]',
                  h.today_completed && 'opacity-70',
                )}
              >
                <span
                  className={cn(
                    'flex h-8 w-8 items-center justify-center rounded-full border-2 transition-colors duration-200',
                    h.today_completed
                      ? 'border-emerald-500 bg-emerald-500 text-white'
                      : 'border-[var(--tg-theme-hint-color,#64748b)]',
                  )}
                >
                  {h.today_completed && <Check size={16} />}
                </span>
                <span className="min-w-0 flex-1">
                  <span
                    className={cn(
                      'block font-medium transition-all duration-200',
                      h.today_completed && 'line-through',
                    )}
                  >
                    {h.name}
                  </span>
                  <span className="text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
                    серия {h.streak} дн.
                  </span>
                </span>
              </button>
            )
          })}
        </div>
      </div>

      <Sheet
        open={createOpen}
        onClose={() => {
          setCreateOpen(false)
          setFormError(null)
        }}
        title="Новая привычка"
      >
        <input
          value={name}
          onChange={(e) => {
            setName(e.target.value)
            if (formError) setFormError(null)
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && name.trim() && !create.isPending) {
              create.mutate()
            }
          }}
          placeholder="Например: Зарядка"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          autoFocus
        />
        {formError && (
          <p className="mb-3 text-sm text-rose-400" role="alert">
            {formError}
          </p>
        )}
        <Button
          className="w-full"
          disabled={!name.trim() || create.isPending}
          onClick={() => create.mutate()}
        >
          Создать
        </Button>
      </Sheet>
    </>
  )
}
