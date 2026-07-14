import { cn } from '@/lib/cn'
import { periodLabel, type Period, recentPeriods, isSamePeriod } from '@/lib/periods'
import { hapticLight } from '@/lib/telegram'

type Props = {
  selected: Period
  onChange: (p: Period) => void
}

export function PeriodPicker({ selected, onChange }: Props) {
  const periods = recentPeriods(12)

  return (
    <div className="-mx-4 overflow-x-auto px-4 pb-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
      <div className="flex gap-2">
        {periods.map((p) => {
          const active = isSamePeriod(p, selected)
          return (
            <button
              key={`${p.year}-${p.month}`}
              type="button"
              onClick={() => {
                hapticLight()
                onChange(p)
              }}
              className={cn(
                'shrink-0 rounded-full px-4 py-2 text-sm font-medium transition',
                active
                  ? 'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]'
                  : 'bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-[var(--tg-theme-hint-color,#94a3b8)]',
              )}
            >
              {periodLabel(p, true)}
            </button>
          )
        })}
      </div>
    </div>
  )
}
