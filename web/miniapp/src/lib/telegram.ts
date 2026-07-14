import WebApp from '@twa-dev/sdk'

export function initTelegram() {
  WebApp.ready()
  WebApp.expand()
  try {
    WebApp.setHeaderColor('secondary_bg_color')
    WebApp.setBackgroundColor('bg_color')
  } catch {
    /* older clients */
  }
}

export function getInitData(): string {
  return WebApp.initData || ''
}

export function isTelegramEnv(): boolean {
  return Boolean(WebApp.initData || WebApp.platform !== 'unknown')
}

export function hapticLight() {
  WebApp.HapticFeedback?.impactOccurred('light')
}

export function hapticSuccess() {
  WebApp.HapticFeedback?.notificationOccurred('success')
}

export function hapticError() {
  WebApp.HapticFeedback?.notificationOccurred('error')
}

export function hapticWarning() {
  WebApp.HapticFeedback?.notificationOccurred('warning')
}

export function tgUser() {
  return WebApp.initDataUnsafe?.user
}

export function showTelegramBackButton(onClick: () => void) {
  const btn = WebApp.BackButton
  if (!btn) return () => undefined
  btn.onClick(onClick)
  btn.show()
  return () => {
    btn.offClick(onClick)
    btn.hide()
  }
}

export function hideTelegramBackButton() {
  WebApp.BackButton?.hide()
}

export async function confirmAction(message: string): Promise<boolean> {
  if (typeof WebApp.showConfirm === 'function') {
    return new Promise((resolve) => {
      WebApp.showConfirm(message, (ok) => resolve(Boolean(ok)))
    })
  }
  return window.confirm(message)
}
