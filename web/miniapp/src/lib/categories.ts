/** Stable palette for major expense categories (T-Bank–inspired). */
const PALETTE = [
  '#FF6B6B',
  '#4ECDC4',
  '#45B7D1',
  '#96CEB4',
  '#FFEAA7',
  '#DDA0DD',
  '#98D8C8',
  '#F7DC6F',
]

const MAJOR_MIN_PERCENT = 5

export type CategorySlice = {
  name: string
  amountCents: number
  percent: number
  color: string
}

export function colorForCategory(name: string): string {
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return PALETTE[Math.abs(hash) % PALETTE.length]!
}

/** Keep only major categories; merge rest into «Прочее». */
export function majorCategories(
  raw: { name: string; amount_cents: number; percent?: number }[],
  totalExpense: number,
): CategorySlice[] {
  if (totalExpense <= 0 || raw.length === 0) return []

  const withPercent = raw.map((c) => ({
    name: c.name,
    amountCents: c.amount_cents,
    percent: c.percent ?? (c.amount_cents / totalExpense) * 100,
  }))

  const major = withPercent.filter((c) => c.percent >= MAJOR_MIN_PERCENT)
  const minor = withPercent.filter((c) => c.percent < MAJOR_MIN_PERCENT)

  const result: CategorySlice[] = major.map((c) => ({
    ...c,
    color: colorForCategory(c.name),
  }))

  if (minor.length > 0) {
    const otherAmount = minor.reduce((s, c) => s + c.amountCents, 0)
    result.push({
      name: 'Прочее',
      amountCents: otherAmount,
      percent: (otherAmount / totalExpense) * 100,
      color: '#94a3b8',
    })
  }

  return result.sort((a, b) => b.amountCents - a.amountCents)
}
