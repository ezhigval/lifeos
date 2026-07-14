import WebApp from '@twa-dev/sdk'

export function initTelegram() {
  WebApp.ready()
  WebApp.expand()
  WebApp.setHeaderColor('secondary_bg_color')
  WebApp.setBackgroundColor('bg_color')
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

export function tgUser() {
  return WebApp.initDataUnsafe?.user
}
