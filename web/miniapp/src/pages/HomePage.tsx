import { useQuery } from '@tanstack/react-query'
import { FinanceCard, useFinancePeriod } from '@/components/finance/FinanceCard'
import { Header } from '@/components/layout/Header'
import { UpcomingTasks } from '@/components/tasks/UpcomingTasks'
import { api, enrichFinanceCategories } from '@/api/client'
import { periodKey } from '@/lib/periods'
import { tgUser } from '@/lib/telegram'

export function HomePage() {
  const user = tgUser()
  const { period, setPeriod } = useFinancePeriod()

  const { data: overview, isLoading } = useQuery({
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
      <Header title={greeting} subtitle={dateStr} />
      <div className="space-y-6 pb-4">
        <UpcomingTasks />
        <div className="px-4">
          <FinanceCard
            overview={overview}
            isLoading={isLoading}
            period={period}
            onPeriodChange={setPeriod}
          />
        </div>
      </div>
    </>
  )
}
