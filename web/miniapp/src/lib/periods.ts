export type Period = {
  year: number
  month: number
}

const MONTHS_RU = [
  'январь',
  'февраль',
  'март',
  'апрель',
  'май',
  'июнь',
  'июль',
  'август',
  'сентябрь',
  'октябрь',
  'ноябрь',
  'декабрь',
]

const MONTHS_SHORT = [
  'янв',
  'фев',
  'мар',
  'апр',
  'май',
  'июн',
  'июл',
  'авг',
  'сен',
  'окт',
  'ноя',
  'дек',
]

export function currentPeriod(): Period {
  const now = new Date()
  return { year: now.getFullYear(), month: now.getMonth() + 1 }
}

export function periodKey(p: Period): string {
  return `${p.year}-${String(p.month).padStart(2, '0')}`
}

export function parsePeriodKey(key: string): Period | null {
  const m = key.match(/^(\d{4})-(\d{2})$/)
  if (!m) return null
  const year = Number(m[1])
  const month = Number(m[2])
  if (month < 1 || month > 12) return null
  return { year, month }
}

export function periodLabel(p: Period, short = false): string {
  const names = short ? MONTHS_SHORT : MONTHS_RU
  const name = names[p.month - 1] ?? ''
  const now = currentPeriod()
  if (p.year === now.year && p.month === now.month) {
    return short ? 'Сейчас' : 'Текущий месяц'
  }
  if (p.year === now.year) {
    return name
  }
  return `${name} ${p.year}`
}

export function periodFullLabel(p: Period): string {
  const name = MONTHS_RU[p.month - 1] ?? ''
  return `${name} ${p.year}`
}

/** Last N months including current, newest first (T-Bank style scroll). */
export function recentPeriods(count = 12): Period[] {
  const out: Period[] = []
  let { year, month } = currentPeriod()
  for (let i = 0; i < count; i++) {
    out.push({ year, month })
    month -= 1
    if (month < 1) {
      month = 12
      year -= 1
    }
  }
  return out
}

export function isSamePeriod(a: Period, b: Period): boolean {
  return a.year === b.year && a.month === b.month
}
