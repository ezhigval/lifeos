import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { AnalyticsSummary } from '@/api/types'
import { Header } from '@/components/layout/Header'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Skeleton } from '@/components/ui/Skeleton'

export function AnalyticsPage() {
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['analytics', 'summary'],
    queryFn: () => api.analyticsSummary(),
  })

  const projects = normalizeProjects(data)

  return (
    <>
      <Header title="Аналитика" subtitle={data?.period_label ?? 'Сводка периода'} />
      <div className="space-y-4 px-4 pb-6">
        {isLoading && (
          <div className="space-y-3">
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-40 w-full" />
          </div>
        )}

        {isError && (
          <QueryError message="Не удалось загрузить аналитику" onRetry={() => void refetch()} />
        )}

        {!isLoading && !isError && !data && (
          <EmptyState title="Нет сводки" description="Попробуй обновить позже" />
        )}

        {data && (
          <>
            <div className="grid grid-cols-2 gap-3">
              <Stat label="Создано задач" value={String(data.tasks_created ?? 0)} />
              <Stat label="Сделано" value={String(data.tasks_completed ?? 0)} />
              <Stat label="Открыто" value={String(data.open_tasks ?? 0)} />
              <Stat label="Выполнение" value={`${formatPercent(data.completion_rate)}%`} />
              <Stat
                label="Привычки"
                value={`${data.habit_completions ?? 0}/${data.habit_count ?? 0}`}
              />
              <Stat
                label="Консистентность"
                value={`${formatPercent(data.habit_consistency)}%`}
              />
            </div>

            <section>
              <h2 className="mb-2 text-sm font-medium text-[var(--tg-theme-hint-color,#94a3b8)]">
                Проекты
              </h2>
              <div className="space-y-2">
                {projects.length === 0 && (
                  <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
                    Нет активных проектов
                  </p>
                )}
                {projects.map((p, i) => (
                  <div
                    key={`${p.title}-${i}`}
                    className="flex items-center justify-between rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3"
                  >
                    <span className="truncate font-medium">{p.title}</span>
                    <span className="shrink-0 text-sm tabular-nums text-[var(--tg-theme-hint-color,#94a3b8)]">
                      {p.percent}
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

/** API returns integers 0–100; defend against a 0–1 fraction if contract drifts. */
function formatPercent(raw: number | undefined): string {
  const n = Number(raw) || 0
  const pct = n > 0 && n <= 1 ? n * 100 : n
  return String(Math.round(pct))
}

/**
 * Live handler marshals Go ProjectKPI without json tags → Title/Percent.
 * Tolerate snake/lower if OpenAPI or future tags land.
 */
function normalizeProjects(
  data: AnalyticsSummary | undefined,
): { title: string; percent: string }[] {
  const raw = data?.projects
  if (!Array.isArray(raw)) return []
  return raw.map((p) => ({
    title: p.title || p.Title || 'Проект',
    percent: p.percent || p.Percent || '—',
  }))
}
