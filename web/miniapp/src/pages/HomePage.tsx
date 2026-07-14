import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { FinanceCard, useFinancePeriod } from '@/components/finance/FinanceCard'
import { Header } from '@/components/layout/Header'
import { UpcomingTasks } from '@/components/tasks/UpcomingTasks'
import { HomeHabits } from '@/components/habits/HomeHabits'
import { QueryError } from '@/components/ui/QueryError'
import { api, enrichFinanceCategories } from '@/api/client'
import { periodKey } from '@/lib/periods'
import { tgUser } from '@/lib/telegram'

export function HomePage() {
  const navigate = useNavigate()
  const user = tgUser()
  const { period, setPeriod } = useFinancePeriod()

  const {
    data: overview,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['finance', periodKey(period)],
    queryFn: () => api.financeOverview(period).then(enrichFinanceCategories),
  })

  const greeting = user?.first_name ? `Привет, ${user.first_name}` : 'LifeOS'
  const dateStr = new Date().toLocaleDateString('ru-RU', {
    weekday: 'short',
    day: 'numeric',
    month: 'long',
  })

  return (
    <>
      <Header
        title={greeting}
        subtitle={dateStr}
        onSettings={() => navigate('/more/settings')}
      />
      <div className="space-y-6 pb-4">
        <UpcomingTasks />
        <HomeHabits />
        <div className="px-4">
          {isError ? (
            <QueryError message="Не удалось загрузить финансы" onRetry={() => void refetch()} />
          ) : (
            <FinanceCard
              overview={overview}
              isLoading={isLoading}
              period={period}
              onPeriodChange={setPeriod}
            />
          )}
        </div>
      </div>
    </>
  )
}
