import { NavLink } from 'react-router-dom'
import { Home, Layers } from 'lucide-react'
import { cn } from '@/lib/cn'

export function BottomNav() {
  return (
    <nav
      className={cn(
        'fixed bottom-0 left-0 right-0 z-40 border-t border-white/5',
        'bg-[var(--tg-theme-bg-color,#0f172a)]/95 backdrop-blur-md',
        'pb-[max(0.5rem,env(safe-area-inset-bottom))]',
      )}
    >
      <div className="mx-auto grid max-w-lg grid-cols-2 gap-1 px-4 py-2">
        <NavLink
          to="/"
          className={({ isActive }) =>
            cn(
              'flex flex-col items-center gap-1 rounded-2xl py-2 text-xs font-medium transition',
              isActive
                ? 'text-[var(--tg-theme-button-color,#22c55e)]'
                : 'text-[var(--tg-theme-hint-color,#94a3b8)]',
            )
          }
        >
          <Home size={22} />
          Главная
        </NavLink>
        <NavLink
          to="/spheres"
          className={({ isActive }) =>
            cn(
              'flex flex-col items-center gap-1 rounded-2xl py-2 text-xs font-medium transition',
              isActive
                ? 'text-[var(--tg-theme-button-color,#22c55e)]'
                : 'text-[var(--tg-theme-hint-color,#94a3b8)]',
            )
          }
        >
          <Layers size={22} />
          Сферы
        </NavLink>
      </div>
    </nav>
  )
}
