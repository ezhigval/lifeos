import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowDownLeft, ArrowUpRight } from 'lucide-react'
import { api } from '@/api/client'
import type { FinanceOverview } from '@/api/types'
import { FinanceRing } from '@/components/finance/FinanceRing'
import { FinanceLegend } from '@/components/finance/FinanceLegend'
import { PeriodPicker } from '@/components/finance/PeriodPicker'
import { TransactionSheet } from '@/components/finance/TransactionSheet'
import { Button } from '@/components/ui/Button'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatMoneyPlain } from '@/lib/money'
import {
  currentPeriod,
  parsePeriodKey,
  periodKey,
  type Period,
} from '@/lib/periods'
import { getSavedPeriod, savePeriod } from '@/lib/storage'
import { hapticSuccess } from '@/lib/telegram'

type Props = {
  /** Prefer overview from GET /finance/overview (enriched); cash-flow fallback OK until then. */
  overview: FinanceOverview | undefined
  isLoading: boolean
  period: Period
  onPeriodChange: (p: Period) => void
}

export function FinanceCard({ overview, isLoading, period, onPeriodChange }: Props) {
  const [sheet, setSheet] = useState<'income' | 'expense' | null>(null)
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: async (data: { type: 'income' | 'expense'; amount: number; extra?: string }) => {
      if (data.type === 'income') {
        return api.recordIncome(data.amount, data.extra)
      }
      return api.recordExpense(data.amount, data.extra || 'Прочее')
    },
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['finance'] })
      setSheet(null)
    },
  })

  return (
    <section className="rounded-3xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-base font-semibold">Финансы</h2>
        {overview && (
          <span className="text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
            {overview.period_label}
          </span>
        )}
      </div>

      <PeriodPicker
        selected={period}
        onChange={(p) => {
          savePeriod(periodKey(p))
          onPeriodChange(p)
        }}
      />

      {isLoading ? (
        <div className="mt-4 space-y-3">
          <Skeleton className="mx-auto h-[200px] w-[200px] rounded-full" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </div>
      ) : overview ? (
        <>
          <div className="mt-4">
            <FinanceRing
              netCents={overview.net_cents}
              expenseCents={overview.expense_cents}
              categories={overview.categories}
              currency={overview.currency}
            />
          </div>

          <div className="mt-4 grid grid-cols-2 gap-3 text-center text-sm">
            <div className="rounded-2xl bg-emerald-500/10 px-3 py-2">
              <div className="flex items-center justify-center gap-1 text-emerald-400">
                <ArrowDownLeft size={14} />
                Доход
              </div>
              <div className="tabular-nums font-semibold">
                {formatMoneyPlain(overview.income_cents, overview.currency)}
              </div>
            </div>
            <div className="rounded-2xl bg-rose-500/10 px-3 py-2">
              <div className="flex items-center justify-center gap-1 text-rose-400">
                <ArrowUpRight size={14} />
                Расход
              </div>
              <div className="tabular-nums font-semibold">
                {formatMoneyPlain(overview.expense_cents, overview.currency)}
              </div>
            </div>
          </div>

          <div className="mt-4">
            <FinanceLegend
              categories={overview.categories}
              currency={overview.currency}
              expenseCents={overview.expense_cents}
            />
          </div>
        </>
      ) : null}

      <div className="mt-4 grid grid-cols-2 gap-3">
        <Button variant="income" size="lg" className="w-full" onClick={() => setSheet('income')}>
          + Доход
        </Button>
        <Button variant="expense" size="lg" className="w-full" onClick={() => setSheet('expense')}>
          − Расход
        </Button>
      </div>

      <TransactionSheet
        type={sheet}
        open={sheet !== null}
        onClose={() => setSheet(null)}
        loading={mutation.isPending}
        onSubmit={(amount, extra) => {
          if (!sheet) return
          mutation.mutate({ type: sheet, amount, extra })
        }}
      />
    </section>
  )
}

export function useFinancePeriod() {
  const saved = getSavedPeriod()
  const parsed = saved ? parsePeriodKey(saved) : null
  const [period, setPeriod] = useState<Period>(parsed ?? currentPeriod())
  return { period, setPeriod }
}
