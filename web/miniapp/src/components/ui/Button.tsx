import { cn } from '@/lib/cn'
import type { ButtonHTMLAttributes } from 'react'

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'income' | 'expense'
  size?: 'sm' | 'md' | 'lg'
}

export function Button({
  className,
  variant = 'primary',
  size = 'md',
  ...props
}: Props) {
  return (
    <button
      className={cn(
        'inline-flex items-center justify-center rounded-2xl font-medium transition active:scale-[0.98] disabled:opacity-50',
        size === 'sm' && 'h-9 px-3 text-sm',
        size === 'md' && 'h-11 px-4 text-sm',
        size === 'lg' && 'h-12 px-5 text-base',
        variant === 'primary' &&
          'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]',
        variant === 'secondary' &&
          'bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-[var(--tg-theme-text-color,#f8fafc)]',
        variant === 'ghost' && 'bg-transparent text-[var(--tg-theme-text-color,#f8fafc)]',
        variant === 'danger' && 'bg-red-500/15 text-red-400',
        variant === 'income' && 'bg-emerald-500/20 text-emerald-400',
        variant === 'expense' && 'bg-rose-500/20 text-rose-400',
        className,
      )}
      {...props}
    />
  )
}
