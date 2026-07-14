import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, ApiClientError } from '@/api/client'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatShortDateTime } from '@/lib/datetime'
import { hapticError, hapticSuccess } from '@/lib/telegram'

export function HealthPage() {
  const queryClient = useQueryClient()
  const [sheet, setSheet] = useState<'weight' | 'steps' | 'sleep' | null>(null)
  const [value, setValue] = useState('')

  const weight = useQuery({
    queryKey: ['health', 'weight', 'latest'],
    queryFn: () => api.latestWeight(),
    retry: false,
  })
  const steps = useQuery({
    queryKey: ['health', 'steps', 'latest'],
    queryFn: () => api.latestSteps(),
    retry: false,
  })
  const sleep = useQuery({
    queryKey: ['health', 'sleep', 'latest'],
    queryFn: () => api.latestSleep(),
    retry: false,
  })

  const save = useMutation({
    mutationFn: async () => {
      const n = Number(value.replace(',', '.'))
      if (!Number.isFinite(n) || n <= 0) throw new Error('invalid')
      if (sheet === 'weight') return api.recordWeight(n)
      if (sheet === 'steps') return api.recordSteps(Math.round(n))
      return api.recordSleep(n)
    },
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['health'] })
      setSheet(null)
      setValue('')
    },
    onError: () => hapticError(),
  })

  const loading = weight.isLoading || steps.isLoading || sleep.isLoading

  return (
    <>
      <Header title="Здоровье" subtitle="Вес, шаги, сон" />
      <div className="space-y-3 px-4 pb-6">
        {loading && <Skeleton className="h-32 w-full" />}

        <MetricCard
          title="Вес"
          value={
            weight.isError
              ? 'нет данных'
              : weight.data
                ? `${weight.data.weight_kg} кг`
                : '…'
          }
          meta={weight.data ? formatShortDateTime(weight.data.logged_at) : undefined}
          onAdd={() => {
            setValue('')
            setSheet('weight')
          }}
        />
        <MetricCard
          title="Шаги"
          value={
            steps.isError ? 'нет данных' : steps.data ? String(steps.data.steps) : '…'
          }
          meta={steps.data ? formatShortDateTime(steps.data.logged_at) : undefined}
          onAdd={() => {
            setValue('')
            setSheet('steps')
          }}
        />
        <MetricCard
          title="Сон"
          value={
            sleep.isError
              ? 'нет данных'
              : sleep.data
                ? `${sleep.data.duration_hours.toFixed(1)} ч`
                : '…'
          }
          meta={sleep.data ? formatShortDateTime(sleep.data.logged_at) : undefined}
          onAdd={() => {
            setValue('')
            setSheet('sleep')
          }}
        />

        {(weight.isError || steps.isError || sleep.isError) &&
          ![weight.error, steps.error, sleep.error].every(
            (e) => e instanceof ApiClientError && e.status === 404,
          ) && (
            <QueryError
              message="Ошибка загрузки"
              onRetry={() => {
                void weight.refetch()
                void steps.refetch()
                void sleep.refetch()
              }}
            />
          )}
      </div>

      <Sheet
        open={sheet !== null}
        onClose={() => setSheet(null)}
        title={
          sheet === 'weight' ? 'Вес (кг)' : sheet === 'steps' ? 'Шаги' : 'Сон (часы)'
        }
      >
        <input
          value={value}
          onChange={(e) => setValue(e.target.value)}
          inputMode="decimal"
          placeholder={sheet === 'sleep' ? '7.5' : sheet === 'steps' ? '8000' : '72.5'}
          className="mb-4 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <Button className="w-full" disabled={!value.trim() || save.isPending} onClick={() => save.mutate()}>
          Записать
        </Button>
      </Sheet>
    </>
  )
}

function MetricCard({
  title,
  value,
  meta,
  onAdd,
}: {
  title: string
  value: string
  meta?: string
  onAdd: () => void
}) {
  return (
    <div className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">{title}</p>
          <p className="mt-1 text-2xl font-bold tabular-nums">{value}</p>
          {meta && <p className="mt-1 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">{meta}</p>}
        </div>
        <Button size="sm" variant="secondary" onClick={onAdd}>
          +
        </Button>
      </div>
    </div>
  )
}
