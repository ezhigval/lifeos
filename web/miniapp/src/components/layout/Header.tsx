import { Settings } from 'lucide-react'
import type { ReactNode } from 'react'

type Props = {
  title: string
  subtitle?: string
  onSettings?: () => void
  children?: ReactNode
}

export function Header({ title, subtitle, onSettings, children }: Props) {
  return (
    <header className="sticky top-0 z-30 bg-[var(--tg-theme-bg-color,#0f172a)]/90 px-4 pb-3 pt-4 backdrop-blur-md">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">{title}</h1>
          {subtitle && (
            <p className="mt-0.5 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
              {subtitle}
            </p>
          )}
        </div>
        {onSettings && (
          <button
            type="button"
            onClick={onSettings}
            className="rounded-full p-2 text-[var(--tg-theme-hint-color,#94a3b8)]"
          >
            <Settings size={22} />
          </button>
        )}
      </div>
      {children}
    </header>
  )
}
