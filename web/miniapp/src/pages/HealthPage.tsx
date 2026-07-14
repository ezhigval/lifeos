import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, ApiClientError } from '@/api/client'
import type { SleepLog, StepLog, WeightLog } from '@/api/types'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
import { formatShortDateTime } from '@/lib/datetime'
import { hapticError, hapticSuccess } from '@/lib/telegram'

type Metric = 'weight' | 'steps' | 'sleep'

async function latestOrNull<T>(fn: () => Promise<T>): Promise<T | null> {
  try {
    return await fn()
  } catch (e) {
    if (e instanceof ApiClientError && e.status === 404) return null
    throw e
  }
}

export function HealthPage() {
  const queryClient = useQueryClient()
  const [sheet, setSheet] = useState<Metric | null>(null)
  const [value, setValue] = useState('')
  const [formError, setFormError] = useState<string | null>(null)

  const weight = useQuery({
    queryKey: ['health', 'weight', 'latest'],
    queryFn: () => latestOrNull(() => api.latestWeight()),
  })
  const steps = useQuery({
    queryKey: ['health', 'steps', 'latest'],
    queryFn: () => latestOrNull(() => api.latestSteps()),
  })
  const sleep = useQuery({
    queryKey: ['health', 'sleep', 'latest'],
    queryFn: () => latestOrNull(() => api.latestSleep()),
  })

  const save = useMutation({
    mutationFn: async () => {
      const n = Number(value.replace(',', '.'))
      if (!Number.isFinite(n) || n <= 0) {
        throw new Error('invalid health value')
      }
      if (sheet === 'weight') {
        if (n < 20 || n > 400) throw new Error('invalid weight')
        return { kind: 'weight' as const, data: await api.recordWeight(n) }
      }
      if (sheet === 'steps') {
        const stepsN = Math.round(n)
        if (stepsN < 1 || stepsN > 200_000) throw new Error('invalid steps')
        return { kind: 'steps' as const, data: await api.recordSteps(stepsN) }
      }
      if (n > 24) throw new Error('invalid sleep hours')
      return { kind: 'sleep' as const, data: await api.recordSleep(n) }
    },
    onSuccess: (res) => {
      hapticSuccess()
      if (res.kind === 'weight') {
        queryClient.setQueryData<WeightLog | null>(['health', 'weight', 'latest'], res.data)
      } else if (res.kind === 'steps') {
        queryClient.setQueryData<StepLog | null>(['health', 'steps', 'latest'], res.data)
      } else {
        queryClient.setQueryData<SleepLog | null>(['health', 'sleep', 'latest'], res.data)
      }
      void queryClient.invalidateQueries({ queryKey: ['health'] })
      setSheet(null)
      setValue('')
      setFormError(null)
    },
    onError: (err) => {
      hapticError()
      const fallback =
        sheet === 'weight'
          ? 'Некорректный вес'
          : sheet === 'steps'
            ? 'Некорректное число шагов'
            : 'Некорректная длительность сна'
      setFormError(ruApiError(err, fallback))
    },
  })

  const loading = weight.isLoading || steps.isLoading || sleep.isLoading
  const hardError = weight.isError || steps.isError || sleep.isError

  const openSheet = (m: Metric) => {
    setFormError(null)
    setValue('')
    setSheet(m)
  }

  return (
    <>
      <Header title="Здоровье" subtitle="Вес, шаги, сон" />
      <div className="space-y-3 px-4 pb-6">
        {loading && <Skeleton className="h-32 w-full" />}

        {!loading && (
          <>
            <MetricCard
              title="Вес"
              value={weight.data ? `${formatWeight(weight.data.weight_kg)} кг` : 'нет данных'}
              meta={weight.data ? formatShortDateTime(weight.data.logged_at) : 'Запиши первое значение'}
              onAdd={() => openSheet('weight')}
            />
            <MetricCard
              title="Шаги"
              value={steps.data ? String(steps.data.steps) : 'нет данных'}
              meta={steps.data ? formatShortDateTime(steps.data.logged_at) : 'Запиши первое значение'}
              onAdd={() => openSheet('steps')}
            />
            <MetricCard
              title="Сон"
              value={
                sleep.data
                  ? `${formatSleepHours(sleep.data)} ч`
                  : 'нет данных'
              }
              meta={sleep.data ? formatShortDateTime(sleep.data.logged_at) : 'Запиши первое значение'}
              onAdd={() => openSheet('sleep')}
            />
          </>
        )}

        {hardError && (
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
        onClose={() => {
          setSheet(null)
          setFormError(null)
        }}
        title={
          sheet === 'weight' ? 'Вес (кг)' : sheet === 'steps' ? 'Шаги' : 'Сон (часы)'
        }
      >
        <input
          value={value}
          onChange={(e) => {
            setValue(e.target.value)
            if (formError) setFormError(null)
          }}
          inputMode="decimal"
          placeholder={sheet === 'sleep' ? '7.5' : sheet === 'steps' ? '8000' : '72.5'}
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
          disabled={!value.trim() || save.isPending}
          onClick={() => save.mutate()}
        >
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

function formatWeight(kg: number): string {
  return Number.isInteger(kg) ? String(kg) : kg.toFixed(1)
}

function formatSleepHours(log: SleepLog): string {
  const hours =
    typeof log.duration_hours === 'number' && Number.isFinite(log.duration_hours)
      ? log.duration_hours
      : (log.duration_minutes ?? 0) / 60
  return hours.toFixed(1)
}
