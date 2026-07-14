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

function getWebApp(): TelegramWebApp | undefined {
  if (typeof window === 'undefined') return undefined
  return (window as Window & { Telegram?: { WebApp?: TelegramWebApp } }).Telegram?.WebApp
}

export function initTelegram() {
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
  try {
    return getWebApp()?.initData || ''
  } catch {
    return ''
  }
}

export function isTelegramEnv(): boolean {
  const wa = getWebApp()
  if (!wa) return false
  return Boolean(wa.initData) || (wa.platform !== undefined && wa.platform !== 'unknown')
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
