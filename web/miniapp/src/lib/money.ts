export function formatMoney(cents: number, currency = 'RUB'): string {
  const amount = cents / 100
  const formatted = new Intl.NumberFormat('ru-RU', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(Math.abs(amount))

  const sign = cents < 0 ? '−' : cents > 0 ? '+' : ''
  const suffix = currency === 'RUB' ? '₽' : currency
  return `${sign}${formatted} ${suffix}`.trim()
}

export function formatMoneyPlain(cents: number, currency = 'RUB'): string {
  const amount = cents / 100
  const formatted = new Intl.NumberFormat('ru-RU', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
  const suffix = currency === 'RUB' ? '₽' : currency
  return `${formatted} ${suffix}`
}

export function parseMoneyInput(value: string): number | null {
  const digits = value.replace(/\s/g, '').replace(',', '.')
  const num = Number(digits)
  if (!Number.isFinite(num) || num <= 0) return null
  return Math.round(num * 100)
}
