import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate, useParams } from 'react-router-dom'
import {
  Archive,
  ChevronDown,
  ChevronRight,
  FolderKanban,
  Plus,
} from 'lucide-react'
import { api } from '@/api/client'
import type { Project, Sphere } from '@/api/types'
import { TaskCard } from '@/components/tasks/TaskCard'
import { Button } from '@/components/ui/Button'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { cn } from '@/lib/cn'
import {
  getExpandedProjects,
  getExpandedSpheres,
  setExpandedProjects,
  setExpandedSpheres,
} from '@/lib/storage'
import { hapticLight, hapticSuccess } from '@/lib/telegram'

export function SphereTree() {
  const navigate = useNavigate()
  const [expandedSpheres, setExpandedSpheresState] = useState(getExpandedSpheres)
  const [expandedProjects, setExpandedProjectsState] = useState(getExpandedProjects)

  const { data: spheres, isLoading } = useQuery({
    queryKey: ['spheres'],
    queryFn: async () => (await api.spheres()).spheres,
  })

  const toggleSphere = (id: string) => {
    hapticLight()
    const next = expandedSpheres.includes(id)
      ? expandedSpheres.filter((s) => s !== id)
      : [...expandedSpheres, id]
    setExpandedSpheresState(next)
    setExpandedSpheres(next)
  }

  const toggleProject = (id: string) => {
    hapticLight()
    const next = expandedProjects.includes(id)
      ? expandedProjects.filter((p) => p !== id)
      : [...expandedProjects, id]
    setExpandedProjectsState(next)
    setExpandedProjects(next)
  }

  if (isLoading) {
    return (
      <div className="space-y-2 px-4">
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-12 w-full" />
      </div>
    )
  }

  return (
    <div className="px-4">
      {(spheres ?? []).map((sphere) => (
        <SphereNode
          key={sphere.id}
          sphere={sphere}
          expanded={expandedSpheres.includes(sphere.id)}
          expandedProjects={expandedProjects}
          onToggleSphere={() => toggleSphere(sphere.id)}
          onToggleProject={toggleProject}
          onOpenSphere={() => navigate(`/spheres/${sphere.id}`)}
          onOpenProject={(pid) => navigate(`/spheres/${sphere.id}/projects/${pid}`)}
        />
      ))}
    </div>
  )
}

function SphereNode({
  sphere,
  expanded,
  expandedProjects,
  onToggleSphere,
  onToggleProject,
  onOpenSphere,
  onOpenProject,
}: {
  sphere: Sphere
  expanded: boolean
  expandedProjects: string[]
  onToggleSphere: () => void
  onToggleProject: (id: string) => void
  onOpenSphere: () => void
  onOpenProject: (id: string) => void
}) {
  const { data: projects } = useQuery({
    queryKey: ['projects', sphere.id],
    queryFn: async () => (await api.projects(sphere.id)).projects,
    enabled: expanded,
  })

  return (
    <div className="mb-1">
      <div className="flex items-center gap-1 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)]">
        <button type="button" onClick={onToggleSphere} className="p-3">
          {expanded ? <ChevronDown size={18} /> : <ChevronRight size={18} />}
        </button>
        <button type="button" onClick={onOpenSphere} className="flex-1 py-3 text-left font-medium">
          {sphere.name}
        </button>
      </div>

      {expanded && (
        <div className="ml-4 mt-1 space-y-1 border-l border-white/10 pl-2">
          {(projects ?? []).map((project) => (
            <ProjectNode
              key={project.id}
              project={project}
              expanded={expandedProjects.includes(project.id)}
              onToggle={() => onToggleProject(project.id)}
              onOpen={() => onOpenProject(project.id)}
            />
          ))}
          {(projects ?? []).length === 0 && (
            <p className="py-2 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Нет проектов</p>
          )}
        </div>
      )}
    </div>
  )
}

function ProjectNode({
  project,
  expanded,
  onToggle,
  onOpen,
}: {
  project: Project
  expanded: boolean
  onToggle: () => void
  onOpen: () => void
}) {
  const { data: tasks } = useQuery({
    queryKey: ['project-tasks', project.id],
    queryFn: async () => (await api.projectTasks(project.id)).tasks,
    enabled: expanded,
  })

  return (
    <div>
      <div className="flex items-center gap-1 rounded-xl bg-[var(--tg-theme-bg-color,#0f172a)]/50">
        <button type="button" onClick={onToggle} className="p-2">
          {expanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
        </button>
        <button type="button" onClick={onOpen} className="flex flex-1 items-center gap-2 py-2 text-left text-sm">
          <FolderKanban size={16} className="text-[var(--tg-theme-hint-color,#94a3b8)]" />
          {project.name}
        </button>
      </div>
      {expanded && (
        <div className="ml-3 space-y-1 py-1">
          {(tasks ?? []).slice(0, 8).map((t) => (
            <p
              key={t.id}
              className={cn(
                'truncate py-1 pl-2 text-sm',
                t.status === 'done' && 'text-[var(--tg-theme-hint-color,#94a3b8)] line-through',
              )}
            >
              • {t.title}
            </p>
          ))}
        </div>
      )}
    </div>
  )
}

export function SphereDetailPage() {
  const { sphereId } = useParams<{ sphereId: string }>()
  const queryClient = useQueryClient()
  const [showArchive, setShowArchive] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [newName, setNewName] = useState('')

  const { data: spheres } = useQuery({
    queryKey: ['spheres'],
    queryFn: async () => (await api.spheres()).spheres,
  })
  const sphere = spheres?.find((s) => s.id === sphereId)

  const { data: projects, isLoading } = useQuery({
    queryKey: ['projects', sphereId],
    queryFn: async () => (await api.projects(sphereId!)).projects,
    enabled: Boolean(sphereId),
  })

  const active = (projects ?? []).filter((p) => p.status === 'active')
  const archived = (projects ?? []).filter((p) => p.status !== 'active')

  const createProject = useMutation({
    mutationFn: () =>
      api.createProject({ name: newName.trim(), sphere_ids: [sphereId!] }),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      setCreateOpen(false)
      setNewName('')
    },
  })

  const archiveProject = useMutation({
    mutationFn: (id: string) => api.archiveProject(id),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['projects'] })
    },
  })

  if (!sphere) {
    return <p className="p-4 text-[var(--tg-theme-hint-color,#94a3b8)]">Сфера не найдена</p>
  }

  return (
    <div className="space-y-4 px-4">
      <section>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="font-semibold">Активные проекты</h2>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus size={16} className="mr-1" />
            Проект
          </Button>
        </div>
        {isLoading && <Skeleton className="h-20 w-full" />}
        <div className="space-y-2">
          {active.map((p) => (
            <ProjectCard
              key={p.id}
              project={p}
              sphereId={sphereId!}
              onArchive={() => archiveProject.mutate(p.id)}
            />
          ))}
          {!isLoading && active.length === 0 && (
            <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Нет активных проектов</p>
          )}
        </div>
      </section>

      <section className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4">
        <h3 className="mb-2 text-sm font-medium text-[var(--tg-theme-hint-color,#94a3b8)]">Статистика</h3>
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <div className="text-2xl font-bold tabular-nums">{active.length}</div>
            <div className="text-[var(--tg-theme-hint-color,#94a3b8)]">активных</div>
          </div>
          <div>
            <div className="text-2xl font-bold tabular-nums">{archived.length}</div>
            <div className="text-[var(--tg-theme-hint-color,#94a3b8)]">в архиве</div>
          </div>
        </div>
      </section>

      <section>
        <button
          type="button"
          className="flex w-full items-center gap-2 py-2 text-sm font-medium text-[var(--tg-theme-hint-color,#94a3b8)]"
          onClick={() => setShowArchive((v) => !v)}
        >
          <Archive size={16} />
          Архив проектов ({archived.length})
          {showArchive ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
        </button>
        {showArchive && (
          <div className="mt-2 space-y-2">
            {archived.map((p) => (
              <div
                key={p.id}
                className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-3 opacity-70"
              >
                {p.name}
              </div>
            ))}
            {archived.length === 0 && (
              <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Архив пуст</p>
            )}
          </div>
        )}
      </section>

      <Sheet open={createOpen} onClose={() => setCreateOpen(false)} title="Новый проект">
        <input
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          placeholder="Название проекта"
          className="mb-4 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <Button
          className="w-full"
          disabled={!newName.trim() || createProject.isPending}
          onClick={() => createProject.mutate()}
        >
          Создать
        </Button>
      </Sheet>
    </div>
  )
}

function ProjectCard({
  project,
  sphereId,
  onArchive,
}: {
  project: Project
  sphereId: string
  onArchive: () => void
}) {
  const navigate = useNavigate()
  const pct = project.target_value
    ? Math.min(
        100,
        (Number(project.current_value || 0) / Number(project.target_value)) * 100,
      )
    : null

  return (
    <button
      type="button"
      onClick={() => navigate(`/spheres/${sphereId}/projects/${project.id}`)}
      className="w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4 text-left"
    >
      <div className="flex items-start justify-between gap-2">
        <span className="font-medium">{project.name}</span>
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            onArchive()
          }}
          className="p-1 text-[var(--tg-theme-hint-color,#94a3b8)]"
        >
          <Archive size={16} />
        </button>
      </div>
      {pct !== null && (
        <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-black/20">
          <div className="h-full rounded-full bg-[var(--tg-theme-button-color,#22c55e)]" style={{ width: `${pct}%` }} />
        </div>
      )}
    </button>
  )
}

export function ProjectDetailPage() {
  const { sphereId, projectId } = useParams<{ sphereId: string; projectId: string }>()
  const queryClient = useQueryClient()
  const [showDone, setShowDone] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [newTitle, setNewTitle] = useState('')

  const { data: projects } = useQuery({
    queryKey: ['projects', sphereId],
    queryFn: async () => (await api.projects(sphereId!)).projects,
  })
  const project = projects?.find((p) => p.id === projectId)

  const { data: tasks, isLoading } = useQuery({
    queryKey: ['project-tasks', projectId],
    queryFn: async () => (await api.projectTasks(projectId!)).tasks,
    enabled: Boolean(projectId),
  })

  const { data: progress } = useQuery({
    queryKey: ['project-progress', projectId],
    queryFn: () => api.projectProgress(projectId!),
    enabled: Boolean(projectId),
  })

  const complete = useMutation({
    mutationFn: (id: string) => api.completeTask(id),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['project-tasks', projectId] })
    },
  })

  const createTask = useMutation({
    mutationFn: () =>
      api.createTask({ title: newTitle.trim(), project_ids: [projectId!] }),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['project-tasks', projectId] })
      setCreateOpen(false)
      setNewTitle('')
    },
  })

  const active = (tasks ?? []).filter((t) => t.status !== 'done' && t.status !== 'cancelled')
  const done = (tasks ?? []).filter((t) => t.status === 'done')

  if (!project) {
    return <p className="p-4">Проект не найден</p>
  }

  return (
    <div className="space-y-4 px-4">
      <section>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="font-semibold">Активные задачи</h2>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus size={16} className="mr-1" />
            Задача
          </Button>
        </div>
        {isLoading && <Skeleton className="h-14 w-full" />}
        <div className="space-y-2">
          {active.map((t) => (
            <TaskCard
              key={t.id}
              title={t.title}
              priority={t.priority}
              onComplete={() => complete.mutate(t.id)}
            />
          ))}
          {!isLoading && active.length === 0 && (
            <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Нет активных задач</p>
          )}
        </div>
      </section>

      {progress?.has_target && (
        <section className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4">
          <h3 className="mb-2 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Прогресс</h3>
          <div className="text-2xl font-bold tabular-nums">{progress.percent}%</div>
          <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
            {progress.current} / {progress.target}
          </p>
        </section>
      )}

      <section>
        <button
          type="button"
          className="flex w-full items-center gap-2 py-2 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]"
          onClick={() => setShowDone((v) => !v)}
        >
          Выполненные ({done.length})
          {showDone ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
        </button>
        {showDone && (
          <div className="space-y-2">
            {done.map((t) => (
              <TaskCard key={t.id} title={t.title} done />
            ))}
          </div>
        )}
      </section>

      <Sheet open={createOpen} onClose={() => setCreateOpen(false)} title="Новая задача">
        <input
          value={newTitle}
          onChange={(e) => setNewTitle(e.target.value)}
          placeholder="Название задачи"
          className="mb-4 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <Button
          className="w-full"
          disabled={!newTitle.trim() || createTask.isPending}
          onClick={() => createTask.mutate()}
        >
          Создать
        </Button>
      </Sheet>
    </div>
  )
}
