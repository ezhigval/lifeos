import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { ArrowDownLeft, ArrowUpRight, Plus, Trash2 } from 'lucide-react'
import { api } from '@/api/client'
import type { FinanceOverview, FinancePlanItem } from '@/api/types'
import { FinanceRing } from '@/components/finance/FinanceRing'
import { FinanceLegend } from '@/components/finance/FinanceLegend'
import { PeriodPicker } from '@/components/finance/PeriodPicker'
import { TransactionSheet } from '@/components/finance/TransactionSheet'
import { Button } from '@/components/ui/Button'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
import { formatMoneyPlain, parseMoneyInput } from '@/lib/money'
import { cn } from '@/lib/cn'
import {
  currentPeriod,
  parsePeriodKey,
  periodKey,
  type Period,
} from '@/lib/periods'
import { getSavedPeriod, savePeriod } from '@/lib/storage'
import { confirmAction, hapticError, hapticSuccess, hapticWarning } from '@/lib/telegram'

type Props = {
  /** Prefer overview from GET /finance/overview (enriched); cash-flow fallback OK until then. */
  overview: FinanceOverview | undefined
  isLoading: boolean
  period: Period
  onPeriodChange: (p: Period) => void
}

export function FinanceCard({ overview, isLoading, period, onPeriodChange }: Props) {
  const [sheet, setSheet] = useState<'income' | 'expense' | null>(null)
  const [planSheet, setPlanSheet] = useState<'income' | 'expense' | null>(null)
  const [planTitle, setPlanTitle] = useState('')
  const [planAmount, setPlanAmount] = useState('')
  const [planInterval, setPlanInterval] = useState('monthly')
  const [planNextDate, setPlanNextDate] = useState(() => toDateKey(new Date()))
  const [planError, setPlanError] = useState<string | null>(null)
  const [txError, setTxError] = useState<string | null>(null)
  const queryClient = useQueryClient()

  const { data: plan, isLoading: planLoading, isError: planErrorQ, refetch: refetchPlan } = useQuery({
    queryKey: ['finance-plan'],
    queryFn: () => api.financePlan(),
  })

  const mutation = useMutation({
    mutationFn: async (data: { type: 'income' | 'expense'; amount: number; extra?: string }) => {
      if (data.type === 'income') {
        return api.recordIncome(data.amount, data.extra)
      }
      return api.recordExpense(data.amount, data.extra || 'Прочее')
    },
    onSuccess: () => {
      hapticSuccess()
      setTxError(null)
      queryClient.invalidateQueries({ queryKey: ['finance'] })
      setSheet(null)
    },
    onError: (err) => {
      hapticError()
      setTxError(ruApiError(err, 'Не удалось записать операцию'))
    },
  })

  const openSheet = (type: 'income' | 'expense') => {
    setTxError(null)
    setSheet(type)
  }

  const openPlanSheet = (type: 'income' | 'expense') => {
    setPlanError(null)
    setPlanTitle('')
    setPlanAmount('')
    setPlanInterval('monthly')
    setPlanNextDate(toDateKey(new Date()))
    setPlanSheet(type)
  }

  const createPlan = useMutation({
    mutationFn: () => {
      const cents = parseMoneyInput(planAmount)
      if (!cents || !planSheet) throw new Error('amount required')
      return api.createFinancePlan({
        kind: planSheet,
        title: planTitle.trim(),
        amount_cents: cents,
        currency: plan?.currency || overview?.currency || 'RUB',
        interval: planInterval,
        next_date: planNextDate,
      })
    },
    onSuccess: () => {
      hapticSuccess()
      setPlanError(null)
      void queryClient.invalidateQueries({ queryKey: ['finance-plan'] })
      setPlanSheet(null)
    },
    onError: (err) => {
      hapticError()
      setPlanError(ruApiError(err, 'Не удалось добавить в план'))
    },
  })

  const deletePlan = useMutation({
    mutationFn: (id: string) => api.deleteFinancePlan(id),
    onSuccess: () => {
      hapticSuccess()
      void queryClient.invalidateQueries({ queryKey: ['finance-plan'] })
    },
    onError: (err) => {
      hapticError()
      setPlanError(ruApiError(err, 'Не удалось удалить'))
    },
  })

  const planItems = plan?.items ?? []

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

          <FinancePlanBlock
            plan={plan}
            isLoading={planLoading}
            isError={planErrorQ}
            onRetry={() => void refetchPlan()}
            items={planItems}
            onAddIncome={() => openPlanSheet('income')}
            onAddExpense={() => openPlanSheet('expense')}
            onDelete={async (item) => {
              if (item.source !== 'plan') return
              const ok = await confirmAction(`Удалить «${item.title}» из плана?`)
              if (!ok) {
                hapticWarning()
                return
              }
              deletePlan.mutate(item.id)
            }}
            deleting={deletePlan.isPending}
          />
        </>
      ) : null}

      {txError && sheet === null && (
        <p className="mt-3 text-sm text-rose-400" role="alert">
          {txError}
        </p>
      )}

      <div className="mt-4 grid grid-cols-2 gap-3">
        <Button variant="income" size="lg" className="w-full" onClick={() => openSheet('income')}>
          + Доход
        </Button>
        <Button variant="expense" size="lg" className="w-full" onClick={() => openSheet('expense')}>
          − Расход
        </Button>
      </div>

      <TransactionSheet
        type={sheet}
        open={sheet !== null}
        onClose={() => {
          setSheet(null)
        }}
        loading={mutation.isPending}
        error={txError}
        onSubmit={(amount, extra) => {
          if (!sheet) return
          setTxError(null)
          mutation.mutate({ type: sheet, amount, extra })
        }}
      />

      <Sheet
        open={planSheet !== null}
        onClose={() => {
          setPlanSheet(null)
          setPlanError(null)
        }}
        title={planSheet === 'income' ? 'Плановый доход' : 'Плановый расход'}
      >
        <input
          value={planTitle}
          onChange={(e) => {
            setPlanTitle(e.target.value)
            if (planError) setPlanError(null)
          }}
          placeholder={planSheet === 'income' ? 'Зарплата' : 'Аренда'}
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-bg-color,#0f172a)] px-4 py-3 outline-none"
        />
        <input
          value={planAmount}
          onChange={(e) => {
            setPlanAmount(e.target.value)
            if (planError) setPlanError(null)
          }}
          inputMode="decimal"
          placeholder="Сумма, ₽"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-bg-color,#0f172a)] px-4 py-3 outline-none"
        />
        <p className="mb-2 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">Интервал</p>
        <div className="mb-3 flex flex-wrap gap-2">
          {['weekly', 'monthly', 'once'].map((iv) => (
            <button
              key={iv}
              type="button"
              onClick={() => setPlanInterval(iv)}
              className={cn(
                'rounded-full px-3 py-1.5 text-sm',
                planInterval === iv
                  ? 'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]'
                  : 'bg-[var(--tg-theme-bg-color,#0f172a)] text-[var(--tg-theme-hint-color,#94a3b8)]',
              )}
            >
              {iv === 'weekly' ? 'еженед.' : iv === 'monthly' ? 'ежемес.' : 'разово'}
            </button>
          ))}
        </div>
        <label className="mb-3 block text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
          Следующая дата
          <input
            type="date"
            value={planNextDate}
            onChange={(e) => setPlanNextDate(e.target.value)}
            className="mt-1 w-full rounded-2xl bg-[var(--tg-theme-bg-color,#0f172a)] px-4 py-3 outline-none"
          />
        </label>
        {planError && (
          <p className="mb-3 text-sm text-rose-400" role="alert">
            {planError}
          </p>
        )}
        <Button
          className="w-full"
          disabled={!planTitle.trim() || !parseMoneyInput(planAmount) || createPlan.isPending}
          onClick={() => createPlan.mutate()}
        >
          Добавить
        </Button>
      </Sheet>
    </section>
  )
}

function FinancePlanBlock({
  plan,
  isLoading,
  isError,
  onRetry,
  items,
  onAddIncome,
  onAddExpense,
  onDelete,
  deleting,
}: {
  plan: import('@/api/types').FinancePlan | undefined
  isLoading: boolean
  isError: boolean
  onRetry: () => void
  items: FinancePlanItem[]
  onAddIncome: () => void
  onAddExpense: () => void
  onDelete: (item: FinancePlanItem) => void
  deleting: boolean
}) {
  const currency = plan?.currency || 'RUB'

  return (
    <div className="mt-4 rounded-2xl bg-[var(--tg-theme-bg-color,#0f172a)]/40 p-3">
      <div className="mb-2 flex items-center justify-between">
        <h3 className="text-sm font-medium">План</h3>
        <div className="flex gap-1">
          <button
            type="button"
            onClick={onAddIncome}
            className="rounded-full p-1.5 text-emerald-400"
            aria-label="Добавить доход"
          >
            <Plus size={16} />
          </button>
          <button
            type="button"
            onClick={onAddExpense}
            className="rounded-full p-1.5 text-rose-400"
            aria-label="Добавить расход"
          >
            <Plus size={16} />
          </button>
        </div>
      </div>

      {isLoading && <Skeleton className="h-10 w-full" />}

      {isError && (
        <QueryError message="Не удалось загрузить план" onRetry={onRetry} />
      )}

      {!isLoading && !isError && plan && (
        <>
          <div className="mb-2 grid grid-cols-2 gap-2 text-xs">
            <div>
              <span className="text-emerald-400">+ </span>
              {formatMoneyPlain(plan.planned_income, currency)}
            </div>
            <div>
              <span className="text-rose-400">− </span>
              {formatMoneyPlain(plan.planned_expense, currency)}
            </div>
          </div>

          {items.length > 0 ? (
            <ul className="space-y-1.5 text-sm">
              {items.map((item) => (
                <li key={item.id} className="flex items-center justify-between gap-2">
                  <span className="min-w-0 truncate text-[var(--tg-theme-hint-color,#94a3b8)]">
                    {item.title}
                    {item.source === 'debt_installment' && (
                      <span className="ml-1 text-[10px] uppercase tracking-wide text-[var(--tg-theme-hint-color,#64748b)]">
                        долг
                      </span>
                    )}
                  </span>
                  <span className="flex shrink-0 items-center gap-1 tabular-nums">
                    <span className={item.kind === 'income' ? 'text-emerald-400' : 'text-rose-400'}>
                      {item.kind === 'income' ? '+' : '−'}
                      {formatMoneyPlain(item.amount_cents, item.currency || currency)}
                    </span>
                    <span className="text-xs text-[var(--tg-theme-hint-color,#64748b)]">
                      {formatPlanDate(item.next_date)}
                    </span>
                    {item.source === 'plan' && (
                      <button
                        type="button"
                        disabled={deleting}
                        onClick={() => onDelete(item)}
                        className="rounded-full p-1 text-[var(--tg-theme-hint-color,#94a3b8)] disabled:opacity-50"
                        aria-label={`Удалить ${item.title}`}
                      >
                        <Trash2 size={14} />
                      </button>
                    )}
                  </span>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
              Добавь плановый доход или расход
            </p>
          )}

          <div className="mt-2 flex gap-2">
            <Button size="sm" variant="secondary" className="flex-1" onClick={onAddIncome}>
              + Доход
            </Button>
            <Button size="sm" variant="secondary" className="flex-1" onClick={onAddExpense}>
              − Расход
            </Button>
          </div>

          <Link
            to="/more/debts"
            className="mt-2 block text-center text-xs text-[var(--tg-theme-link-color,#22c55e)]"
          >
            Платежи по долгам →
          </Link>
        </>
      )}
    </div>
  )
}

function toDateKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function formatPlanDate(iso: string): string {
  const d = new Date(`${iso}T12:00:00`)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })
}

export function useFinancePeriod() {
  const saved = getSavedPeriod()
  const parsed = saved ? parsePeriodKey(saved) : null
  const [period, setPeriod] = useState<Period>(parsed ?? currentPeriod())
  return { period, setPeriod }
}
