import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { Header } from '@/components/layout/Header'
import { QueryError } from '@/components/ui/QueryError'
import { Skeleton } from '@/components/ui/Skeleton'

export function AnalyticsPage() {
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['analytics', 'summary'],
    queryFn: () => api.analyticsSummary(),
  })

  return (
    <>
      <Header title="Аналитика" subtitle={data?.period_label ?? 'Сводка'} />
      <div className="space-y-4 px-4 pb-6">
        {isLoading && <Skeleton className="h-40 w-full" />}
        {isError && (
          <QueryError message="Не удалось загрузить аналитику" onRetry={() => void refetch()} />
        )}
        {data && (
          <>
            <div className="grid grid-cols-2 gap-3">
              <Stat label="Создано" value={String(data.tasks_created)} />
              <Stat label="Сделано" value={String(data.tasks_completed)} />
              <Stat label="Открыто" value={String(data.open_tasks)} />
              <Stat
                label="Completion"
                value={`${Math.round((data.completion_rate || 0) * 100)}%`}
              />
              <Stat label="Привычки" value={`${data.habit_completions}/${data.habit_count}`} />
              <Stat
                label="Консистентность"
                value={`${Math.round((data.habit_consistency || 0) * 100)}%`}
              />
            </div>

            <section>
              <h2 className="mb-2 text-sm font-medium text-[var(--tg-theme-hint-color,#94a3b8)]">
                Проекты
              </h2>
              <div className="space-y-2">
                {(data.projects ?? []).length === 0 && (
                  <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Нет данных</p>
                )}
                {(data.projects ?? []).map((p, i) => (
                  <div
                    key={i}
                    className="flex items-center justify-between rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3"
                  >
                    <span className="truncate font-medium">{p.Title || p.title || 'Проект'}</span>
                    <span className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
                      {p.Percent || p.percent || '—'}
                    </span>
                  </div>
                ))}
              </div>
            </section>
          </>
        )}
      </div>
    </>
  )
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4">
      <div className="text-2xl font-bold tabular-nums">{value}</div>
      <div className="mt-1 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">{label}</div>
    </div>
  )
}
