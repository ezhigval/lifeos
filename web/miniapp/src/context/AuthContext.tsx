import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'
import { setAccessToken, authWithInitData, authWithDevCredentials } from '@/api/client'
import { getInitData, isTelegramEnv } from '@/lib/telegram'

type AuthState =
  | { status: 'loading' }
  | { status: 'ready' }
  | { status: 'error'; message: string }

const AuthContext = createContext<AuthState>({ status: 'loading' })

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ status: 'loading' })

  useEffect(() => {
    let cancelled = false

    async function login() {
      try {
        const initData = getInitData()
        if (initData) {
          const token = await authWithInitData(initData)
          if (!cancelled) {
            setAccessToken(token)
            setState({ status: 'ready' })
          }
          return
        }

        const apiKey = import.meta.env.VITE_DEV_API_KEY as string | undefined
        const telegramId = Number(import.meta.env.VITE_DEV_TELEGRAM_ID)
        if (apiKey && telegramId > 0) {
          const token = await authWithDevCredentials(apiKey, telegramId)
          if (!cancelled) {
            setAccessToken(token)
            setState({ status: 'ready' })
          }
          return
        }

        if (!cancelled) {
          setState({
            status: 'error',
            message: isTelegramEnv()
              ? 'Не удалось авторизоваться через Telegram'
              : 'Открой из Telegram или задай VITE_DEV_API_KEY и VITE_DEV_TELEGRAM_ID',
          })
        }
      } catch (e) {
        const apiKey = import.meta.env.VITE_DEV_API_KEY as string | undefined
        const telegramId = Number(import.meta.env.VITE_DEV_TELEGRAM_ID)
        if (apiKey && telegramId > 0) {
          try {
            const token = await authWithDevCredentials(apiKey, telegramId)
            if (!cancelled) {
              setAccessToken(token)
              setState({ status: 'ready' })
            }
            return
          } catch {
            /* fall through */
          }
        }
        if (!cancelled) {
          setState({
            status: 'error',
            message: e instanceof Error ? e.message : 'Ошибка авторизации',
          })
        }
      }
    }

    login()
    return () => {
      cancelled = true
    }
  }, [])

  return <AuthContext.Provider value={state}>{children}</AuthContext.Provider>
}

export function useAuth() {
  return useContext(AuthContext)
}
