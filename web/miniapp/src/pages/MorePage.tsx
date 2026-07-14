import { useNavigate } from 'react-router-dom'
import {
  AlarmClock,
  BarChart3,
  Briefcase,
  Calendar,
  FileText,
  HeartPulse,
  Settings,
  Sparkles,
  Wallet,
} from 'lucide-react'
import { Header } from '@/components/layout/Header'
import { ListRow } from '@/components/ui/ListRow'
import { hapticLight, tgUser } from '@/lib/telegram'

const sections = [
  {
    title: 'Ежедневный цикл',
    items: [
      {
        to: '/more/habits',
        title: 'Привычки',
        subtitle: 'Отметить за сегодня',
        icon: <Sparkles size={20} />,
      },
      {
        to: '/more/calendar',
        title: 'Календарь',
        subtitle: 'События на сегодня',
        icon: <Calendar size={20} />,
      },
      {
        to: '/more/reminders',
        title: 'Напоминания',
        subtitle: 'Отложенные пуши',
        icon: <AlarmClock size={20} />,
      },
      {
        to: '/more/settings',
        title: 'Настройки',
        subtitle: 'Обзоры, quiet hours, сферы',
        icon: <Settings size={20} />,
      },
    ],
  },
  {
    title: 'Запись',
    items: [
      {
        to: '/more/notes',
        title: 'Заметки',
        subtitle: 'Быстрый inbox',
        icon: <FileText size={20} />,
      },
      {
        to: '/more/health',
        title: 'Здоровье',
        subtitle: 'Вес, шаги, сон',
        icon: <HeartPulse size={20} />,
      },
      {
        to: '/more/debts',
        title: 'Долги',
        subtitle: 'Кредиторы и платежи',
        icon: <Wallet size={20} />,
      },
    ],
  },
  {
    title: 'Обзор',
    items: [
      {
        to: '/more/career',
        title: 'Карьера',
        subtitle: 'Контакты и навыки',
        icon: <Briefcase size={20} />,
      },
      {
        to: '/more/analytics',
        title: 'Аналитика',
        subtitle: 'Сводка периода',
        icon: <BarChart3 size={20} />,
      },
    ],
  },
] as const

export function MorePage() {
  const navigate = useNavigate()
  const user = tgUser()

  return (
    <>
      <Header
        title="Ещё"
        subtitle={
          user?.username
            ? `@${user.username} · день · запись · обзор`
            : 'День · запись · обзор'
        }
      />
      <div className="space-y-6 px-4 pb-4">
        {sections.map((section) => (
          <section key={section.title}>
            <h2 className="mb-2 px-1 text-xs font-medium uppercase tracking-wide text-[var(--tg-theme-hint-color,#94a3b8)]">
              {section.title}
            </h2>
            <div className="space-y-2">
              {section.items.map((item) => (
                <ListRow
                  key={item.to}
                  title={item.title}
                  subtitle={item.subtitle}
                  icon={item.icon}
                  onClick={() => {
                    hapticLight()
                    navigate(item.to)
                  }}
                />
              ))}
            </div>
          </section>
        ))}
      </div>
    </>
  )
}
