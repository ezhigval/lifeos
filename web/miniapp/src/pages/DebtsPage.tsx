import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import { api } from '@/api/client'
import type { Debt } from '@/api/types'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
import { formatMoneyPlain, parseMoneyInput } from '@/lib/money'
import { cn } from '@/lib/cn'
import { hapticError, hapticSuccess } from '@/lib/telegram'

type InstallmentInterval = 'none' | 'weekly' | 'monthly'

export function DebtsPage() {
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [payId, setPayId] = useState<string | null>(null)
  const [creditor, setCreditor] = useState('')
  const [amount, setAmount] = useState('')
  const [due, setDue] = useState('')
  const [installmentAmount, setInstallmentAmount] = useState('')
  const [installmentInterval, setInstallmentInterval] = useState<InstallmentInterval>('none')
  const [nextPaymentDate, setNextPaymentDate] = useState('')
  const [payAmount, setPayAmount] = useState('')
  const [formError, setFormError] = useState<string | null>(null)
  const [payError, setPayError] = useState<string | null>(null)

  const invalidateFinance = () => {
    void queryClient.invalidateQueries({ queryKey: ['debts'] })
    void queryClient.invalidateQueries({ queryKey: ['finance'] })
    void queryClient.invalidateQueries({ queryKey: ['finance-plan'] })
  }

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['debts'],
    queryFn: async () => {
      const res = await api.debts()
      return Array.isArray(res.debts) ? res.debts : []
    },
  })

  const debts = data ?? []
  const payingDebt = payId ? debts.find((d) => d.id === payId) : undefined

  const create = useMutation({
    mutationFn: () => {
      const cents = parseMoneyInput(amount)
      if (!cents) throw new Error('amount_cents is required')
      const instCents = parseMoneyInput(installmentAmount)
      return api.createDebt(creditor.trim(), cents, due || undefined, {
        ...(instCents ? { installment_cents: instCents } : {}),
        installment_interval: installmentInterval,
        ...(nextPaymentDate ? { next_payment_date: nextPaymentDate } : {}),
      })
    },
    onSuccess: (debt) => {
      hapticSuccess()
      queryClient.setQueryData<Debt[]>(['debts'], (old) => [debt, ...(old ?? [])])
      invalidateFinance()
      setCreateOpen(false)
      setCreditor('')
      setAmount('')
      setDue('')
      setInstallmentAmount('')
      setInstallmentInterval('none')
      setNextPaymentDate('')
      setFormError(null)
    },
    onError: (err) => {
      hapticError()
      setFormError(ruApiError(err, 'Не удалось создать долг'))
    },
  })

  const pay = useMutation({
    mutationFn: ({ cents, regular }: { cents: number; regular: boolean }) => {
      if (!payId) throw new Error('debt id required')
      const remaining = payingDebt?.remaining_cents
      if (remaining != null && cents > remaining) {
        throw new Error('payment exceeds remaining debt')
      }
      return api.payDebt(payId, cents, regular)
    },
    onMutate: async ({ cents }) => {
      if (!payId) return { prev: undefined }
      await queryClient.cancelQueries({ queryKey: ['debts'] })
      const prev = queryClient.getQueryData<Debt[]>(['debts'])
      queryClient.setQueryData<Debt[]>(['debts'], (old) =>
        (old ?? []).map((d) => {
          if (d.id !== payId) return d
          const paid = d.paid_cents + cents
          const remaining = Math.max(0, d.remaining_cents - cents)
          return { ...d, paid_cents: paid, remaining_cents: remaining }
        }),
      )
      return { prev }
    },
    onSuccess: () => {
      hapticSuccess()
      invalidateFinance()
      setPayId(null)
      setPayAmount('')
      setPayError(null)
    },
    onError: (err, _vars, ctx) => {
      hapticError()
      if (ctx?.prev) queryClient.setQueryData(['debts'], ctx.prev)
      setPayError(ruApiError(err, 'Не удалось записать платёж'))
    },
  })

  const openCreate = () => {
    setFormError(null)
    setCreditor('')
    setAmount('')
    setDue('')
    setInstallmentAmount('')
    setInstallmentInterval('none')
    setNextPaymentDate('')
    setCreateOpen(true)
  }

  const openPay = (id: string) => {
    const debt = debts.find((d) => d.id === id)
    setPayError(null)
    setPayAmount(
      debt?.installment_cents
        ? String(debt.installment_cents / 100)
        : '',
    )
    setPayId(id)
  }

  const openCount = debts.filter((d) => d.remaining_cents > 0).length

  return (
    <>
      <Header
        title="Долги"
        subtitle={debts.length ? `Открыто ${openCount}` : 'Кредиторы и выплаты'}
      />
      <div className="space-y-4 px-4 pb-6">
        {debts.length > 0 && (
          <div className="flex justify-end">
            <Button size="sm" onClick={openCreate}>
              <Plus size={16} className="mr-1" />
              Долг
            </Button>
          </div>
        )}

        {isLoading && (
          <div className="space-y-2">
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-20 w-full" />
          </div>
        )}

        {isError && (
          <QueryError message="Не удалось загрузить долги" onRetry={() => void refetch()} />
        )}

        {!isLoading && !isError && debts.length === 0 && (
          <EmptyState
            title="Долгов нет"
            description="Когда появится — зафиксируй сумму и срок"
            actionLabel="Добавить"
            onAction={openCreate}
          />
        )}

        {!isLoading && !isError && debts.length > 0 && (
          <div className="space-y-2">
            {debts.map((d) => (
              <div
                key={d.id}
                className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="font-medium">{d.creditor}</p>
                    <p className="mt-1 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
                      Остаток {formatMoneyPlain(d.remaining_cents, d.currency)}
                      {d.due_date ? ` · до ${formatDue(d.due_date)}` : ''}
                    </p>
                    {d.installment_cents && d.installment_interval && d.installment_interval !== 'none' && (
                      <p className="mt-1 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
                        Платёж {formatMoneyPlain(d.installment_cents, d.currency)} ·{' '}
                        {intervalLabel(d.installment_interval)}
                        {d.next_payment_date ? ` · след. ${formatDue(d.next_payment_date)}` : ''}
                      </p>
                    )}
                    {d.remaining_cents <= 0 && (
                      <p className="mt-1 text-xs text-emerald-400">Закрыт</p>
                    )}
                  </div>
                  {d.remaining_cents > 0 && (
                    <Button size="sm" variant="secondary" onClick={() => openPay(d.id)}>
                      Платеж
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <Sheet
        open={createOpen}
        onClose={() => {
          setCreateOpen(false)
          setFormError(null)
        }}
        title="Новый долг"
      >
        <input
          value={creditor}
          onChange={(e) => {
            setCreditor(e.target.value)
            if (formError) setFormError(null)
          }}
          placeholder="Кому"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          autoFocus
        />
        <input
          value={amount}
          onChange={(e) => {
            setAmount(e.target.value)
            if (formError) setFormError(null)
          }}
          inputMode="decimal"
          placeholder="Сумма, ₽"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <label className="mb-1 block text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
          Срок (опционально)
        </label>
        <input
          type="date"
          value={due}
          onChange={(e) => setDue(e.target.value)}
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <input
          value={installmentAmount}
          onChange={(e) => setInstallmentAmount(e.target.value)}
          inputMode="decimal"
          placeholder="Сумма платежа, ₽"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <p className="mb-2 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">Интервал платежа</p>
        <div className="mb-3 flex flex-wrap gap-2">
          {(['none', 'weekly', 'monthly'] as const).map((iv) => (
            <button
              key={iv}
              type="button"
              onClick={() => setInstallmentInterval(iv)}
              className={cn(
                'rounded-full px-3 py-1.5 text-sm',
                installmentInterval === iv
                  ? 'bg-[var(--tg-theme-button-color,#22c55e)] text-[var(--tg-theme-button-text-color,#fff)]'
                  : 'bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-[var(--tg-theme-hint-color,#94a3b8)]',
              )}
            >
              {intervalLabel(iv)}
            </button>
          ))}
        </div>
        <label className="mb-1 block text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">
          Следующий платёж (опционально)
        </label>
        <input
          type="date"
          value={nextPaymentDate}
          onChange={(e) => setNextPaymentDate(e.target.value)}
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        {formError && (
          <p className="mb-3 text-sm text-rose-400" role="alert">
            {formError}
          </p>
        )}
        <Button
          className="w-full"
          disabled={!creditor.trim() || !parseMoneyInput(amount) || create.isPending}
          onClick={() => create.mutate()}
        >
          Создать
        </Button>
      </Sheet>

      <Sheet
        open={Boolean(payId)}
        onClose={() => {
          setPayId(null)
          setPayError(null)
        }}
        title={payingDebt ? `Платёж · ${payingDebt.creditor}` : 'Платёж по долгу'}
      >
        {payingDebt && (
          <p className="mb-3 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
            Остаток {formatMoneyPlain(payingDebt.remaining_cents, payingDebt.currency)}
          </p>
        )}
        <input
          value={payAmount}
          onChange={(e) => {
            setPayAmount(e.target.value)
            if (payError) setPayError(null)
          }}
          inputMode="decimal"
          placeholder="Сумма, ₽"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          autoFocus
        />
        {payError && (
          <p className="mb-3 text-sm text-rose-400" role="alert">
            {payError}
          </p>
        )}
        <div className="space-y-2">
          <Button
            className="w-full"
            disabled={!parseMoneyInput(payAmount) || pay.isPending}
            onClick={() => {
              const cents = parseMoneyInput(payAmount)
              if (cents) pay.mutate({ cents, regular: true })
            }}
          >
            Регулярный платёж
          </Button>
          <Button
            className="w-full"
            variant="secondary"
            disabled={!parseMoneyInput(payAmount) || pay.isPending}
            onClick={() => {
              const cents = parseMoneyInput(payAmount)
              if (cents) pay.mutate({ cents, regular: false })
            }}
          >
            Досрочный
          </Button>
        </div>
      </Sheet>
    </>
  )
}

function formatDue(isoDate: string): string {
  const d = new Date(`${isoDate}T00:00:00`)
  if (Number.isNaN(d.getTime())) return isoDate
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}

function intervalLabel(iv: InstallmentInterval): string {
  switch (iv) {
    case 'weekly':
      return 'еженедельно'
    case 'monthly':
      return 'ежемесячно'
    default:
      return 'без графика'
  }
}
