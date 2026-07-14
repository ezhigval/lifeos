const STORAGE_KEY = 'lifeos.miniapp.session.v1'

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
  if (!canUseStorage()) return
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
  } catch {
    /* private mode / quota */
  }
}

export function clearSession(): void {
  if (!canUseStorage()) return
  try {
    window.localStorage.removeItem(STORAGE_KEY)
  } catch {
    /* ignore */
  }
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
