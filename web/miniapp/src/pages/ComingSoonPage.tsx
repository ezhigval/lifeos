import { useNavigate } from 'react-router-dom'
import { Header } from '@/components/layout/Header'
import { EmptyState } from '@/components/ui/EmptyState'
import { Button } from '@/components/ui/Button'

type Props = {
  title: string
  subtitle?: string
  description: string
  hint?: string
}

/** Temporary screen until domain UI is wired to API (Phase B/C). */
export function ComingSoonPage({ title, subtitle, description, hint }: Props) {
  const navigate = useNavigate()

  return (
    <>
      <Header title={title} subtitle={subtitle ?? 'Скоро'} />
      <div className="px-4">
        <EmptyState title={description} description={hint}>
          <div className="mt-4 flex flex-col gap-2">
            <p className="text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
              Пока данные удобнее вводить через бота (NL). Этот экран — точка входа Phase B.
            </p>
            <Button variant="secondary" size="sm" onClick={() => navigate('/more')}>
              Назад к разделам
            </Button>
          </div>
        </EmptyState>
      </div>
    </>
  )
}
