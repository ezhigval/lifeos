import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '@/api/client'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Skeleton } from '@/components/ui/Skeleton'
import { cn } from '@/lib/cn'
import { ruApiError } from '@/lib/apiError'
import { hapticError, hapticLight, hapticSuccess } from '@/lib/telegram'
import { useState } from 'react'

/** Today’s habit checkboxes only — full tracker lives under More → Привычки. */
export function HomeHabits() {
  const queryClient = useQueryClient()
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
    onSuccess: () => {
      hapticSuccess()
      setActionError(null)
      void queryClient.invalidateQueries({ queryKey: ['habits'] })
    },
    onError: (err) => {
      hapticError()
      setActionError(ruApiError(err, 'Не удалось отметить'))
    },
  })

  const habits = data ?? []

  return (
    <section>
      <div className="mb-3 flex items-center justify-between px-4">
        <h2 className="text-base font-semibold">Привычки сегодня</h2>
        <Link to="/more/habits" className="text-sm text-[var(--tg-theme-link-color,#22c55e)]">
          Трекер →
        </Link>
      </div>
      <div className="space-y-2 px-4">
        {isLoading && <Skeleton className="h-12 w-full" />}
        {isError && (
          <QueryError message="Не удалось загрузить привычки" onRetry={() => void refetch()} />
        )}
        {actionError && (
          <p className="text-sm text-rose-400" role="alert">
            {actionError}
          </p>
        )}
        {!isLoading && !isError && habits.length === 0 && (
          <EmptyState title="Нет привычек" description="Добавь в разделе трекера" />
        )}
        {habits.map((h) => (
          <label
            key={h.id}
            className="flex cursor-pointer items-center gap-3 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-3 py-3"
          >
            <input
              type="checkbox"
              className="h-5 w-5 rounded accent-emerald-500"
              checked={h.today_completed}
              disabled={h.today_completed || track.isPending}
              onChange={() => {
                if (h.today_completed) return
                hapticLight()
                track.mutate(h.id)
              }}
            />
            <span className={cn('flex-1 font-medium', h.today_completed && 'line-through opacity-60')}>
              {h.name}
            </span>
            <span className="text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">🔥 {h.streak}</span>
          </label>
        ))}
      </div>
    </section>
  )
}
