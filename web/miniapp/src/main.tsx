import { Component, StrictMode, useEffect, useState, type ErrorInfo, type ReactNode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, useAuth } from '@/context/AuthContext'
import { freezeInitData, initTelegram } from '@/lib/telegram'
import App from './App'
import './index.css'

// Module reached the WebView. Cancel the HTML boot watchdog before React paints.
window.__LIFEOS_BOOT_OK__?.()

// Capture Telegram launch payload before the router can touch location.hash.
freezeInitData()
initTelegram()

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
      // Telegram WebView toggles focus on keyboard/sheets and would refetch everything.
      refetchOnWindowFocus: false,
    },
  },
})

class ErrorBoundary extends Component<{ children: ReactNode }, { error?: string }> {
  state: { error?: string } = {}

  static getDerivedStateFromError(err: Error) {
    return { error: err.message || 'Unknown UI error' }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('miniapp crashed', error, info)
  }

  render() {
    if (this.state.error) {
      const chunk =
        /dynamically imported module|loading chunk|failed to fetch|importing a module/i.test(
          this.state.error,
        )
      return (
        <div style={{ padding: 24, color: '#f8fafc', background: '#0f172a', minHeight: '100%' }}>
          <h1 style={{ fontSize: 18, marginBottom: 8 }}>Ошибка Mini App</h1>
          <p style={{ color: '#94a3b8', fontSize: 14 }}>
            {chunk
              ? 'Не удалось догрузить экран. Проверь сеть и открой ещё раз.'
              : this.state.error}
          </p>
          <button
            type="button"
            style={{
              marginTop: 16,
              borderRadius: 16,
              border: 'none',
              background: '#22c55e',
              color: '#fff',
              padding: '8px 16px',
              fontSize: 14,
              fontWeight: 500,
            }}
            onClick={() => window.location.reload()}
          >
            Повторить
          </button>
        </div>
      )
    }
    return this.props.children
  }
}

function Root() {
  const auth = useAuth()
  const [stuck, setStuck] = useState(false)

  useEffect(() => {
    if (auth.status !== 'loading') return
    const timer = window.setTimeout(() => setStuck(true), 6_000)
    return () => window.clearTimeout(timer)
  }, [auth.status])

  if (auth.status === 'loading') {
    return (
      <div
        style={{
          display: 'flex',
          minHeight: '100%',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 12,
          padding: 32,
          background: '#0f172a',
          color: '#94a3b8',
          textAlign: 'center',
        }}
      >
        <div>Загрузка…</div>
        {stuck ? (
          <button
            type="button"
            style={{
              borderRadius: 16,
              border: 'none',
              background: '#22c55e',
              color: '#fff',
              padding: '8px 16px',
              fontSize: 14,
              fontWeight: 500,
            }}
            onClick={() => window.location.reload()}
          >
            Повторить
          </button>
        ) : null}
      </div>
    )
  }

  if (auth.status === 'error') {
    return (
      <div
        style={{
          display: 'flex',
          minHeight: '100%',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 12,
          padding: 32,
          textAlign: 'center',
          background: '#0f172a',
          color: '#f8fafc',
        }}
      >
        <p style={{ fontSize: 18, fontWeight: 500, margin: 0 }}>Не удалось войти</p>
        <p style={{ fontSize: 14, color: '#94a3b8', margin: 0, maxWidth: 320 }}>{auth.message}</p>
        <button
          type="button"
          style={{
            marginTop: 8,
            borderRadius: 16,
            border: 'none',
            background: 'var(--tg-theme-button-color, #22c55e)',
            color: 'var(--tg-theme-button-text-color, #fff)',
            padding: '8px 16px',
            fontSize: 14,
            fontWeight: 500,
            cursor: 'pointer',
          }}
          onClick={() => window.location.reload()}
        >
          Повторить
        </button>
      </div>
    )
  }

  return <App />
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          {/* BrowserRouter keeps Telegram's #tgWebAppData=… hash intact.
              HashRouter would overwrite it and break initData signatures. */}
          <BrowserRouter basename="/app">
            <Root />
          </BrowserRouter>
        </AuthProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  </StrictMode>,
)
