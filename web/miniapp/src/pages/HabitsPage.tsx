import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Check, MoreVertical, Plus, Trash2, Pencil } from 'lucide-react'
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

type HabitForm = { name: string; startDate: string; endDate: string }

const emptyForm: HabitForm = { name: '', startDate: '', endDate: '' }

function fmtDate(d?: string | null): string {
  if (!d) return ''
  const parts = d.split('-')
  if (parts.length !== 3) return d
  return `${parts[2]}.${parts[1]}.${parts[0]}`
}

export function HabitsPage() {
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [form, setForm] = useState<HabitForm>(emptyForm)
  const [formError, setFormError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const [menuHabit, setMenuHabit] = useState<HabitDay | null>(null)
  const [editHabit, setEditHabit] = useState<HabitDay | null>(null)
  const [editForm, setEditForm] = useState<HabitForm>(emptyForm)
  const [editError, setEditError] = useState<string | null>(null)
  const [confirmDelete, setConfirmDelete] = useState<HabitDay | null>(null)

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
    mutationFn: () =>
      api.createHabit({
        name: form.name.trim(),
        start_date: form.startDate || null,
        end_date: form.endDate || null,
      }),
    onSuccess: () => {
      hapticSuccess()
      void queryClient.invalidateQueries({ queryKey: ['habits'] })
      setCreateOpen(false)
      setForm(emptyForm)
      setFormError(null)
    },
    onError: (err) => {
      hapticError()
      setFormError(ruApiError(err, 'Не удалось создать привычку'))
    },
  })

  const update = useMutation({
    mutationFn: () => {
      if (!editHabit) throw new Error('no habit')
      return api.updateHabit(editHabit.id, {
        name: editForm.name.trim(),
        start_date: editForm.startDate || null,
        end_date: editForm.endDate || null,
      })
    },
    onSuccess: () => {
      hapticSuccess()
      void queryClient.invalidateQueries({ queryKey: ['habits'] })
      setEditHabit(null)
      setEditError(null)
    },
    onError: (err) => {
      hapticError()
      setEditError(ruApiError(err, 'Не удалось обновить привычку'))
    },
  })

  const remove = useMutation({
    mutationFn: (id: string) => api.deleteHabit(id),
    onSuccess: () => {
      hapticSuccess()
      void queryClient.invalidateQueries({ queryKey: ['habits'] })
      setConfirmDelete(null)
      setMenuHabit(null)
    },
    onError: (err) => {
      hapticError()
      setActionError(ruApiError(err, 'Не удалось удалить привычку'))
      setConfirmDelete(null)
    },
  })

  const habits = data ?? []
  const doneCount = habits.filter((h) => h.today_completed).length
  const trackingId = track.isPending ? track.variables : null

  const openCreate = () => {
    setFormError(null)
    setForm(emptyForm)
    setCreateOpen(true)
  }

  const openEdit = (h: HabitDay) => {
    setMenuHabit(null)
    setEditError(null)
    setEditForm({
      name: h.name,
      startDate: h.start_date ?? '',
      endDate: h.end_date ?? '',
    })
    setEditHabit(h)
  }

  const dateInputCls =
    'mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none'

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
            const inactive = h.active === false
            return (
              <div
                key={h.id}
                className={cn(
                  'flex items-stretch gap-2',
                  inactive && 'opacity-50',
                )}
              >
                <button
                  type="button"
                  disabled={h.today_completed || busy || inactive}
                  onClick={() => {
                    hapticLight()
                    track.mutate(h.id)
                  }}
                  className={cn(
                    'flex min-w-0 flex-1 items-center gap-3 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4 text-left transition active:scale-[0.99]',
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
                      {h.end_date && (
                        <>
                          {' · '}
                          {inactive ? 'срок истёк' : `до ${fmtDate(h.end_date)}`}
                        </>
                      )}
                      {!h.end_date && h.start_date && ` · с ${fmtDate(h.start_date)}`}
                    </span>
                  </span>
                </button>
                <button
                  type="button"
                  aria-label="Действия с привычкой"
                  onClick={() => {
                    setMenuHabit(h)
                  }}
                  className="flex w-12 items-center justify-center rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-[var(--tg-theme-hint-color,#94a3b8)]"
                >
                  <MoreVertical size={18} />
                </button>
              </div>
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
          value={form.name}
          onChange={(e) => {
            setForm((f) => ({ ...f, name: e.target.value }))
            if (formError) setFormError(null)
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && form.name.trim() && !create.isPending) {
              create.mutate()
            }
          }}
          placeholder="Например: Зарядка"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          autoFocus
        />
        <label className="mb-1 block text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
          Начало (необязательно)
        </label>
        <input
          type="date"
          value={form.startDate}
          onChange={(e) => setForm((f) => ({ ...f, startDate: e.target.value }))}
          className={dateInputCls}
        />
        <label className="mb-1 block text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
          Срок / окончание (необязательно)
        </label>
        <input
          type="date"
          value={form.endDate}
          onChange={(e) => setForm((f) => ({ ...f, endDate: e.target.value }))}
          className={dateInputCls}
        />
        {formError && (
          <p className="mb-3 text-sm text-rose-400" role="alert">
            {formError}
          </p>
        )}
        <Button
          className="w-full"
          disabled={!form.name.trim() || create.isPending}
          onClick={() => create.mutate()}
        >
          Создать
        </Button>
      </Sheet>

      {/* Actions menu */}
      <Sheet
        open={menuHabit !== null}
        onClose={() => setMenuHabit(null)}
        title={menuHabit?.name ?? ''}
      >
        <div className="space-y-2">
          <Button
            variant="secondary"
            className="w-full justify-start"
            onClick={() => {
              if (menuHabit) openEdit(menuHabit)
            }}
          >
            <Pencil size={16} className="mr-2" />
            Изменить
          </Button>
          <Button
            variant="secondary"
            className="w-full justify-start text-rose-400"
            onClick={() => {
              const h = menuHabit
              setMenuHabit(null)
              if (h) setConfirmDelete(h)
            }}
          >
            <Trash2 size={16} className="mr-2" />
            Удалить
          </Button>
        </div>
      </Sheet>

      {/* Edit sheet */}
      <Sheet
        open={editHabit !== null}
        onClose={() => {
          setEditHabit(null)
          setEditError(null)
        }}
        title="Изменить привычку"
      >
        <input
          value={editForm.name}
          onChange={(e) => {
            setEditForm((f) => ({ ...f, name: e.target.value }))
            if (editError) setEditError(null)
          }}
          placeholder="Название"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <label className="mb-1 block text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
          Начало
        </label>
        <input
          type="date"
          value={editForm.startDate}
          onChange={(e) => setEditForm((f) => ({ ...f, startDate: e.target.value }))}
          className={dateInputCls}
        />
        <label className="mb-1 block text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
          Срок / окончание (пусто = бессрочно)
        </label>
        <input
          type="date"
          value={editForm.endDate}
          onChange={(e) => setEditForm((f) => ({ ...f, endDate: e.target.value }))}
          className={dateInputCls}
        />
        {editError && (
          <p className="mb-3 text-sm text-rose-400" role="alert">
            {editError}
          </p>
        )}
        <Button
          className="w-full"
          disabled={!editForm.name.trim() || update.isPending}
          onClick={() => update.mutate()}
        >
          Сохранить
        </Button>
      </Sheet>

      {/* Delete confirmation */}
      <Sheet
        open={confirmDelete !== null}
        onClose={() => setConfirmDelete(null)}
        title="Удалить привычку?"
      >
        <p className="mb-4 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
          «{confirmDelete?.name}» и вся история трекинга будут удалены безвозвратно.
        </p>
        <div className="flex gap-2">
          <Button
            variant="secondary"
            className="flex-1"
            onClick={() => setConfirmDelete(null)}
          >
            Отмена
          </Button>
          <Button
            className="flex-1 bg-rose-600 text-white"
            disabled={remove.isPending}
            onClick={() => {
              if (confirmDelete) remove.mutate(confirmDelete.id)
            }}
          >
            Удалить
          </Button>
        </div>
      </Sheet>
    </>
  )
}
