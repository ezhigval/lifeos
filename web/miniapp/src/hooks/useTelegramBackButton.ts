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

    const cleanup = showTelegramBackButton(() => {
      if (canGoBackInApp()) {
        navigate(-1)
      } else {
        navigate(fallbackTo, { replace: true })
      }
    })

    return cleanup
  }, [active, fallbackTo, navigate])
}
