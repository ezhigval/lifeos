import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { HashRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, useAuth } from '@/context/AuthContext'
import { initTelegram } from '@/lib/telegram'
import App from './App'
import './index.css'

initTelegram()

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
})

function Root() {
  const auth = useAuth()

  if (auth.status === 'loading') {
    return (
      <div className="flex min-h-full items-center justify-center p-8 text-[var(--tg-theme-hint-color,#94a3b8)]">
        Загрузка…
      </div>
    )
  }

  if (auth.status === 'error') {
    return (
      <div className="flex min-h-full flex-col items-center justify-center gap-2 p-8 text-center">
        <p className="text-lg font-medium">Не удалось войти</p>
        <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">{auth.message}</p>
      </div>
    )
  }

  return <App />
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <HashRouter>
          <Root />
        </HashRouter>
      </AuthProvider>
    </QueryClientProvider>
  </StrictMode>,
)
