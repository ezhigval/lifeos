import { renderApp } from './root'
import { bootMark } from '@/lib/bootTiming'
import './index.css'

bootMark('module')
window.__LIFEOS_BOOT_OK__?.()

renderApp('/app')

// After the module has started React. A classic script tag in the HTML
// runs first and, inside Telegram, postEvent can stall before this file
// is ever evaluated — the static "Загрузка LifeOS…" line never goes away.
setTimeout(() => {
  if (document.querySelector('script[data-lifeos-tg], script[src*="telegram-web-app"]')) return
  const s = document.createElement('script')
  s.src = '/app/telegram-web-app.js'
  s.async = true
  s.dataset.lifeosTg = '1'
  document.body.appendChild(s)
}, 0)

function registerStaticCache() {
  if (!import.meta.env.PROD || !('serviceWorker' in navigator)) return
  const run = () => {
    const base = import.meta.env.BASE_URL
    navigator.serviceWorker.register(`${base}sw.js`, { scope: base }).catch(() => {
      /* private mode / unsupported WebView */
    })
  }
  if (typeof window.requestIdleCallback === 'function') {
    window.requestIdleCallback(run, { timeout: 2000 })
  } else {
    window.setTimeout(run, 1500)
  }
}

registerStaticCache()
