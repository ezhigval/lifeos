import { cn } from '@/lib/cn'

export function Skeleton({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        'animate-pulse rounded-xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)]',
        className,
      )}
    />
  )
}
