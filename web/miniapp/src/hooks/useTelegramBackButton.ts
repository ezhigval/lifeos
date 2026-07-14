import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { hideTelegramBackButton, showTelegramBackButton } from '@/lib/telegram'

/** Shows Telegram BackButton when `active`; navigates back on press. */
export function useTelegramBackButton(active: boolean, fallbackTo = '/') {
  const navigate = useNavigate()

  useEffect(() => {
    if (!active) {
      hideTelegramBackButton()
      return
    }

    const cleanup = showTelegramBackButton(() => {
      if (window.history.length > 1) {
        navigate(-1)
      } else {
        navigate(fallbackTo, { replace: true })
      }
    })

    return cleanup
  }, [active, fallbackTo, navigate])
}
