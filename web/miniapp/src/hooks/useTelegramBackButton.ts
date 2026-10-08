import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { hideTelegramBackButton, showTelegramBackButton } from '@/lib/telegram'

/** React Router stores stack index on history.state.idx; history.length is unreliable in TG WebView. */
function canGoBackInApp(): boolean {
  const state = window.history.state as { idx?: number } | null
  if (typeof state?.idx === 'number') return state.idx > 0
  return false
}

/** Shows Telegram BackButton when `active`; navigates back on press. */
export function useTelegramBackButton(active: boolean, fallbackTo = '/') {
  const navigate = useNavigate()

  useEffect(() => {
    if (!active) {
      hideTelegramBackButton()
      return
    }

    const onBack = () => {
      if (canGoBackInApp()) {
        navigate(-1)
      } else {
        navigate(fallbackTo, { replace: true })
      }
    }

    // telegram-web-app.js is injected after React mounts. The first show()
    // no-ops until that script exists, so keep trying briefly.
    let cleanup = showTelegramBackButton(onBack)
    const timer = window.setInterval(() => {
      const btn = window.Telegram?.WebApp?.BackButton
      if (!btn?.show || !btn?.onClick) return
      window.clearInterval(timer)
      cleanup()
      cleanup = showTelegramBackButton(onBack)
    }, 300)
    const stop = window.setTimeout(() => window.clearInterval(timer), 3000)

    return () => {
      window.clearInterval(timer)
      window.clearTimeout(stop)
      cleanup()
    }
  }, [active, fallbackTo, navigate])
}
