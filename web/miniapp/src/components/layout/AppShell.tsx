import { Outlet, useLocation } from 'react-router-dom'
import { BottomNav } from '@/components/layout/BottomNav'
import { useTelegramBackButton } from '@/hooks/useTelegramBackButton'

const ROOT_PATHS = new Set(['/', '/spheres', '/more'])

function normalizePath(pathname: string): string {
  if (pathname.length > 1 && pathname.endsWith('/')) {
    return pathname.slice(0, -1)
  }
  return pathname || '/'
}

function isNestedPath(pathname: string): boolean {
  const path = normalizePath(pathname)
  if (ROOT_PATHS.has(path)) return false
  return (
    path.startsWith('/spheres/') ||
    path.startsWith('/more/') ||
    path.startsWith('/tasks/')
  )
}

export function AppShell() {
  const { pathname: rawPath } = useLocation()
  const pathname = normalizePath(rawPath)
  const nested = isNestedPath(pathname)
  // Nested /more/* (habits, calendar, settings, …) → hub; tasks → Home; else spheres tree.
  const fallback = pathname.startsWith('/more')
    ? '/more'
    : pathname.startsWith('/tasks')
      ? '/'
      : '/spheres'
  useTelegramBackButton(nested, fallback)

  return (
    <div className="mx-auto min-h-full max-w-lg pb-24">
      <Outlet />
      <BottomNav />
    </div>
  )
}
