import { useEffect, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { Button } from '@/components/ui/Button'
import { Sheet } from '@/components/ui/Sheet'
import { ruApiError } from '@/lib/apiError'
import { cn } from '@/lib/cn'
import { hapticError, hapticSuccess } from '@/lib/telegram'

type Props = {
  open: boolean
  onClose: () => void
  /** User-local calendar day (YYYY-MM-DD) the event is created for. */
  day: string
}

const TIME_PRESETS: Array<{ label: string; hour: number }> = [
  { label: 'Утро 09:00', hour: 9 },
  { label: 'День 13:00', hour: 13 },
  { label: 'Вечер 19:00', hour: 19 },
]

/**
 * MA-B5: bottom sheet to create a calendar event on a given day.
 * Combines the selected day with a local time and sends RFC3339 starts_at.
 */
export function CreateEventSheet({ open, onClose, day }: Props) {
  const queryClient = useQueryClient()
  const [title, setTitle] = useState('')
  const [time, setTime] = useState('12:00')
  const [formError, setFormError] = useState<string | null>(null)

  // Reset the form each time the sheet opens.
  useEffect(() => {
    if (open) {
      setTitle('')
      setTime('12:00')
      setFormError(null)
    }
  }, [open])

  const create = useMutation({
    mutationFn: () => {
      const startsAt = toRFC3339(day, time)
      return api.createCalendarEvent(title.trim(), startsAt)
    },
    onSuccess: () => {
      hapticSuccess()
      void queryClient.invalidateQueries({ queryKey: ['calendar'] })
      void queryClient.invalidateQueries({ queryKey: ['tasks', 'calendar'] })
      onClose()
    },
    onError: (err) => {
      hapticError()
      setFormError(ruApiError(err, 'Не удалось создать событие'))
    },
  })

  const canSubmit = title.trim().length > 0 && /^\d{2}:\d{2}$/.test(time) && !create.isPending

  return (
    <Sheet open={open} onClose={onClose} title="Новое событие">
      <input
        value={title}
        onChange={(e) => {
          setTitle(e.target.value)
          if (formError) setFormError(null)
        }}
        onKeyDown={(e) => {
          if (e.key === 'Enter' && canSubmit) create.mutate()
        }}
        placeholder="Название события"
        className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        autoFocus
      />

      <p className="mb-2 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">Время · {dayLabel(day)}</p>
      <div className="mb-3 flex flex-wrap gap-2">
        {TIME_PRESETS.map((p) => {
          const value = `${String(p.hour).padStart(2, '0')}:00`
          return (
            <button
              key={value}
              type="button"
              onClick={() => setTime(value)}
              className={cn(
                'rounded-full px-3 py-1.5 text-sm',
                time === value
                  ? 'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]'
                  : 'bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-[var(--tg-theme-hint-color,#94a3b8)]',
              )}
            >
              {p.label}
            </button>
          )
        })}
      </div>

      <label className="mb-3 block text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
        Или укажите время
        <input
          type="time"
          value={time}
          onChange={(e) => setTime(e.target.value)}
          className="mt-1 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
      </label>

      {formError && (
        <p className="mb-3 text-sm text-rose-400" role="alert">
          {formError}
        </p>
      )}

      <Button className="w-full" disabled={!canSubmit} onClick={() => create.mutate()}>
        {create.isPending ? 'Создание…' : 'Создать'}
      </Button>
    </Sheet>
  )
}

/** Combine a user-local day + HH:mm into an instant (browser-local timezone) as RFC3339. */
function toRFC3339(day: string, time: string): string {
  const d = new Date(`${day}T${time || '12:00'}:00`)
  return d.toISOString()
}

function dayLabel(iso: string): string {
  const today = toDateKey(new Date())
  if (iso === today) return 'сегодня'
  const d = new Date(`${iso}T12:00:00`)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleDateString('ru-RU', { weekday: 'short', day: 'numeric', month: 'short' })
}

function toDateKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}
