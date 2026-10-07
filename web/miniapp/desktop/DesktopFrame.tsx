import type { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'
import {
  BarChart3,
  Bell,
  Calendar,
  FileText,
  HeartPulse,
  Home,
  Layers,
  ListChecks,
  Settings,
  Sparkles,
  Wallet,
} from 'lucide-react'
import { cn } from '@/lib/cn'

const items = [
  { to: '/', label: 'Главная', icon: Home, end: true },
  { to: '/spheres', label: 'Сферы', icon: Layers, end: false },
  { to: '/more/habits', label: 'Привычки', icon: Sparkles, end: true },
  { to: '/more/calendar', label: 'Календарь', icon: Calendar, end: true },
  { to: '/more/reminders', label: 'Напоминания', icon: Bell, end: true },
  { to: '/more/triage', label: 'Разбор', icon: ListChecks, end: true },
  { to: '/more/notes', label: 'Заметки', icon: FileText, end: true },
  { to: '/more/health', label: 'Здоровье', icon: HeartPulse, end: true },
  { to: '/more/debts', label: 'Долги', icon: Wallet, end: true },
  { to: '/more/analytics', label: 'Аналитика', icon: BarChart3, end: true },
  { to: '/more/settings', label: 'Настройки', icon: Settings, end: true },
] as const

export function DesktopFrame({ children }: { children: ReactNode }) {
  return (
    <div
      className="flex h-screen w-screen overflow-hidden"
      style={{ display: 'flex', width: '100vw', height: '100vh' }}
    >
      <aside className="flex h-full w-60 shrink-0 flex-col border-r border-white/10 bg-black/20">
        <div className="px-4 py-5">
          <div className="text-sm font-semibold tracking-wide">LifeOS</div>
          <div className="mt-1 text-xs text-[#94a3b8]">экраны на этом компьютере</div>
        </div>
        <nav className="min-h-0 flex-1 space-y-0.5 overflow-y-auto px-2 pb-4">
          {items.map(({ to, label, icon: Icon, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                cn(
                  'flex min-h-10 items-center gap-2.5 rounded-xl px-3 text-sm font-medium',
                  isActive
                    ? 'bg-white/10 text-[var(--tg-theme-button-color,#22c55e)]'
                    : 'text-[#cbd5e1] hover:bg-white/5',
                )
              }
            >
              <Icon size={18} />
              {label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <div className="min-w-0 flex-1 overflow-y-auto" style={{ flex: '1 1 auto', minWidth: 0 }}>
        <div className="mx-auto w-full max-w-3xl">{children}</div>
      </div>
    </div>
  )
}
