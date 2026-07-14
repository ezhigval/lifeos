import { Outlet } from 'react-router-dom'
import { BottomNav } from '@/components/layout/BottomNav'

export function AppShell() {
  return (
    <div className="mx-auto min-h-full max-w-lg pb-24">
      <Outlet />
      <BottomNav />
    </div>
  )
}
