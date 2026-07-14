import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { api } from '@/api/client'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatShortDateTime } from '@/lib/datetime'
import { confirmAction, hapticError, hapticSuccess, hapticWarning } from '@/lib/telegram'

export function NotesPage() {
  const queryClient = useQueryClient()
  const [q, setQ] = useState('')
  const [search, setSearch] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [body, setBody] = useState('')

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['notes', search],
    queryFn: async () => (await api.notes(search || undefined)).notes,
  })

  const create = useMutation({
    mutationFn: () => api.createNote(body.trim()),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['notes'] })
      setCreateOpen(false)
      setBody('')
    },
    onError: () => hapticError(),
  })

  const remove = useMutation({
    mutationFn: (id: string) => api.deleteNote(id),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['notes'] })
    },
    onError: () => hapticError(),
  })

  return (
    <>
      <Header title="Заметки" subtitle="Inbox" />
      <div className="space-y-4 px-4 pb-6">
        <div className="flex gap-2">
          <input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Поиск…"
            className="min-w-0 flex-1 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
            onKeyDown={(e) => {
              if (e.key === 'Enter') setSearch(q.trim())
            }}
          />
          <Button size="sm" variant="secondary" onClick={() => setSearch(q.trim())}>
            Найти
          </Button>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus size={16} />
          </Button>
        </div>

        {isLoading && <Skeleton className="h-24 w-full" />}
        {isError && <QueryError message="Не удалось загрузить заметки" onRetry={() => void refetch()} />}

        {!isLoading && !isError && (data ?? []).length === 0 && (
          <EmptyState title="Пусто" description="Добавь первую заметку" actionLabel="Создать" onAction={() => setCreateOpen(true)} />
        )}

        <div className="space-y-2">
          {(data ?? []).map((n) => (
            <div
              key={n.id}
              className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4"
            >
              <div className="flex items-start gap-2">
                <p className="min-w-0 flex-1 whitespace-pre-wrap text-sm">{n.body}</p>
                <button
                  type="button"
                  className="rounded-full p-2 text-rose-400"
                  aria-label="Удалить"
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
      </div>

      <Sheet open={createOpen} onClose={() => setCreateOpen(false)} title="Новая заметка">
        <textarea
          value={body}
          onChange={(e) => setBody(e.target.value)}
          rows={5}
          placeholder="Текст…"
          className="mb-4 w-full resize-none rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <Button className="w-full" disabled={!body.trim() || create.isPending} onClick={() => create.mutate()}>
          Сохранить
        </Button>
      </Sheet>
    </>
  )
}
