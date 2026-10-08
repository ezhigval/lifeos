import { useEffect, useRef, useState, type FormEvent } from 'react'
import {
  ApiClientError,
  assistantChat,
  type AssistantChatTurn,
} from '@/api/client'

const STORAGE_KEY = 'lifeos.desktop.chat.v1'
const HISTORY_LIMIT = 24

type Bubble = {
  id: string
  role: 'user' | 'assistant'
  text: string
  tools?: string[]
}

type StoredChat = {
  history: AssistantChatTurn[]
  bubbles: Bubble[]
}

function loadStored(): StoredChat {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (!raw) return { history: [], bubbles: [] }
    const parsed = JSON.parse(raw) as StoredChat
    if (!Array.isArray(parsed.history) || !Array.isArray(parsed.bubbles)) {
      return { history: [], bubbles: [] }
    }
    return parsed
  } catch {
    return { history: [], bubbles: [] }
  }
}

function remember(history: AssistantChatTurn[], bubbles: Bubble[]) {
  try {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify({ history, bubbles }))
  } catch {
    /* quota */
  }
}

function bubblesFromHistory(history: AssistantChatTurn[]): Bubble[] {
  const out: Bubble[] = []
  for (const turn of history) {
    const text = turn.content?.trim()
    if ((turn.role === 'user' || turn.role === 'assistant') && text) {
      out.push({ id: `${out.length}-${turn.role}`, role: turn.role, text })
    }
  }
  return out
}

function chatError(err: unknown): string {
  if (err instanceof ApiClientError) {
    if (err.status === 501) {
      return 'Ассистент на сервере выключен. Нужны LIFEOS_LLM_ENABLED и LIFEOS_LLM_AGENT_ENABLED.'
    }
    if (err.status === 502 || err.status === 0) return 'Ассистент не ответил. Попробуй ещё раз.'
    if (err.status === 401) return 'Сессия истекла. Войди снова кодом из Telegram.'
    return err.message || 'Не удалось отправить'
  }
  return err instanceof Error ? err.message : 'Не удалось отправить'
}

export function DesktopChat() {
  const [history, setHistory] = useState<AssistantChatTurn[]>(() => loadStored().history)
  const [bubbles, setBubbles] = useState<Bubble[]>(() => loadStored().bubbles)
  const [draft, setDraft] = useState('')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')
  const listRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = listRef.current
    if (el) el.scrollTop = el.scrollHeight
  }, [bubbles, pending, error])

  async function send(e: FormEvent) {
    e.preventDefault()
    const text = draft.trim()
    if (!text || pending) return
    setDraft('')
    setError('')
    setPending(true)
    const optimistic: Bubble = { id: `local-${Date.now()}`, role: 'user', text }
    setBubbles((prev) => [...prev, optimistic])
    try {
      const sent = history.length > HISTORY_LIMIT ? history.slice(history.length - HISTORY_LIMIT) : history
      const res = await assistantChat(text, sent)
      const nextHistory = Array.isArray(res.history) ? res.history : sent
      const next = bubblesFromHistory(nextHistory)
      const tools = res.tools_run ?? []
      if (tools.length > 0 && next.length > 0 && next[next.length - 1].role === 'assistant') {
        const last = next[next.length - 1]
        next[next.length - 1] = { ...last, tools }
      }
      setHistory(nextHistory)
      setBubbles(next)
      remember(nextHistory, next)
    } catch (err) {
      setBubbles((prev) => prev.filter((b) => b.id !== optimistic.id))
      setDraft(text)
      setError(chatError(err))
    } finally {
      setPending(false)
    }
  }

  return (
    <aside className="flex h-full w-80 shrink-0 flex-col border-l border-white/10 bg-black/25">
      <div className="border-b border-white/10 px-4 py-4">
        <div className="text-sm font-semibold">Ассистент</div>
        <div className="mt-1 text-xs text-[#94a3b8]">тот же агент, что в Telegram</div>
      </div>
      <div ref={listRef} className="min-h-0 flex-1 space-y-3 overflow-y-auto px-3 py-3">
        {bubbles.length === 0 ? (
          <p className="px-1 text-sm leading-relaxed text-[#94a3b8]">
            Напиши задачу, расход или вопрос. Ответ идёт в твой аккаунт через{' '}
            <span className="text-[#cbd5e1]">/api/v1/assistant/chat</span>.
          </p>
        ) : null}
        {bubbles.map((bubble) => (
          <div
            key={bubble.id}
            className={
              bubble.role === 'user'
                ? 'ml-6 rounded-2xl bg-white/10 px-3 py-2 text-sm text-[#f8fafc]'
                : 'mr-4 rounded-2xl bg-[#14532d]/40 px-3 py-2 text-sm text-[#f8fafc]'
            }
          >
            <div className="whitespace-pre-wrap break-words">{bubble.text}</div>
            {bubble.tools && bubble.tools.length > 0 ? (
              <div className="mt-1 text-[11px] text-[#86efac]">Сделано: {bubble.tools.join(', ')}</div>
            ) : null}
          </div>
        ))}
        {pending ? <div className="px-1 text-xs text-[#94a3b8]">Думаю…</div> : null}
      </div>
      <form onSubmit={send} className="border-t border-white/10 p-3">
        {error ? <p className="mb-2 text-xs text-[#fca5a5]">{error}</p> : null}
        <textarea
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault()
              e.currentTarget.form?.requestSubmit()
            }
          }}
          rows={3}
          placeholder="Сообщение"
          aria-label="Сообщение ассистенту"
          className="w-full resize-none rounded-xl border border-white/10 bg-black/30 px-3 py-2 text-sm text-[#f8fafc] outline-none placeholder:text-[#64748b]"
        />
        <button
          type="submit"
          disabled={pending || draft.trim() === ''}
          className="mt-2 w-full rounded-xl bg-[var(--tg-theme-button-color,#22c55e)] px-3 py-2 text-sm font-semibold text-white disabled:opacity-50"
        >
          {pending ? 'Жду ответ…' : 'Отправить'}
        </button>
      </form>
    </aside>
  )
}
