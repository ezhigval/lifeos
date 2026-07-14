import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'
import { setAccessToken, authWithInitData, authWithDevCredentials } from '@/api/client'
import { getInitData, initTelegram, isTelegramEnv } from '@/lib/telegram'

type AuthState =
  | { status: 'loading' }
  | { status: 'ready' }
  | { status: 'error'; message: string }

const AuthContext = createContext<AuthState>({ status: 'loading' })

const AUTH_TIMEOUT_MS = 12_000

async function withTimeout<T>(p: Promise<T>, ms: number, label: string): Promise<T> {
  let timer: ReturnType<typeof setTimeout> | undefined
  try {
    return await Promise.race([
      p,
      new Promise<T>((_, reject) => {
        timer = setTimeout(() => reject(new Error(`${label} timeout`)), ms)
      }),
    ])
  } finally {
    if (timer) clearTimeout(timer)
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ status: 'loading' })

  useEffect(() => {
    let cancelled = false

    async function login() {
      try {
        initTelegram()
        // Give Telegram WebView a tick to populate initData after ready().
        await new Promise((r) => setTimeout(r, 50))

        const initData = getInitData()
        if (initData) {
          const token = await withTimeout(
            authWithInitData(initData),
            AUTH_TIMEOUT_MS,
            'auth/telegram-webapp',
          )
          if (!cancelled) {
            setAccessToken(token)
            setState({ status: 'ready' })
          }
          return
        }

        const apiKey = import.meta.env.VITE_DEV_API_KEY as string | undefined
        const telegramId = Number(import.meta.env.VITE_DEV_TELEGRAM_ID)
        if (apiKey && telegramId > 0) {
          const token = await withTimeout(
            authWithDevCredentials(apiKey, telegramId),
            AUTH_TIMEOUT_MS,
            'auth/token',
          )
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
              ? 'Telegram не передал initData. Закрой Mini App и открой снова кнопкой «📱 Mini App».'
              : 'Открой из Telegram или задай VITE_DEV_API_KEY и VITE_DEV_TELEGRAM_ID',
          })
        }
      } catch (e) {
        console.error('miniapp auth failed', e)
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

    void login()
    return () => {
      cancelled = true
    }
  }, [])

  return <AuthContext.Provider value={state}>{children}</AuthContext.Provider>
}

export function useAuth() {
  return useContext(AuthContext)
}
