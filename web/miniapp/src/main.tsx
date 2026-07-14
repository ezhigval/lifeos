import { Component, StrictMode, type ErrorInfo, type ReactNode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, useAuth } from '@/context/AuthContext'
import { freezeInitData, initTelegram } from '@/lib/telegram'
import App from './App'
import './index.css'

// Capture Telegram launch payload before the router can touch location.hash.
freezeInitData()
initTelegram()

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
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
      return (
        <div style={{ padding: 24, color: '#f8fafc', background: '#0f172a', minHeight: '100%' }}>
          <h1 style={{ fontSize: 18, marginBottom: 8 }}>Ошибка Mini App</h1>
          <p style={{ color: '#94a3b8', fontSize: 14 }}>{this.state.error}</p>
        </div>
      )
    }
    return this.props.children
  }
}

function Root() {
  const auth = useAuth()

  if (auth.status === 'loading') {
    return (
      <div
        style={{
          display: 'flex',
          minHeight: '100%',
          alignItems: 'center',
          justifyContent: 'center',
          padding: 32,
          background: '#0f172a',
          color: '#94a3b8',
        }}
      >
        Загрузка…
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
          gap: 8,
          padding: 32,
          textAlign: 'center',
          background: '#0f172a',
          color: '#f8fafc',
        }}
      >
        <p style={{ fontSize: 18, fontWeight: 500, margin: 0 }}>Не удалось войти</p>
        <p style={{ fontSize: 14, color: '#94a3b8', margin: 0 }}>{auth.message}</p>
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
