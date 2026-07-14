import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '@/api/client'
import { TaskCard } from '@/components/tasks/TaskCard'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
import { hapticError, hapticLight, hapticSuccess } from '@/lib/telegram'
import { useState } from 'react'

const LIMIT = 7

export function UpcomingTasks() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [actionError, setActionError] = useState<string | null>(null)

  const {
    data: todayTasks,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['tasks', 'today'],
    queryFn: async () => {
      const res = await api.tasksToday()
      return Array.isArray(res.tasks) ? res.tasks : []
    },
  })

  const complete = useMutation({
    mutationFn: (id: string) => api.completeTask(id),
    onSuccess: () => {
      hapticSuccess()
      setActionError(null)
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      queryClient.invalidateQueries({ queryKey: ['priorities'] })
      queryClient.invalidateQueries({ queryKey: ['project-tasks'] })
    },
    onError: (err) => {
      hapticError()
      setActionError(ruApiError(err, 'Не удалось выполнить задачу'))
    },
  })

  const items = (todayTasks ?? [])
    .filter((t) => t.status !== 'done' && t.status !== 'cancelled')
    .slice(0, LIMIT)

  return (
    <section>
      <div className="mb-3 flex items-center justify-between px-4">
        <h2 className="text-base font-semibold">Ближайшие задачи</h2>
        <Link to="/spheres" className="text-sm text-[var(--tg-theme-link-color,#22c55e)]">
          Все →
        </Link>
      </div>

      <div className="space-y-2 px-4">
        {isLoading &&
          Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-14 w-full" />)}

        {isError && !isLoading && (
          <QueryError message="Не удалось загрузить задачи" onRetry={() => void refetch()} />
        )}

        {actionError && !isError && (
          <p className="text-sm text-rose-400" role="alert">
            {actionError}
          </p>
        )}

        {!isLoading && !isError && items.length === 0 && (
          <EmptyState
            title="Всё чисто на сегодня"
            description="Новые задачи — в боте или в проекте сферы"
          />
        )}

        {!isError &&
          items.map((t) => (
            <TaskCard
              key={t.id}
              title={t.title}
              detail={t.due_date ? formatDue(t.due_date) : 'сегодня'}
              priority={t.priority || 'medium'}
              kind={t.kind || 'task'}
              onComplete={() => {
                hapticLight()
                setActionError(null)
                complete.mutate(t.id)
              }}
              onOpen={() => navigate(`/tasks/${t.id}`)}
            />
          ))}
      </div>
    </section>
  )
}

function formatDue(iso: string): string {
  try {
    const d = new Date(iso + 'T12:00:00')
    return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })
  } catch {
    return iso
  }
}
