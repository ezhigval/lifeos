import { NavLink } from 'react-router-dom'
import { Home, Layers, Menu } from 'lucide-react'
import { cn } from '@/lib/cn'

const tabs = [
  { to: '/', label: 'Главная', icon: Home, end: true },
  { to: '/spheres', label: 'Сферы', icon: Layers, end: false },
  { to: '/more', label: 'Ещё', icon: Menu, end: false },
] as const

export function BottomNav() {
  return (
    <nav
      className={cn(
        'fixed bottom-0 left-0 right-0 z-40 border-t border-white/5',
        'bg-[var(--tg-theme-bg-color,#0f172a)]/95 backdrop-blur-md',
        'pb-[max(0.5rem,env(safe-area-inset-bottom))]',
      )}
    >
      <div className="mx-auto grid max-w-lg grid-cols-3 gap-1 px-3 py-2">
        {tabs.map(({ to, label, icon: Icon, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              cn(
                'flex min-h-11 flex-col items-center justify-center gap-0.5 rounded-2xl py-1.5 text-xs font-medium transition-colors',
                isActive
                  ? 'text-[var(--tg-theme-button-color,#22c55e)]'
                  : 'text-[var(--tg-theme-hint-color,#94a3b8)]',
              )
            }
          >
            <Icon size={22} />
            {label}
          </NavLink>
        ))}
      </div>
    </nav>
  )
}
