import { Button } from '@/components/ui/Button'

type Props = {
  message?: string
  onRetry?: () => void
}

export function QueryError({ message = 'Не удалось загрузить данные', onRetry }: Props) {
  return (
    <div className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-6 text-center">
      <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">{message}</p>
      {onRetry && (
        <Button variant="secondary" size="sm" className="mt-3" onClick={onRetry}>
          Повторить
        </Button>
      )}
    </div>
  )
}
