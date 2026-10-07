import { isSessionFresh, loadSession, saveSession, type StoredSession } from '@/lib/session'

export function markDesktopApp(): void {
  window.__LIFEOS_DESKTOP__ = true
}

/** Copy a still-valid login from the desktop file into localStorage before React boots. */
export async function hydrateDesktopSession(): Promise<void> {
  const local = loadSession()
  if (local && isSessionFresh(local)) return
  try {
    const res = await fetch('/desktop/session', { signal: AbortSignal.timeout(2_000) })
    if (!res.ok) return
    const data = (await res.json()) as StoredSession
    if (!data?.accessToken || typeof data.expiresAt !== 'number') return
    const session: StoredSession = {
      accessToken: data.accessToken,
      expiresAt: data.expiresAt,
      telegramId: typeof data.telegramId === 'number' && data.telegramId > 0 ? data.telegramId : undefined,
    }
    if (!isSessionFresh(session)) return
    saveSession(session)
  } catch {
    /* no file yet — the login screen stays */
  }
}
