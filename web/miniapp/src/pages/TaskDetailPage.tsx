import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate, useParams } from 'react-router-dom'
import { Archive, Trash2 } from 'lucide-react'
import { api } from '@/api/client'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { QueryError } from '@/components/ui/QueryError'
import { Skeleton } from '@/components/ui/Skeleton'
import { cn } from '@/lib/cn'
import {
  confirmAction,
  hapticError,
  hapticSuccess,
  hapticWarning,
} from '@/lib/telegram'

const PRIORITIES = [
  { value: 'urgent', label: 'Срочно' },
  { value: 'high', label: 'Высокий' },
  { value: 'medium', label: 'Средний' },
  { value: 'low', label: 'Низкий' },
] as const

export function TaskDetailPage() {
  const { taskId } = useParams<{ taskId: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['task', taskId],
    queryFn: () => api.getTask(taskId!),
    enabled: Boolean(taskId),
  })

  const [title, setTitle] = useState('')
  const [priority, setPriority] = useState('medium')
  const [dueDate, setDueDate] = useState('')
  const [description, setDescription] = useState('')
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    if (!data) return
    setTitle(data.title)
    setPriority(data.priority || 'medium')
    setDueDate(data.due_date ?? '')
    setDescription(data.description ?? '')
    setDirty(false)
  }, [data])

  const invalidateLists = () => {
    queryClient.invalidateQueries({ queryKey: ['tasks'] })
    queryClient.invalidateQueries({ queryKey: ['priorities'] })
    queryClient.invalidateQueries({ queryKey: ['project-tasks'] })
    queryClient.invalidateQueries({ queryKey: ['task', taskId] })
  }

  const save = useMutation({
    mutationFn: () => {
      const desc = description.trim()
      // EditTask API: null fields are no-ops; clearing needs clear_* flags.
      return api.updateTask(taskId!, {
        title: title.trim(),
        priority,
        ...(dueDate ? { due_date: dueDate } : { clear_due_date: true }),
        ...(desc ? { description: desc } : { clear_description: true }),
      })
    },
    onSuccess: () => {
      hapticSuccess()
      setDirty(false)
      invalidateLists()
    },
    onError: () => hapticError(),
  })

  const complete = useMutation({
    mutationFn: () => api.completeTask(taskId!),
    onSuccess: () => {
      hapticSuccess()
      invalidateLists()
      navigate(-1)
    },
    onError: () => hapticError(),
  })

  const archive = useMutation({
    mutationFn: () => api.archiveTask(taskId!),
    onSuccess: () => {
      hapticSuccess()
      invalidateLists()
      navigate(-1)
    },
    onError: () => hapticError(),
  })

  const remove = useMutation({
    mutationFn: () => api.deleteTask(taskId!),
    onSuccess: () => {
      hapticSuccess()
      invalidateLists()
      navigate(-1)
    },
    onError: () => hapticError(),
  })

  const onArchive = async () => {
    hapticWarning()
    const ok = await confirmAction('Архивировать задачу? Она пропадёт из активных списков.')
    if (ok) archive.mutate()
  }

  const onDelete = async () => {
    hapticWarning()
    const ok = await confirmAction('Удалить задачу безвозвратно?')
    if (ok) remove.mutate()
  }

  if (isLoading) {
    return (
      <>
        <Header title="Задача" subtitle="Загрузка…" />
        <div className="space-y-3 px-4">
          <Skeleton className="h-12 w-full" />
          <Skeleton className="h-24 w-full" />
        </div>
      </>
    )
  }

  if (isError || !data) {
    return (
      <>
        <Header title="Задача" />
        <div className="px-4">
          <QueryError message="Не удалось загрузить задачу" onRetry={() => void refetch()} />
        </div>
      </>
    )
  }

  const done = data.status === 'done'
  const archived = data.status === 'cancelled'

  return (
    <>
      <Header
        title={done ? 'Выполнена' : archived ? 'В архиве' : 'Задача'}
        subtitle={statusLabel(data.status)}
      />
      <div className="space-y-5 px-4 pb-8">
        <label className="block">
          <span className="mb-1.5 block text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
            Название
          </span>
          <input
            value={title}
            disabled={done || archived}
            onChange={(e) => {
              setTitle(e.target.value)
              setDirty(true)
            }}
            className="w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none disabled:opacity-60"
          />
        </label>

        <div>
          <p className="mb-2 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Приоритет</p>
          <div className="flex flex-wrap gap-2">
            {PRIORITIES.map((p) => (
              <button
                key={p.value}
                type="button"
                disabled={done || archived}
                onClick={() => {
                  setPriority(p.value)
                  setDirty(true)
                }}
                className={cn(
                  'rounded-full px-3 py-1.5 text-sm disabled:opacity-50',
                  priority === p.value
                    ? 'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]'
                    : 'bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-[var(--tg-theme-hint-color,#94a3b8)]',
                )}
              >
                {p.label}
              </button>
            ))}
          </div>
        </div>

        <label className="block">
          <span className="mb-1.5 block text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
            Срок
          </span>
          <input
            type="date"
            value={dueDate}
            disabled={done || archived}
            onChange={(e) => {
              setDueDate(e.target.value)
              setDirty(true)
            }}
            className="w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none disabled:opacity-60"
          />
        </label>

        <label className="block">
          <span className="mb-1.5 block text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
            Описание
          </span>
          <textarea
            value={description}
            disabled={done || archived}
            rows={3}
            onChange={(e) => {
              setDescription(e.target.value)
              setDirty(true)
            }}
            placeholder="Опционально"
            className="w-full resize-none rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none disabled:opacity-60"
          />
        </label>

        {!done && !archived && (
          <div className="space-y-2">
            <Button
              className="w-full"
              disabled={!dirty || !title.trim() || save.isPending}
              onClick={() => save.mutate()}
            >
              Сохранить
            </Button>
            <Button
              className="w-full"
              variant="secondary"
              disabled={complete.isPending}
              onClick={() => complete.mutate()}
            >
              Выполнить
            </Button>
          </div>
        )}

        <div className="space-y-2 border-t border-white/5 pt-4">
          {!archived && !done && (
            <Button
              className="w-full"
              variant="secondary"
              disabled={archive.isPending}
              onClick={() => void onArchive()}
            >
              <Archive size={16} className="mr-2" />
              Архивировать
            </Button>
          )}
          <Button
            className="w-full"
            variant="danger"
            disabled={remove.isPending}
            onClick={() => void onDelete()}
          >
            <Trash2 size={16} className="mr-2" />
            Удалить
          </Button>
        </div>
      </div>
    </>
  )
}

function statusLabel(status: string): string {
  switch (status) {
    case 'done':
      return 'Готово'
    case 'cancelled':
      return 'Архив'
    case 'in_progress':
      return 'В работе'
    default:
      return 'К выполнению'
  }
}
