import { X } from 'lucide-react'
import { cn } from '@/lib/cn'
import type { ReactNode } from 'react'

type Props = {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
}

export function Sheet({ open, onClose, title, children }: Props) {
  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex flex-col justify-end">
      <button
        type="button"
        className="absolute inset-0 bg-black/50"
        aria-label="Закрыть"
        onClick={onClose}
      />
      <div
        className={cn(
          'relative max-h-[85vh] overflow-y-auto rounded-t-3xl',
          'bg-[var(--tg-theme-bg-color,#0f172a)] px-4 pb-8 pt-4',
        )}
      >
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold">{title}</h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded-full p-2 text-[var(--tg-theme-hint-color,#94a3b8)]"
          >
            <X size={20} />
          </button>
        </div>
        {children}
      </div>
    </div>
  )
}
