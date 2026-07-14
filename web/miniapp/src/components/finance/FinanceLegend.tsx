import type { FinanceCategory } from '@/api/types'
import { colorForCategory } from '@/lib/categories'
import { formatMoneyPlain } from '@/lib/money'

type Props = {
  categories: FinanceCategory[]
  currency: string
  /** When set, empty copy distinguishes zero spend vs missing breakdown. */
  expenseCents?: number
}

export function FinanceLegend({ categories, currency, expenseCents = 0 }: Props) {
  if (categories.length === 0) {
    return (
      <p className="text-center text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
        {expenseCents > 0
          ? 'Разбивка по категориям пока недоступна'
          : 'Нет расходов за период'}
      </p>
    )
  }

  return (
    <ul className="space-y-3">
      {categories.map((cat) => {
        const color = cat.color_hint || colorForCategory(cat.name)
        const pct = Math.round(cat.percent)
        return (
          <li key={cat.name}>
            <div className="mb-1 flex items-center justify-between text-sm">
              <span className="flex items-center gap-2">
                <span
                  className="h-2.5 w-2.5 shrink-0 rounded-full"
                  style={{ backgroundColor: color }}
                />
                {cat.name}
              </span>
              <span className="tabular-nums text-[var(--tg-theme-hint-color,#94a3b8)]">
                {formatMoneyPlain(cat.amount_cents, currency)}
              </span>
            </div>
            <div className="h-1.5 overflow-hidden rounded-full bg-[var(--tg-theme-secondary-bg-color,#1e293b)]">
              <div
                className="h-full rounded-full transition-all duration-300"
                style={{ width: `${pct}%`, backgroundColor: color }}
              />
            </div>
          </li>
        )
      })}
    </ul>
  )
}
