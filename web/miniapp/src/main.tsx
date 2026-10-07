import { renderApp } from './root'

renderApp('/app')

// After the module has started React. A classic script tag in the HTML
// runs first and, inside Telegram, postEvent can stall before this file
// is ever evaluated — the static "Загрузка LifeOS…" line never goes away.
setTimeout(() => {
  if (document.querySelector('script[data-lifeos-tg]')) return
  const s = document.createElement('script')
  s.src = '/app/telegram-web-app.js'
  s.async = true
  s.dataset.lifeosTg = '1'
  document.body.appendChild(s)
}, 0)
