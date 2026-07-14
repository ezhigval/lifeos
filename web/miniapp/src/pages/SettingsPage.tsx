import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { api } from '@/api/client'
import type { Sphere } from '@/api/types'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatTimeOfDay, inputToTimeOfDay, timeOfDayToInput } from '@/lib/datetime'
import {
  confirmAction,
  hapticError,
  hapticSuccess,
  hapticWarning,
} from '@/lib/telegram'

export function SettingsPage() {
  const queryClient = useQueryClient()
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['settings'],
    queryFn: () => api.settings(),
  })

  const [morning, setMorning] = useState('08:00')
  const [evening, setEvening] = useState('21:00')
  const [quietStart, setQuietStart] = useState('23:00')
  const [quietEnd, setQuietEnd] = useState('07:00')

  useEffect(() => {
    if (!data) return
    setMorning(timeOfDayToInput(data.morning_review_at))
    setEvening(timeOfDayToInput(data.evening_review_at))
    setQuietStart(timeOfDayToInput(data.quiet_hours_start ?? { hour: 23, minute: 0 }))
    setQuietEnd(timeOfDayToInput(data.quiet_hours_end ?? { hour: 7, minute: 0 }))
  }, [data])

  const saveMorning = useMutation({
    mutationFn: () => {
      const t = inputToTimeOfDay(morning)
      return api.updateMorningReview(t.hour, t.minute)
    },
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
    onError: () => hapticError(),
  })

  const saveEvening = useMutation({
    mutationFn: () => {
      const t = inputToTimeOfDay(evening)
      return api.updateEveningReview(t.hour, t.minute)
    },
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
    onError: () => hapticError(),
  })

  const saveQuiet = useMutation({
    mutationFn: () => {
      const s = inputToTimeOfDay(quietStart)
      const e = inputToTimeOfDay(quietEnd)
      return api.updateQuietHours(s.hour, s.minute, e.hour, e.minute)
    },
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
    onError: () => hapticError(),
  })

  return (
    <>
      <Header title="Настройки" subtitle="Обзоры, quiet hours, сферы" />
      <div className="space-y-6 px-4 pb-8">
        {isLoading && <Skeleton className="h-40 w-full" />}
        {isError && (
          <QueryError message="Не удалось загрузить настройки" onRetry={() => void refetch()} />
        )}

        {data && (
          <>
            <section className="space-y-3 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4">
              <h2 className="font-semibold">Утренний обзор</h2>
              <p className="text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
                Сейчас: {formatTimeOfDay(data.morning_review_at)}
              </p>
              <input
                type="time"
                value={morning}
                onChange={(e) => setMorning(e.target.value)}
                className="w-full rounded-2xl bg-black/20 px-4 py-3 outline-none"
              />
              <Button
                className="w-full"
                size="sm"
                disabled={saveMorning.isPending}
                onClick={() => saveMorning.mutate()}
              >
                Сохранить утро
              </Button>
            </section>

            <section className="space-y-3 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4">
              <h2 className="font-semibold">Вечерний обзор</h2>
              <p className="text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
                Сейчас: {formatTimeOfDay(data.evening_review_at)}
              </p>
              <input
                type="time"
                value={evening}
                onChange={(e) => setEvening(e.target.value)}
                className="w-full rounded-2xl bg-black/20 px-4 py-3 outline-none"
              />
              <Button
                className="w-full"
                size="sm"
                disabled={saveEvening.isPending}
                onClick={() => saveEvening.mutate()}
              >
                Сохранить вечер
              </Button>
            </section>

            <section className="space-y-3 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4">
              <h2 className="font-semibold">Quiet hours</h2>
              <p className="text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
                {data.quiet_hours_start
                  ? `${formatTimeOfDay(data.quiet_hours_start)} – ${formatTimeOfDay(data.quiet_hours_end)}`
                  : 'Не заданы'}
              </p>
              <div className="grid grid-cols-2 gap-2">
                <label className="text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
                  С
                  <input
                    type="time"
                    value={quietStart}
                    onChange={(e) => setQuietStart(e.target.value)}
                    className="mt-1 w-full rounded-2xl bg-black/20 px-3 py-3 outline-none"
                  />
                </label>
                <label className="text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
                  До
                  <input
                    type="time"
                    value={quietEnd}
                    onChange={(e) => setQuietEnd(e.target.value)}
                    className="mt-1 w-full rounded-2xl bg-black/20 px-3 py-3 outline-none"
                  />
                </label>
              </div>
              <Button
                className="w-full"
                size="sm"
                disabled={saveQuiet.isPending}
                onClick={() => saveQuiet.mutate()}
              >
                Сохранить quiet hours
              </Button>
            </section>
          </>
        )}

        <SpheresSettings />
      </div>
    </>
  )
}

function SpheresSettings() {
  const queryClient = useQueryClient()
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['spheres'],
    queryFn: async () => (await api.spheres()).spheres,
  })

  const [createOpen, setCreateOpen] = useState(false)
  const [edit, setEdit] = useState<Sphere | null>(null)
  const [name, setName] = useState('')

  const create = useMutation({
    mutationFn: () => api.createSphere(name.trim()),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['spheres'] })
      setCreateOpen(false)
      setName('')
    },
    onError: () => hapticError(),
  })

  const update = useMutation({
    mutationFn: () => api.updateSphere(edit!.id, name.trim(), edit!.sort_order),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['spheres'] })
      setEdit(null)
      setName('')
    },
    onError: () => hapticError(),
  })

  const remove = useMutation({
    mutationFn: (id: string) => api.deleteSphere(id),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['spheres'] })
    },
    onError: () => hapticError(),
  })

  const onDelete = async (s: Sphere) => {
    hapticWarning()
    const ok = await confirmAction(`Удалить сферу «${s.name}»?`)
    if (ok) remove.mutate(s.id)
  }

  return (
    <section className="space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="font-semibold">Сферы жизни</h2>
        <Button
          size="sm"
          onClick={() => {
            setName('')
            setCreateOpen(true)
          }}
        >
          <Plus size={16} className="mr-1" />
          Сфера
        </Button>
      </div>

      {isLoading && <Skeleton className="h-16 w-full" />}
      {isError && <QueryError message="Не удалось загрузить сферы" onRetry={() => void refetch()} />}

      {!isLoading && !isError && (data ?? []).length === 0 && (
        <EmptyState title="Сфер нет" description="Создай первую" />
      )}

      <div className="space-y-2">
        {(data ?? []).map((s) => (
          <div
            key={s.id}
            className="flex items-center gap-2 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-3 py-3"
          >
            <span className="min-w-0 flex-1 truncate font-medium">{s.name}</span>
            <button
              type="button"
              className="rounded-full p-2 text-[var(--tg-theme-hint-color,#94a3b8)]"
              aria-label="Изменить"
              onClick={() => {
                setEdit(s)
                setName(s.name)
              }}
            >
              <Pencil size={16} />
            </button>
            <button
              type="button"
              className="rounded-full p-2 text-rose-400"
              aria-label="Удалить"
              onClick={() => void onDelete(s)}
            >
              <Trash2 size={16} />
            </button>
          </div>
        ))}
      </div>

      <Sheet open={createOpen} onClose={() => setCreateOpen(false)} title="Новая сфера">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Название"
          className="mb-4 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <Button
          className="w-full"
          disabled={!name.trim() || create.isPending}
          onClick={() => create.mutate()}
        >
          Создать
        </Button>
      </Sheet>

      <Sheet open={Boolean(edit)} onClose={() => setEdit(null)} title="Редактировать сферу">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="mb-4 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <Button
          className="w-full"
          disabled={!name.trim() || update.isPending}
          onClick={() => update.mutate()}
        >
          Сохранить
        </Button>
      </Sheet>
    </section>
  )
}
