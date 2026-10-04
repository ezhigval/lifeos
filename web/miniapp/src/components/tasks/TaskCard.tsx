import { useRef, useState, type PointerEvent as ReactPointerEvent } from 'react'
import { Check } from 'lucide-react'
import { cn } from '@/lib/cn'

const PRIORITY_COLORS: Record<string, string> = {
  urgent: 'bg-red-500',
  high: 'bg-orange-500',
  medium: 'bg-amber-500',
  low: 'bg-slate-500',
}

const KIND_LABEL: Record<string, string> = {
  task: 'Задача',
  reminder: 'Напоминание',
  meeting: 'Встреча',
}

type Props = {
  title: string
  detail?: string
  priority?: string
  kind?: string
  done?: boolean
  /** TASK-011 item 7: sphere badges (N:M task↔sphere) */
  spheres?: { id: string; name: string }[]
  /** Swipe right → complete */
  onComplete?: () => void
  /** Tap → open detail */
  onOpen?: () => void
}

const SPHERE_BADGE_COLORS = [
  'bg-sky-500/15 text-sky-400',
  'bg-violet-500/15 text-violet-400',
  'bg-emerald-500/15 text-emerald-400',
  'bg-rose-500/15 text-rose-400',
  'bg-amber-500/15 text-amber-400',
]

function hashSphereId(id: string): number {
  let h = 0
  for (let i = 0; i < id.length; i++) h = (h * 31 + id.charCodeAt(i)) | 0
  return Math.abs(h)
}

const SWIPE_THRESHOLD = 72

export function TaskCard({
  title,
  detail,
  priority = 'medium',
  kind = 'task',
  done,
  spheres,
  onComplete,
  onOpen,
}: Props) {
  const startX = useRef(0)
  const [dx, setDx] = useState(0)
  const [swiping, setSwiping] = useState(false)

  const onPointerDown = (e: ReactPointerEvent<HTMLDivElement>) => {
    if (done || !onComplete) return
    startX.current = e.clientX
    setSwiping(true)
    e.currentTarget.setPointerCapture(e.pointerId)
  }

  const onPointerMove = (e: ReactPointerEvent<HTMLDivElement>) => {
    if (!swiping) return
    const delta = e.clientX - startX.current
    // only reveal complete swipe to the right
    setDx(Math.max(0, Math.min(120, delta)))
  }

  const finish = () => {
    if (!swiping) return
    setSwiping(false)
    if (dx >= SWIPE_THRESHOLD && onComplete) {
      onComplete()
    }
    setDx(0)
  }

  const reveal = Math.min(1, dx / SWIPE_THRESHOLD)

  return (
    <div className="relative overflow-hidden rounded-2xl">
      <div
        className="absolute inset-y-0 left-0 flex w-[120px] items-center justify-start bg-emerald-600/90 pl-4 text-sm font-medium text-white"
        style={{ opacity: reveal }}
        aria-hidden
      >
        <Check size={18} className="mr-1" /> Готово
      </div>
      <div
        role="button"
        tabIndex={0}
        onClick={() => {
          if (dx > 8) return
          onOpen?.()
        }}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') onOpen?.()
        }}
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={finish}
        onPointerCancel={finish}
        className={cn(
          'relative flex items-start gap-3 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-3 touch-pan-y',
          done && 'opacity-50',
          onOpen && 'active:brightness-110',
        )}
        style={{ transform: `translateX(${dx}px)`, transition: swiping ? 'none' : 'transform 160ms ease' }}
      >
        <span
          className={cn(
            'mt-1.5 h-2.5 w-2.5 shrink-0 rounded-full',
            done ? 'bg-emerald-500' : PRIORITY_COLORS[priority] ?? PRIORITY_COLORS.medium,
          )}
        />
        <div className="min-w-0 flex-1 text-left">
          <div className="flex items-center gap-2">
            <p className={cn('truncate font-medium', done && 'line-through')}>{title}</p>
          </div>
          <p className="mt-0.5 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
            {[KIND_LABEL[kind] ?? kind, detail].filter(Boolean).join(' · ')}
          </p>
          {spheres && spheres.length > 0 && (
            <div className="mt-1 flex flex-wrap gap-1">
              {spheres.slice(0, 3).map((s) => (
                <span
                  key={s.id}
                  className={cn(
                    'rounded-full px-2 py-0.5 text-[10px] font-medium',
                    SPHERE_BADGE_COLORS[hashSphereId(s.id) % SPHERE_BADGE_COLORS.length],
                  )}
                >
                  {s.name}
                </span>
              ))}
              {spheres.length > 3 && (
                <span className="rounded-full bg-slate-500/15 px-2 py-0.5 text-[10px] text-slate-400">
                  +{spheres.length - 3}
                </span>
              )}
            </div>
          )}
          {!done && onComplete && (
            <p className="mt-1 text-[10px] text-[var(--tg-theme-hint-color,#64748b)]">
              Свайп вправо — выполнить · тап — открыть
            </p>
          )}
        </div>
      </div>
    </div>
  )
}
