import type { ReactNode } from 'react'
import { Button } from '@/components/ui/Button'

type Props = {
  title: string
  description?: string
  actionLabel?: string
  onAction?: () => void
  children?: ReactNode
}

export function EmptyState({ title, description, actionLabel, onAction, children }: Props) {
  return (
    <div className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-8 text-center">
      <p className="font-medium">{title}</p>
      {description && (
        <p className="mt-1 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">{description}</p>
      )}
      {actionLabel && onAction && (
        <Button variant="secondary" size="sm" className="mt-4" onClick={onAction}>
          {actionLabel}
        </Button>
      )}
      {children}
    </div>
  )
}
