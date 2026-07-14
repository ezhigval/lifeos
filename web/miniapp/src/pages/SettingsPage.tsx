import { Header } from '@/components/layout/Header'
import { ListRow } from '@/components/ui/ListRow'
import { EmptyState } from '@/components/ui/EmptyState'

export function SettingsPage() {
  return (
    <>
      <Header title="Настройки" subtitle="Профиль и режим уведомлений" />
      <div className="space-y-4 px-4 pb-4">
        <EmptyState
          title="Формы подключим в Phase B"
          description="Утренний/вечерний обзор и quiet hours уже есть в API — UI сделаем после backend auth."
        />
        <div className="space-y-2">
          <ListRow
            title="Время обзоров"
            subtitle="Утро / вечер — через бота: Настройки"
            disabled
          />
          <ListRow
            title="Quiet hours"
            subtitle="Не беспокоить — через бота"
            disabled
          />
          <ListRow
            title="Сферы"
            subtitle="CRUD сфер — Phase B (MA-B7)"
            disabled
          />
        </div>
      </div>
    </>
  )
}
