import type { TimeOfDay } from '@/api/types'

export function pad2(n: number) {
  return String(n).padStart(2, '0')
}

export function timeOfDayToInput(t: TimeOfDay | null | undefined): string {
  if (!t) return '00:00'
  return `${pad2(t.hour)}:${pad2(t.minute)}`
}

export function inputToTimeOfDay(value: string): TimeOfDay {
  const [h, m] = value.split(':').map(Number)
  return { hour: h || 0, minute: m || 0 }
}

export function formatTimeOfDay(t: TimeOfDay | null | undefined): string {
  if (!t) return '—'
  return `${pad2(t.hour)}:${pad2(t.minute)}`
}

export function formatShortDateTime(iso: string): string {
  try {
    return new Date(iso).toLocaleString('ru-RU', {
      day: 'numeric',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return iso
  }
}
