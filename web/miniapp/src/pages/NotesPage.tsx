import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { api } from '@/api/client'
import type { Note } from '@/api/types'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
import { formatShortDateTime } from '@/lib/datetime'
import { confirmAction, hapticError, hapticSuccess, hapticWarning } from '@/lib/telegram'

export function NotesPage() {
  const queryClient = useQueryClient()
  const [q, setQ] = useState('')
  const [search, setSearch] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [body, setBody] = useState('')
  const [formError, setFormError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['notes', search],
    queryFn: async () => {
      const res = await api.notes(search || undefined)
      return Array.isArray(res.notes) ? res.notes : []
    },
  })

  const notes = data ?? []

  const create = useMutation({
    mutationFn: () => api.createNote(body.trim()),
    onSuccess: (note) => {
      hapticSuccess()
      queryClient.setQueryData<Note[]>(['notes', search], (old) => {
        if (search && !note.body.toLowerCase().includes(search.toLowerCase())) {
          return old ?? []
        }
        return [note, ...(old ?? []).filter((n) => n.id !== note.id)]
      })
      void queryClient.invalidateQueries({ queryKey: ['notes'] })
      setCreateOpen(false)
      setBody('')
      setFormError(null)
    },
    onError: (err) => {
      hapticError()
      setFormError(ruApiError(err, 'Не удалось сохранить заметку'))
    },
  })

  const remove = useMutation({
    mutationFn: (id: string) => api.deleteNote(id),
    onMutate: async (id) => {
      setActionError(null)
      await queryClient.cancelQueries({ queryKey: ['notes'] })
      const prev = queryClient.getQueryData<Note[]>(['notes', search])
      queryClient.setQueryData<Note[]>(['notes', search], (old) =>
        (old ?? []).filter((n) => n.id !== id),
      )
      return { prev }
    },
    onSuccess: () => hapticSuccess(),
    onError: (err, _id, ctx) => {
      hapticError()
      if (ctx?.prev) queryClient.setQueryData(['notes', search], ctx.prev)
      setActionError(ruApiError(err, 'Не удалось удалить'))
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: ['notes'] })
    },
  })

  const openCreate = () => {
    setFormError(null)
    setBody('')
    setCreateOpen(true)
  }

  const runSearch = () => setSearch(q.trim())
  const removingId = remove.isPending ? remove.variables : null

  return (
    <>
      <Header title="Заметки" subtitle={search ? `Поиск: ${search}` : 'Inbox'} />
      <div className="space-y-4 px-4 pb-6">
        <div className="flex gap-2">
          <input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Поиск…"
            className="min-w-0 flex-1 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
            onKeyDown={(e) => {
              if (e.key === 'Enter') runSearch()
            }}
          />
          <Button size="sm" variant="secondary" onClick={runSearch}>
            Найти
          </Button>
          {notes.length > 0 && (
            <Button size="sm" onClick={openCreate} aria-label="Создать">
              <Plus size={16} />
            </Button>
          )}
        </div>

        {actionError && (
          <p className="text-sm text-rose-400" role="alert">
            {actionError}
          </p>
        )}

        {isLoading && (
          <div className="space-y-2">
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-24 w-full" />
          </div>
        )}

        {isError && (
          <QueryError message="Не удалось загрузить заметки" onRetry={() => void refetch()} />
        )}

        {!isLoading && !isError && notes.length === 0 && (
          <EmptyState
            title={search ? 'Ничего не найдено' : 'Пусто'}
            description={search ? 'Попробуй другой запрос' : 'Добавь первую заметку'}
            actionLabel={search ? undefined : 'Создать'}
            onAction={search ? undefined : openCreate}
          />
        )}

        {!isLoading && !isError && notes.length > 0 && (
          <div className="space-y-2">
            {notes.map((n) => (
              <div
                key={n.id}
                className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4"
              >
                <div className="flex items-start gap-2">
                  <p className="min-w-0 flex-1 whitespace-pre-wrap text-sm">{n.body}</p>
                  <button
                    type="button"
                    className="rounded-full p-2 text-rose-400 disabled:opacity-50"
                    aria-label="Удалить"
                    disabled={removingId === n.id}
                    onClick={async () => {
                      hapticWarning()
                      if (await confirmAction('Удалить заметку?')) remove.mutate(n.id)
                    }}
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
                <p className="mt-2 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
                  {formatShortDateTime(n.created_at)}
                </p>
              </div>
            ))}
          </div>
        )}
      </div>

      <Sheet
        open={createOpen}
        onClose={() => {
          setCreateOpen(false)
          setFormError(null)
        }}
        title="Новая заметка"
      >
        <textarea
          value={body}
          onChange={(e) => {
            setBody(e.target.value)
            if (formError) setFormError(null)
          }}
          rows={5}
          placeholder="Текст…"
          className="mb-3 w-full resize-none rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          autoFocus
        />
        {formError && (
          <p className="mb-3 text-sm text-rose-400" role="alert">
            {formError}
          </p>
        )}
        <Button
          className="w-full"
          disabled={!body.trim() || create.isPending}
          onClick={() => create.mutate()}
        >
          Сохранить
        </Button>
      </Sheet>
    </>
  )
}
