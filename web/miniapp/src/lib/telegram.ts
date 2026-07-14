type TelegramUser = {
  id?: number
  first_name?: string
  last_name?: string
  username?: string
  language_code?: string
}

type TelegramWebApp = {
  initData?: string
  initDataUnsafe?: { user?: TelegramUser }
  platform?: string
  version?: string
  ready?: () => void
  expand?: () => void
  setHeaderColor?: (color: string) => void
  setBackgroundColor?: (color: string) => void
  HapticFeedback?: {
    impactOccurred?: (style: string) => void
    notificationOccurred?: (type: string) => void
  }
}

declare global {
  interface Window {
    Telegram?: { WebApp?: TelegramWebApp }
    __LIFEOS_INIT_DATA__?: string
  }
}

function getWebApp(): TelegramWebApp | undefined {
  if (typeof window === 'undefined') return undefined
  return window.Telegram?.WebApp
}

/**
 * Capture initData ASAP. Telegram puts tgWebAppData into location.hash;
 * HashRouter (or any hash rewrite) can wipe it on later navigations / reloads.
 * We freeze the value once so auth always sees the original signed payload.
 */
export function freezeInitData(): string {
  if (typeof window === 'undefined') return ''
  if (typeof window.__LIFEOS_INIT_DATA__ === 'string' && window.__LIFEOS_INIT_DATA__) {
    return window.__LIFEOS_INIT_DATA__
  }

  let raw = ''
  try {
    raw = getWebApp()?.initData || ''
  } catch {
    raw = ''
  }

  // Fallback: read tgWebAppData from the launch hash before routers touch it.
  if (!raw) {
    try {
      const hash = window.location.hash.startsWith('#')
        ? window.location.hash.slice(1)
        : window.location.hash
      const params = new URLSearchParams(hash.includes('=') ? hash : '')
      raw = params.get('tgWebAppData') || ''
    } catch {
      /* ignore */
    }
  }

  if (raw) {
    window.__LIFEOS_INIT_DATA__ = raw
  }
  return raw
}

export function initTelegram() {
  freezeInitData()
  try {
    const wa = getWebApp()
    if (!wa) return
    wa.ready?.()
    wa.expand?.()
    try {
      wa.setHeaderColor?.('secondary_bg_color')
    } catch {
      /* older clients */
    }
    try {
      wa.setBackgroundColor?.('bg_color')
    } catch {
      /* older clients */
    }
  } catch (err) {
    console.warn('telegram init failed', err)
  }
}

export function getInitData(): string {
  return freezeInitData()
}

export function isTelegramEnv(): boolean {
  const wa = getWebApp()
  if (getInitData()) return true
  if (!wa) return false
  return wa.platform !== undefined && wa.platform !== 'unknown'
}

export function hapticLight() {
  try {
    getWebApp()?.HapticFeedback?.impactOccurred?.('light')
  } catch {
    /* ignore */
  }
}

export function hapticSuccess() {
  try {
    getWebApp()?.HapticFeedback?.notificationOccurred?.('success')
  } catch {
    /* ignore */
  }
}

export function tgUser(): TelegramUser | undefined {
  try {
    return getWebApp()?.initDataUnsafe?.user
  } catch {
    return undefined
  }
}
