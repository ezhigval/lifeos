import { useEffect, useState, type ReactNode } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/lib/cn'

type Props = {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
}

const ANIM_MS = 220

export function Sheet({ open, onClose, title, children }: Props) {
  const [mounted, setMounted] = useState(open)
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    if (open) {
      setMounted(true)
      const id = requestAnimationFrame(() => {
        requestAnimationFrame(() => setVisible(true))
      })
      return () => cancelAnimationFrame(id)
    }
    setVisible(false)
    const t = window.setTimeout(() => setMounted(false), ANIM_MS)
    return () => window.clearTimeout(t)
  }, [open])

  useEffect(() => {
    if (!mounted) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [mounted, onClose])

  if (!mounted) return null

  return (
    <div className="fixed inset-0 z-50 flex flex-col justify-end" role="dialog" aria-modal="true">
      <button
        type="button"
        className={cn(
          'absolute inset-0 bg-black/50 transition-opacity duration-200',
          visible ? 'opacity-100' : 'opacity-0',
        )}
        aria-label="Закрыть"
        onClick={onClose}
      />
      <div
        className={cn(
          'relative max-h-[85vh] overflow-y-auto rounded-t-3xl will-change-transform',
          'bg-[var(--tg-theme-bg-color,#0f172a)] px-4 pb-[max(2rem,env(safe-area-inset-bottom))] pt-4',
          'transition-transform duration-200 ease-out',
          visible ? 'translate-y-0' : 'translate-y-full',
        )}
      >
        <div className="mx-auto mb-3 h-1 w-10 rounded-full bg-white/20" aria-hidden />
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold">{title}</h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded-full p-2 text-[var(--tg-theme-hint-color,#94a3b8)] active:scale-95"
            aria-label="Закрыть"
          >
            <X size={20} />
          </button>
        </div>
        {children}
      </div>
    </div>
  )
}
