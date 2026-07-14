import WebAppSDK from '@twa-dev/sdk'

type TgWebApp = typeof WebAppSDK

function getWebApp(): TgWebApp | null {
  try {
    if (WebAppSDK && typeof (WebAppSDK as { ready?: unknown }).ready === 'function') {
      return WebAppSDK
    }
  } catch {
    /* ignore */
  }
  const fromWindow = (window as unknown as { Telegram?: { WebApp?: TgWebApp } }).Telegram
    ?.WebApp
  if (fromWindow && typeof fromWindow.ready === 'function') {
    return fromWindow
  }
  return null
}

export function initTelegram() {
  const wa = getWebApp()
  if (!wa) return
  try {
    wa.ready()
    wa.expand?.()
    wa.setHeaderColor?.('secondary_bg_color')
    wa.setBackgroundColor?.('bg_color')
  } catch {
    /* outside Telegram or older client */
  }
}

export function getInitData(): string {
  return getWebApp()?.initData || ''
}

export function isTelegramEnv(): boolean {
  const wa = getWebApp()
  if (!wa) return false
  return Boolean(wa.initData || (wa.platform && wa.platform !== 'unknown'))
}

export function hapticLight() {
  getWebApp()?.HapticFeedback?.impactOccurred('light')
}

export function hapticSuccess() {
  getWebApp()?.HapticFeedback?.notificationOccurred('success')
}

export function hapticError() {
  getWebApp()?.HapticFeedback?.notificationOccurred('error')
}

export function hapticWarning() {
  getWebApp()?.HapticFeedback?.notificationOccurred('warning')
}

export function tgUser() {
  return getWebApp()?.initDataUnsafe?.user
}

export function showTelegramBackButton(onClick: () => void) {
  const btn = getWebApp()?.BackButton
  if (!btn) return () => undefined
  btn.onClick(onClick)
  btn.show()
  return () => {
    btn.offClick(onClick)
    btn.hide()
  }
}

export function hideTelegramBackButton() {
  getWebApp()?.BackButton?.hide()
}

export async function confirmAction(message: string): Promise<boolean> {
  const wa = getWebApp()
  if (wa && typeof wa.showConfirm === 'function') {
    return new Promise((resolve) => {
      wa.showConfirm(message, (ok) => resolve(Boolean(ok)))
    })
  }
  return window.confirm(message)
}
