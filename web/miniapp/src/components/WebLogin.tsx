import { useState, type CSSProperties, type FormEvent } from 'react'
import {
  requestTelegramLoginCode,
  setAccessToken,
  verifyTelegramLoginCode,
} from '@/api/client'
import { buildSession, saveSession } from '@/lib/session'

const field: CSSProperties = {
  width: '100%',
  maxWidth: 320,
  boxSizing: 'border-box',
  borderRadius: 12,
  border: '1px solid #334155',
  background: '#1e293b',
  color: '#f8fafc',
  padding: '12px 14px',
  fontSize: 16,
}

const button: CSSProperties = {
  width: '100%',
  maxWidth: 320,
  borderRadius: 16,
  border: 'none',
  background: 'var(--tg-theme-button-color, #22c55e)',
  color: 'var(--tg-theme-button-text-color, #fff)',
  padding: '12px 16px',
  fontSize: 15,
  fontWeight: 600,
  cursor: 'pointer',
}

export function WebLogin() {
  const [step, setStep] = useState<'nick' | 'code'>('nick')
  const [username, setUsername] = useState('')
  const [code, setCode] = useState('')
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

  async function sendCode(e: FormEvent) {
    e.preventDefault()
    setError('')
    setPending(true)
    try {
      await requestTelegramLoginCode(username.trim())
      setStep('code')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось отправить код')
    } finally {
      setPending(false)
    }
  }

  async function verify(e: FormEvent) {
    e.preventDefault()
    setError('')
    setPending(true)
    try {
      const result = await verifyTelegramLoginCode(username.trim(), code.trim())
      setAccessToken(result.accessToken)
      saveSession(buildSession(result.accessToken, result.expiresIn, result.telegramId))
      // Mini App is mounted at /app/. The desktop window is the site root.
      const home = window.location.pathname.startsWith('/app') ? '/app/' : '/'
      window.location.assign(home)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось войти')
      setPending(false)
    }
  }

  return (
    <div
      style={{
        display: 'flex',
        minHeight: '100%',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 24,
        background: '#0f172a',
        color: '#f8fafc',
      }}
    >
      <form
        onSubmit={step === 'nick' ? sendCode : verify}
        style={{
          display: 'flex',
          flexDirection: 'column',
          gap: 12,
          width: '100%',
          maxWidth: 320,
        }}
      >
        <h1 style={{ fontSize: 22, fontWeight: 600, margin: 0 }}>Вход в LifeOS</h1>
        <p style={{ fontSize: 14, color: '#94a3b8', margin: 0, lineHeight: 1.45 }}>
          {step === 'nick'
            ? 'Ник Telegram. Бот пришлёт код в чат. Если бот ещё не знаком с тобой — сначала нажми /start.'
            : `Код отправлен в Telegram для @${username.replace(/^@/, '')}.`}
        </p>
        {step === 'nick' ? (
          <input
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="username"
            autoComplete="username"
            autoCapitalize="none"
            autoCorrect="off"
            spellCheck={false}
            required
            style={field}
          />
        ) : (
          <input
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
            placeholder="код из Telegram"
            inputMode="numeric"
            autoComplete="one-time-code"
            required
            style={{ ...field, letterSpacing: 4 }}
          />
        )}
        {error ? (
          <p style={{ fontSize: 14, color: '#fca5a5', margin: 0 }}>{error}</p>
        ) : null}
        <button type="submit" disabled={pending} style={{ ...button, opacity: pending ? 0.7 : 1 }}>
          {pending ? 'Секунду…' : step === 'nick' ? 'Получить код' : 'Войти'}
        </button>
        {step === 'code' ? (
          <button
            type="button"
            onClick={() => {
              setStep('nick')
              setCode('')
              setError('')
            }}
            style={{
              background: 'transparent',
              border: 'none',
              color: '#94a3b8',
              fontSize: 14,
              cursor: 'pointer',
              padding: 0,
            }}
          >
            Другой ник
          </button>
        ) : null}
      </form>
    </div>
  )
}
