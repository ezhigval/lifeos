import { lazy, Suspense, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import { FinanceCard, useFinancePeriod } from '@/components/finance/FinanceCard'
import { Header } from '@/components/layout/Header'
import { UpcomingTasks } from '@/components/tasks/UpcomingTasks'
import { HomeHabits } from '@/components/habits/HomeHabits'

const CreateTaskSheet = lazy(() =>
  import('@/components/tasks/CreateTaskSheet').then((m) => ({ default: m.CreateTaskSheet })),
)
import { QueryError } from '@/components/ui/QueryError'
import { api, enrichFinanceCategories } from '@/api/client'
import { periodKey } from '@/lib/periods'
import { hapticLight, tgUser } from '@/lib/telegram'

export function HomePage() {
  const navigate = useNavigate()
  const user = tgUser()
  const [createOpen, setCreateOpen] = useState(false)
  const [sheetMounted, setSheetMounted] = useState(false)
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

      {/* FAB: quick task creation (MA-B1) */}
      <button
        type="button"
        aria-label="Новая задача"
        onClick={() => {
          hapticLight()
          setSheetMounted(true)
          setCreateOpen(true)
        }}
        className={
          'fixed bottom-20 right-4 z-40 flex h-14 w-14 items-center justify-center rounded-full ' +
          'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)] ' +
          'shadow-lg transition active:scale-95'
        }
      >
        <Plus size={24} />
      </button>

      {sheetMounted ? (
        <Suspense fallback={null}>
          <CreateTaskSheet open={createOpen} onClose={() => setCreateOpen(false)} />
        </Suspense>
      ) : null}
    </>
  )
}
