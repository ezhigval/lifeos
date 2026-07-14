import type { ReactNode } from 'react'
import { ChevronRight } from 'lucide-react'
import { cn } from '@/lib/cn'

type Props = {
  title: string
  subtitle?: string
  icon?: ReactNode
  trailing?: ReactNode
  onClick?: () => void
  disabled?: boolean
  className?: string
}

export function ListRow({
  title,
  subtitle,
  icon,
  trailing,
  onClick,
  disabled,
  className,
}: Props) {
  const content = (
    <>
      {icon && (
        <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-black/20 text-[var(--tg-theme-button-color,#22c55e)]">
          {icon}
        </span>
      )}
      <span className="min-w-0 flex-1 text-left">
        <span className="block truncate font-medium">{title}</span>
        {subtitle && (
          <span className="mt-0.5 block truncate text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
            {subtitle}
          </span>
        )}
      </span>
      {trailing ??
        (onClick ? (
          <ChevronRight size={18} className="shrink-0 text-[var(--tg-theme-hint-color,#94a3b8)]" />
        ) : null)}
    </>
  )

  const classes = cn(
    'flex w-full items-center gap-3 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-3 py-3',
    onClick && !disabled && 'active:scale-[0.99] transition-transform',
    disabled && 'opacity-60',
    className,
  )

  if (onClick) {
    return (
      <button type="button" onClick={onClick} disabled={disabled} className={classes}>
        {content}
      </button>
    )
  }

  return <div className={classes}>{content}</div>
}
