import { Component, useEffect, type ErrorInfo, type ReactNode } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { BottomNav } from '@/components/layout/BottomNav'
import { useDesktopFrame } from '@/components/layout/shell'
import { bootMark } from '@/lib/bootTiming'
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

function fallbackFor(pathname: string): string {
  if (pathname.startsWith('/more')) return '/more'
  if (pathname.startsWith('/tasks')) return '/'
  return '/spheres'
}

function screenErrorText(message: string): string {
  if (/chunk|empty|dynamic|failed to fetch|loading chunk|imported module/i.test(message)) {
    return 'Не удалось открыть раздел. Нажми «Обновить» или вернись назад.'
  }
  return message || 'Не удалось открыть раздел'
}

function ScreenActions({ fallback }: { fallback: string }) {
  const navigate = useNavigate()

  function goBack() {
    const state = window.history.state as { idx?: number } | null
    if (typeof state?.idx === 'number' && state.idx > 0) {
      navigate(-1)
      return
    }
    navigate(fallback, { replace: true })
  }

  return (
    <div className="sticky top-0 z-40 flex items-center justify-between gap-2 bg-[var(--tg-theme-bg-color,#0f172a)]/95 px-3 py-2 backdrop-blur-md">
      <button
        type="button"
        onClick={goBack}
        className="rounded-full px-3 py-1.5 text-sm font-medium text-[var(--tg-theme-link-color,#22c55e)]"
      >
        Назад
      </button>
      <button
        type="button"
        onClick={() => window.location.reload()}
        className="rounded-full px-3 py-1.5 text-sm font-medium text-[var(--tg-theme-hint-color,#94a3b8)]"
      >
        Обновить
      </button>
    </div>
  )
}

type BoundaryProps = { resetKey: string; fallback: string; children: ReactNode }
type BoundaryState = { error?: string }

class SectionBoundary extends Component<BoundaryProps, BoundaryState> {
  state: BoundaryState = {}

  static getDerivedStateFromError(err: Error): BoundaryState {
    return { error: err.message || 'Не удалось открыть раздел' }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('section crashed', error, info)
  }

  componentDidUpdate(prev: BoundaryProps) {
    if (prev.resetKey !== this.props.resetKey && this.state.error) {
      this.setState({ error: undefined })
    }
  }

  render() {
    if (!this.state.error) return this.props.children
    return (
      <div className="px-4 py-8 text-center">
        <p className="text-base font-medium">Раздел не открылся</p>
        <p className="mx-auto mt-2 max-w-xs text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
          {screenErrorText(this.state.error)}
        </p>
      </div>
    )
  }
}

export function AppShell() {
  const { pathname: rawPath } = useLocation()
  const pathname = normalizePath(rawPath)
  const nested = isNestedPath(pathname)
  const frame = useDesktopFrame()
  const fallback = fallbackFor(pathname)
  useTelegramBackButton(frame ? false : nested, fallback)
  useEffect(() => {
    bootMark('shell')
  }, [])

  const page = (
    <SectionBoundary resetKey={pathname} fallback={fallback}>
      <Outlet />
    </SectionBoundary>
  )

  if (frame) {
    const Frame = frame
    return <Frame>{page}</Frame>
  }

  return (
    <div className="mx-auto min-h-full max-w-lg pb-24">
      {nested ? <ScreenActions fallback={fallback} /> : null}
      {page}
      <BottomNav />
    </div>
  )
}
