const STORAGE_KEY = 'lifeos.miniapp.session.v1'

declare global {
  interface Window {
    __LIFEOS_DESKTOP__?: boolean
  }
}

export type StoredSession = {
  accessToken: string
  /** Unix epoch milliseconds */
  expiresAt: number
  telegramId?: number
}

const EXPIRY_SKEW_MS = 60_000

function canUseStorage(): boolean {
  try {
    return typeof window !== 'undefined' && !!window.localStorage
  } catch {
    return false
  }
}

/** Decode JWT exp without verifying signature (server still verifies). */
export function jwtExpiresAtMs(token: string): number | null {
  try {
    const part = token.split('.')[1]
    if (!part) return null
    const json = atob(part.replace(/-/g, '+').replace(/_/g, '/'))
    const payload = JSON.parse(json) as { exp?: number }
    if (typeof payload.exp !== 'number' || payload.exp <= 0) return null
    return payload.exp * 1000
  } catch {
    return null
  }
}

export function loadSession(): StoredSession | null {
  if (!canUseStorage()) return null
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as StoredSession
    if (!parsed?.accessToken || typeof parsed.expiresAt !== 'number') return null
    return parsed
  } catch {
    return null
  }
}

export function saveSession(session: StoredSession): void {
  writeLocal(session)
  void syncDesktop(session)
}

/** Write the browser copy and wait until the desktop file is updated. */
export async function commitSession(session: StoredSession): Promise<void> {
  writeLocal(session)
  await syncDesktop(session)
}

export function clearSession(): void {
  if (canUseStorage()) {
    try {
      window.localStorage.removeItem(STORAGE_KEY)
    } catch {
      /* ignore */
    }
  }
  void syncDesktop(null)
}

function writeLocal(session: StoredSession): void {
  if (!canUseStorage()) return
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
  } catch {
    /* private mode / quota */
  }
}

function desktopApp(): boolean {
  return typeof window !== 'undefined' && window.__LIFEOS_DESKTOP__ === true
}

async function syncDesktop(session: StoredSession | null): Promise<void> {
  if (!desktopApp()) return
  try {
    if (!session) {
      await fetch('/desktop/session', { method: 'DELETE', signal: AbortSignal.timeout(2_000) })
      return
    }
    await fetch('/desktop/session', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(session),
      signal: AbortSignal.timeout(2_000),
    })
  } catch {
    /* localStorage already has the login; the verify proxy also writes the file */
  }
}

/** Mini App lives at /app/. The desktop window lives at /. */
export function homePath(): string {
  const base = import.meta.env.BASE_URL || '/'
  return base.endsWith('/') ? base : `${base}/`
}

export function isSessionFresh(session: StoredSession, now = Date.now()): boolean {
  return session.expiresAt - EXPIRY_SKEW_MS > now
}

export function sessionMatchesTelegram(
  session: StoredSession,
  telegramId: number | undefined,
): boolean {
  if (!telegramId || !session.telegramId) return true
  return session.telegramId === telegramId
}

export function buildSession(
  accessToken: string,
  expiresInSec: number | undefined,
  telegramId?: number,
): StoredSession {
  const fromJwt = jwtExpiresAtMs(accessToken)
  const fromTtl =
    typeof expiresInSec === 'number' && expiresInSec > 0
      ? Date.now() + expiresInSec * 1000
      : null
  return {
    accessToken,
    expiresAt: fromJwt ?? fromTtl ?? Date.now() + 24 * 60 * 60 * 1000,
    telegramId: telegramId && telegramId > 0 ? telegramId : undefined,
  }
}
