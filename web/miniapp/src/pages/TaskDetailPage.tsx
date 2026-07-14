import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate, useParams } from 'react-router-dom'
import { Archive, Trash2 } from 'lucide-react'
import { api } from '@/api/client'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { QueryError } from '@/components/ui/QueryError'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
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

const KINDS = [
  { value: 'task' as const, label: 'Задача' },
  { value: 'reminder' as const, label: 'Напоминание' },
  { value: 'meeting' as const, label: 'Встреча' },
]

export function TaskDetailPage() {
  const { taskId } = useParams<{ taskId: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [actionError, setActionError] = useState<string | null>(null)

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['task', taskId],
    queryFn: () => api.getTask(taskId!),
    enabled: Boolean(taskId),
  })

  const { data: linkedNote } = useQuery({
    queryKey: ['note', data?.note_id],
    queryFn: () => api.getNote(data!.note_id!),
    enabled: Boolean(data?.note_id),
  })

  const [title, setTitle] = useState('')
  const [priority, setPriority] = useState('medium')
  const [kind, setKind] = useState<'task' | 'reminder' | 'meeting'>('task')
  const [dueDate, setDueDate] = useState('')
  const [address, setAddress] = useState('')
  const [description, setDescription] = useState('')
  const [noteText, setNoteText] = useState('')
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    if (!data) return
    setTitle(data.title)
    setPriority(data.priority || 'medium')
    setKind(data.kind || 'task')
    setDueDate(data.due_date ?? '')
    setAddress(data.address ?? '')
    setDescription(data.description ?? '')
    setNoteText('')
    setDirty(false)
  }, [data])

  useEffect(() => {
    if (linkedNote) setNoteText(linkedNote.body)
  }, [linkedNote])

  const invalidateLists = () => {
    queryClient.invalidateQueries({ queryKey: ['tasks'] })
    queryClient.invalidateQueries({ queryKey: ['priorities'] })
    queryClient.invalidateQueries({ queryKey: ['project-tasks'] })
    queryClient.invalidateQueries({ queryKey: ['task', taskId] })
  }

  const save = useMutation({
    mutationFn: async () => {
      const desc = description.trim()
      const addr = address.trim()
      const note = noteText.trim()
      let noteId = data?.note_id

      if (note) {
        if (noteId) {
          await api.updateNote(noteId, { body: note })
        } else {
          const created = await api.createNote(note)
          noteId = created.id
        }
      }

      return api.updateTask(taskId!, {
        title: title.trim(),
        priority,
        kind,
        ...(dueDate ? { due_date: dueDate } : { clear_due_date: true }),
        ...(addr ? { address: addr } : { clear_address: true }),
        ...(note && noteId ? { note_id: noteId } : {}),
        ...(!note && data?.note_id ? { clear_note_id: true } : {}),
        ...(desc ? { description: desc } : { clear_description: true }),
      })
    },
    onSuccess: () => {
      hapticSuccess()
      setDirty(false)
      setActionError(null)
      invalidateLists()
      if (data?.note_id || noteText.trim()) {
        queryClient.invalidateQueries({ queryKey: ['note'] })
      }
    },
    onError: (err) => {
      hapticError()
      setActionError(ruApiError(err, 'Не удалось сохранить'))
    },
  })

  const complete = useMutation({
    mutationFn: () => api.completeTask(taskId!),
    onSuccess: () => {
      hapticSuccess()
      setActionError(null)
      invalidateLists()
      navigate(-1)
    },
    onError: (err) => {
      hapticError()
      setActionError(ruApiError(err, 'Не удалось выполнить'))
    },
  })

  const reopen = useMutation({
    mutationFn: () => api.reopenTask(taskId!),
    onSuccess: () => {
      hapticSuccess()
      setActionError(null)
      invalidateLists()
    },
    onError: (err) => {
      hapticError()
      setActionError(ruApiError(err, 'Не удалось вернуть задачу'))
    },
  })

  const archive = useMutation({
    mutationFn: () => api.archiveTask(taskId!),
    onSuccess: () => {
      hapticSuccess()
      setActionError(null)
      invalidateLists()
      navigate(-1)
    },
    onError: (err) => {
      hapticError()
      setActionError(ruApiError(err, 'Не удалось архивировать'))
    },
  })

  const remove = useMutation({
    mutationFn: () => api.deleteTask(taskId!),
    onSuccess: () => {
      hapticSuccess()
      setActionError(null)
      invalidateLists()
      navigate(-1)
    },
    onError: (err) => {
      hapticError()
      setActionError(ruApiError(err, 'Не удалось удалить'))
    },
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

  const markDirty = () => {
    setDirty(true)
    if (actionError) setActionError(null)
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
  const readOnly = done || archived

  return (
    <>
      <Header
        title={done ? 'Выполнена' : archived ? 'В архиве' : 'Задача'}
        subtitle={statusLabel(data.status)}
      />
      <div className="space-y-5 px-4 pb-8">
        {actionError && (
          <p className="text-sm text-rose-400" role="alert">
            {actionError}
          </p>
        )}

        <label className="block">
          <span className="mb-1.5 block text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
            Название
          </span>
          <input
            value={title}
            disabled={readOnly}
            onChange={(e) => {
              setTitle(e.target.value)
              markDirty()
            }}
            className="w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none disabled:opacity-60"
          />
        </label>

        <div>
          <p className="mb-2 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Тип</p>
          <div className="flex flex-wrap gap-2">
            {KINDS.map((k) => (
              <button
                key={k.value}
                type="button"
                disabled={readOnly}
                onClick={() => {
                  setKind(k.value)
                  markDirty()
                }}
                className={cn(
                  'rounded-full px-3 py-1.5 text-sm disabled:opacity-50',
                  kind === k.value
                    ? 'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]'
                    : 'bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-[var(--tg-theme-hint-color,#94a3b8)]',
                )}
              >
                {k.label}
              </button>
            ))}
          </div>
          {kind === 'reminder' && (
            <p className="mt-2 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
              Напоминание — push-уведомление в указанное время
            </p>
          )}
        </div>

        <div>
          <p className="mb-2 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Приоритет</p>
          <div className="flex flex-wrap gap-2">
            {PRIORITIES.map((p) => (
              <button
                key={p.value}
                type="button"
                disabled={readOnly}
                onClick={() => {
                  setPriority(p.value)
                  markDirty()
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
            disabled={readOnly}
            onChange={(e) => {
              setDueDate(e.target.value)
              markDirty()
            }}
            className="w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none disabled:opacity-60"
          />
        </label>

        <label className="block">
          <span className="mb-1.5 block text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
            Адрес
          </span>
          <input
            value={address}
            disabled={readOnly}
            onChange={(e) => {
              setAddress(e.target.value)
              markDirty()
            }}
            placeholder="Опционально"
            className="w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none disabled:opacity-60"
          />
        </label>

        <label className="block">
          <span className="mb-1.5 block text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
            Описание
          </span>
          <textarea
            value={description}
            disabled={readOnly}
            rows={3}
            onChange={(e) => {
              setDescription(e.target.value)
              markDirty()
            }}
            placeholder="Опционально"
            className="w-full resize-none rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none disabled:opacity-60"
          />
        </label>

        <label className="block">
          <span className="mb-1.5 block text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
            Заметка
          </span>
          <textarea
            value={noteText}
            disabled={readOnly}
            rows={4}
            onChange={(e) => {
              setNoteText(e.target.value)
              markDirty()
            }}
            placeholder="Текст заметки, привязанной к задаче"
            className="w-full resize-none rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none disabled:opacity-60"
          />
        </label>

        {!readOnly && (
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

        {done && (
          <Button
            className="w-full"
            variant="secondary"
            disabled={reopen.isPending}
            onClick={() => reopen.mutate()}
          >
            Вернуть в активные
          </Button>
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
