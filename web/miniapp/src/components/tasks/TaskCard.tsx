import { Check } from 'lucide-react'
import { cn } from '@/lib/cn'

const PRIORITY_COLORS: Record<string, string> = {
  urgent: 'bg-red-500',
  high: 'bg-orange-500',
  medium: 'bg-amber-500',
  low: 'bg-slate-500',
}

type Props = {
  title: string
  detail?: string
  priority?: string
  done?: boolean
  onComplete?: () => void
}

export function TaskCard({ title, detail, priority = 'medium', done, onComplete }: Props) {
  return (
    <div
      className={cn(
        'flex items-start gap-3 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-3',
        done && 'opacity-50',
      )}
    >
      <button
        type="button"
        disabled={done || !onComplete}
        onClick={onComplete}
        className={cn(
          'mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full border-2 transition',
          done
            ? 'border-emerald-500 bg-emerald-500 text-white'
            : 'border-[var(--tg-theme-hint-color,#64748b)]',
        )}
      >
        {done && <Check size={14} />}
      </button>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span
            className={cn('h-2 w-2 shrink-0 rounded-full', PRIORITY_COLORS[priority] ?? PRIORITY_COLORS.medium)}
          />
          <p className={cn('truncate font-medium', done && 'line-through')}>{title}</p>
        </div>
        {detail && (
          <p className="mt-0.5 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">{detail}</p>
        )}
      </div>
    </div>
  )
}
