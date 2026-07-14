import type { FinanceCategory } from '@/api/types'
import { colorForCategory } from '@/lib/categories'
import { formatMoneyPlain } from '@/lib/money'

const SIZE = 200
const CX = SIZE / 2
const CY = SIZE / 2
const R_OUTER = 88
const R_INNER = 62
const STROKE_OUTER = 22
const STROKE_BALANCE = 14

function polar(cx: number, cy: number, r: number, angleDeg: number) {
  const rad = ((angleDeg - 90) * Math.PI) / 180
  return { x: cx + r * Math.cos(rad), y: cy + r * Math.sin(rad) }
}

function arcPath(
  cx: number,
  cy: number,
  r: number,
  startAngle: number,
  endAngle: number,
) {
  const start = polar(cx, cy, r, endAngle)
  const end = polar(cx, cy, r, startAngle)
  const large = endAngle - startAngle > 180 ? 1 : 0
  return `M ${start.x} ${start.y} A ${r} ${r} 0 ${large} 0 ${end.x} ${end.y}`
}

type Props = {
  netCents: number
  expenseCents: number
  categories: FinanceCategory[]
  currency: string
}

export function FinanceRing({ netCents, expenseCents, categories, currency }: Props) {
  const positive = netCents >= 0
  const balanceColor = positive ? '#22c55e' : '#ef4444'

  let angle = 0
  const segments = categories.filter((c) => c.amount_cents > 0)
  const totalSeg = segments.reduce((s, c) => s + c.amount_cents, 0) || expenseCents || 1

  const balanceRatio =
    expenseCents > 0
      ? Math.min(Math.abs(netCents) / expenseCents, 1)
      : netCents !== 0
        ? 1
        : 0
  const balanceSweep = balanceRatio * 360

  return (
    <div className="relative mx-auto flex items-center justify-center" style={{ width: SIZE, height: SIZE }}>
      <svg width={SIZE} height={SIZE} viewBox={`0 0 ${SIZE} ${SIZE}`} className="overflow-visible">
        {/* Track */}
        <circle
          cx={CX}
          cy={CY}
          r={R_OUTER}
          fill="none"
          stroke="var(--tg-theme-secondary-bg-color,#1e293b)"
          strokeWidth={STROKE_OUTER}
          opacity={0.6}
        />

        {/* Category arcs (expenses) — T-Bank style outer ring */}
        {segments.map((cat) => {
          const sweep = (cat.amount_cents / totalSeg) * 360
          if (sweep < 0.5) return null
          const start = angle
          angle += sweep
          const color = cat.color_hint || colorForCategory(cat.name)
          return (
            <path
              key={cat.name}
              d={arcPath(CX, CY, R_OUTER, start, start + sweep - 1.5)}
              fill="none"
              stroke={color}
              strokeWidth={STROKE_OUTER}
              strokeLinecap="butt"
            />
          )
        })}

        {/* Balance inner ring */}
        <circle
          cx={CX}
          cy={CY}
          r={R_INNER}
          fill="none"
          stroke="var(--tg-theme-secondary-bg-color,#1e293b)"
          strokeWidth={STROKE_BALANCE}
          opacity={0.5}
        />
        {balanceSweep > 0 && (
          <path
            d={arcPath(CX, CY, R_INNER, 0, balanceSweep)}
            fill="none"
            stroke={balanceColor}
            strokeWidth={STROKE_BALANCE}
            strokeLinecap="round"
          />
        )}
      </svg>

      <div className="absolute inset-0 flex flex-col items-center justify-center text-center">
        <span className="text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">Баланс</span>
        <span
          className="tabular-nums text-2xl font-bold tracking-tight"
          style={{ color: balanceColor }}
        >
          {formatMoneyPlain(netCents, currency)}
        </span>
      </div>
    </div>
  )
}
