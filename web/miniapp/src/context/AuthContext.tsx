import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'
import {
  setAccessToken,
  setUnauthorizedHandler,
  authWithInitData,
  authWithDevCredentials,
} from '@/api/client'
import { getInitData, initTelegram, isTelegramEnv, telegramIdFromInitData, tgUser } from '@/lib/telegram'
import {
  buildSession,
  clearSession,
  isSessionFresh,
  loadSession,
  saveSession,
  sessionMatchesTelegram,
  type StoredSession,
} from '@/lib/session'

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

async function waitForInitData(maxMs: number): Promise<string> {
  const started = Date.now()
  let data = getInitData()
  while (!data && Date.now() - started < maxMs) {
    await new Promise((r) => setTimeout(r, 50))
    initTelegram()
    data = getInitData()
  }
  return data
}

function applySession(session: StoredSession) {
  setAccessToken(session.accessToken)
  saveSession(session)
}

function tryRestoreSession(): boolean {
  const session = loadSession()
  if (!session || !isSessionFresh(session)) {
    if (session) clearSession()
    return false
  }
  const tgId = telegramIdFromInitData() ?? tgUser()?.id
  if (!sessionMatchesTelegram(session, tgId)) {
    clearSession()
    return false
  }
  setAccessToken(session.accessToken)
  return true
}

async function loginWithInitData(initData: string): Promise<StoredSession> {
  const result = await withTimeout(
    authWithInitData(initData),
    AUTH_TIMEOUT_MS,
    'auth/telegram-webapp',
  )
  // Prefer server-echoed telegram_id (parsed from HMAC-verified initData.user.id).
  const telegramId =
    result.telegramId ?? telegramIdFromInitData(initData) ?? tgUser()?.id
  const session = buildSession(result.accessToken, result.expiresIn, telegramId)
  applySession(session)
  return session
}

async function loginWithDev(): Promise<StoredSession | null> {
  const apiKey = import.meta.env.VITE_DEV_API_KEY as string | undefined
  const telegramId = Number(import.meta.env.VITE_DEV_TELEGRAM_ID)
  if (!apiKey || !(telegramId > 0)) return null
  const result = await withTimeout(
    authWithDevCredentials(apiKey, telegramId),
    AUTH_TIMEOUT_MS,
    'auth/token',
  )
  const session = buildSession(result.accessToken, result.expiresIn, telegramId)
  applySession(session)
  return session
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ status: 'loading' })

  useEffect(() => {
    let cancelled = false
    let refreshInFlight: Promise<boolean> | null = null

    async function refreshFromInitData(): Promise<boolean> {
      if (refreshInFlight) return refreshInFlight
      refreshInFlight = (async () => {
        try {
          initTelegram()
          const initData = getInitData() || (await waitForInitData(800))
          if (!initData) return false
          await loginWithInitData(initData)
          return true
        } catch (err) {
          console.warn('silent re-auth failed', err)
          clearSession()
          setAccessToken(null)
          return false
        } finally {
          refreshInFlight = null
        }
      })()
      return refreshInFlight
    }

    setUnauthorizedHandler(refreshFromInitData)

    async function login() {
      try {
        initTelegram()

        // 1) Prefer persisted JWT — do NOT touch initData if session is fresh.
        if (tryRestoreSession()) {
          if (!cancelled) setState({ status: 'ready' })
          return
        }

        // 2) One-shot Telegram bootstrap when we have no/expired session.
        const initData = await waitForInitData(3_000)
        if (initData) {
          await loginWithInitData(initData)
          if (!cancelled) setState({ status: 'ready' })
          return
        }

        // 3) Dev fallback outside Telegram.
        const dev = await loginWithDev()
        if (dev) {
          if (!cancelled) setState({ status: 'ready' })
          return
        }

        if (!cancelled) {
          setState({
            status: 'error',
            message: isTelegramEnv()
              ? 'Нет initData. Закрой окно и открой синей кнопкой «📱 Открыть Mini App» в чате (или Menu → Mini App). Reply-клавиатура и ссылка из текста не передают initData.'
              : 'Открой из Telegram или задай VITE_DEV_API_KEY и VITE_DEV_TELEGRAM_ID',
          })
        }
      } catch (e) {
        console.error('miniapp auth failed', e)
        try {
          const dev = await loginWithDev()
          if (dev) {
            if (!cancelled) setState({ status: 'ready' })
            return
          }
        } catch {
          /* fall through */
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
      setUnauthorizedHandler(null)
    }
  }, [])

  return <AuthContext.Provider value={state}>{children}</AuthContext.Provider>
}

export function useAuth() {
  return useContext(AuthContext)
}
