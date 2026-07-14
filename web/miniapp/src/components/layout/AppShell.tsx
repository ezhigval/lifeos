import { Outlet, useLocation } from 'react-router-dom'
import { BottomNav } from '@/components/layout/BottomNav'
import { useTelegramBackButton } from '@/hooks/useTelegramBackButton'

const ROOT_PATHS = new Set(['/', '/spheres', '/more'])

function isNestedPath(pathname: string): boolean {
  if (ROOT_PATHS.has(pathname)) return false
  return (
    pathname.startsWith('/spheres/') ||
    pathname.startsWith('/more/') ||
    pathname.startsWith('/tasks/')
  )
}

export function AppShell() {
  const { pathname } = useLocation()
  const nested = isNestedPath(pathname)
  const fallback = pathname.startsWith('/more') ? '/more' : '/spheres'
  useTelegramBackButton(nested, fallback)

  return (
    <div className="mx-auto min-h-full max-w-lg pb-24">
      <Outlet />
      <BottomNav />
    </div>
  )
}
