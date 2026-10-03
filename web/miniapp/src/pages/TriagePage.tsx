import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { CalendarClock, ClipboardList } from 'lucide-react'
import { api } from '@/api/client'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { QueryError } from '@/components/ui/QueryError'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
import { hapticError, hapticSuccess } from '@/lib/telegram'

/**
 * MA-C6: overloaded-day triage. Shows the same proposal the bot command
 * produces (GET /planning/triage) and lets the user defer low-priority
 * tasks to tomorrow in one tap (POST /planning/triage/defer).
 */
export function TriagePage() {
  const queryClient = useQueryClient()
  const [actionError, setActionError] = useState<string | null>(null)

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['triage'],
    queryFn: () => api.triageProposal(),
  })

  const defer = useMutation({
    mutationFn: () => api.triageDefer(data?.low_priority_ids ?? []),
    onSuccess: (res) => {
      hapticSuccess()
      setActionError(null)
      void queryClient.invalidateQueries({ queryKey: ['triage'] })
      void queryClient.invalidateQueries({ queryKey: ['tasks'] })
      void queryClient.invalidateQueries({ queryKey: ['priorities'] })
      void queryClient.invalidateQueries({ queryKey: ['project-tasks'] })
      // Feedback via transient alert text is enough for a mini utility page.
      window.alert(
        res.moved > 0
          ? `Перенесено задач на завтра: ${res.moved}`
          : 'Нечего переносить — задачи уже не перегружают сегодня.',
      )
    },
    onError: (err) => {
      hapticError()
      setActionError(ruApiError(err, 'Не удалось перенести задачи'))
    },
  })

  const lowCount = data?.low_priority_ids.length ?? 0

  return (
    <>
      <Header title="Разгрузка дня" subtitle="Triage перегруженного сегодня" />
      <div className="space-y-4 px-4 pb-8">
        {isLoading && (
          <>
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-12 w-full" />
          </>
        )}

        {isError && (
          <QueryError message="Не удалось получить предложение" onRetry={() => void refetch()} />
        )}

        {data && !isLoading && (
          <>
            <div className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4">
              <p className="mb-2 flex items-center gap-2 text-sm font-medium">
                <ClipboardList size={16} className="text-[var(--tg-theme-button-color,#22c55e)]" />
                Предложение
              </p>
              <p className="whitespace-pre-wrap text-sm leading-relaxed">{data.text}</p>
            </div>

            {actionError && (
              <p className="text-sm text-rose-400" role="alert">
                {actionError}
              </p>
            )}

            {lowCount > 0 ? (
              <Button
                className="w-full"
                variant="secondary"
                disabled={defer.isPending}
                onClick={() => defer.mutate()}
              >
                <CalendarClock size={16} className="mr-2" />
                {defer.isPending ? 'Переносим…' : `Перенести низкий приоритет на завтра (${lowCount})`}
              </Button>
            ) : (
              <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
                Задач низкого приоритета на сегодня нет — переносить нечего.
              </p>
            )}
          </>
        )}
      </div>
    </>
  )
}
