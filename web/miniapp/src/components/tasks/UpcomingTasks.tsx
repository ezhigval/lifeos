import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '@/api/client'
import { TaskCard } from '@/components/tasks/TaskCard'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Skeleton } from '@/components/ui/Skeleton'
import { hapticError, hapticLight, hapticSuccess } from '@/lib/telegram'

const LIMIT = 7

export function UpcomingTasks() {
  const queryClient = useQueryClient()

  const {
    data: priorities,
    isLoading: loadingP,
    isError: errorP,
    refetch: refetchP,
  } = useQuery({
    queryKey: ['priorities'],
    queryFn: async () => (await api.priorities()).priorities,
  })

  const {
    data: todayTasks,
    isLoading: loadingT,
    isError: errorT,
    refetch: refetchT,
  } = useQuery({
    queryKey: ['tasks', 'today'],
    queryFn: async () => (await api.tasksToday()).tasks,
  })

  const complete = useMutation({
    mutationFn: (id: string) => api.completeTask(id),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      queryClient.invalidateQueries({ queryKey: ['priorities'] })
    },
    onError: () => hapticError(),
  })

  const isLoading = loadingP || loadingT
  const isError = errorP || errorT

  const items = buildUpcomingList(priorities ?? [], todayTasks ?? [])

  return (
    <section>
      <div className="mb-3 flex items-center justify-between px-4">
        <h2 className="text-base font-semibold">Ближайшие задачи</h2>
        <Link
          to="/spheres"
          className="text-sm text-[var(--tg-theme-link-color,#22c55e)]"
        >
          Все →
        </Link>
      </div>

      <div className="space-y-2 px-4">
        {isLoading &&
          Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-14 w-full" />
          ))}

        {isError && !isLoading && (
          <QueryError
            message="Не удалось загрузить задачи"
            onRetry={() => {
              void refetchP()
              void refetchT()
            }}
          />
        )}

        {!isLoading && !isError && items.length === 0 && (
          <EmptyState
            title="Всё чисто на ближайшие дни"
            description="Новые задачи — в боте или в проекте сферы"
          />
        )}

        {!isError &&
          items.map((item) => (
            <TaskCard
              key={item.key}
              title={item.title}
              detail={item.detail}
              priority={item.priority}
              done={item.done}
              onComplete={
                item.taskId
                  ? () => {
                      hapticLight()
                      complete.mutate(item.taskId!)
                    }
                  : undefined
              }
            />
          ))}
      </div>
    </section>
  )
}

type UpcomingItem = {
  key: string
  title: string
  detail?: string
  priority: string
  done?: boolean
  taskId?: string
}

function buildUpcomingList(
  priorities: { kind: string; title: string; detail: string; score: number }[],
  tasks: { id: string; title: string; status: string; priority: string; due_date?: string }[],
): UpcomingItem[] {
  const seen = new Set<string>()
  const out: UpcomingItem[] = []

  for (const t of tasks) {
    if (t.status === 'done' || t.status === 'cancelled') continue
    if (out.length >= LIMIT) break
    seen.add(t.title.toLowerCase())
    out.push({
      key: t.id,
      taskId: t.id,
      title: t.title,
      detail: t.due_date ? formatDue(t.due_date) : 'сегодня',
      priority: t.priority || 'medium',
    })
  }

  for (const p of priorities) {
    if (out.length >= LIMIT) break
    const key = p.title.toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    const matched = tasks.find(
      (t) =>
        t.status !== 'done' &&
        t.status !== 'cancelled' &&
        t.title.toLowerCase() === key,
    )
    out.push({
      key: matched?.id ?? `p-${p.title}`,
      taskId: matched?.id,
      title: p.title,
      detail: p.detail || priorityLabel(scoreToPriority(p.score)),
      priority: scoreToPriority(p.score),
    })
  }

  return out.slice(0, LIMIT)
}

function scoreToPriority(score: number): string {
  if (score >= 80) return 'urgent'
  if (score >= 60) return 'high'
  if (score >= 40) return 'medium'
  return 'low'
}

function priorityLabel(p: string): string {
  switch (p) {
    case 'urgent':
      return 'срочно'
    case 'high':
      return 'высокий'
    case 'low':
      return 'низкий'
    default:
      return 'средний'
  }
}

function formatDue(iso: string): string {
  try {
    const d = new Date(iso + 'T12:00:00')
    return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })
  } catch {
    return iso
  }
}
