import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import { api } from '@/api/client'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatMoneyPlain, parseMoneyInput } from '@/lib/money'
import { hapticError, hapticSuccess } from '@/lib/telegram'

export function DebtsPage() {
  const queryClient = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [payId, setPayId] = useState<string | null>(null)
  const [creditor, setCreditor] = useState('')
  const [amount, setAmount] = useState('')
  const [due, setDue] = useState('')
  const [payAmount, setPayAmount] = useState('')

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['debts'],
    queryFn: async () => (await api.debts()).debts,
  })

  const create = useMutation({
    mutationFn: () => {
      const cents = parseMoneyInput(amount)
      if (!cents) throw new Error('amount')
      return api.createDebt(creditor.trim(), cents, due || undefined)
    },
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['debts'] })
      setCreateOpen(false)
      setCreditor('')
      setAmount('')
      setDue('')
    },
    onError: () => hapticError(),
  })

  const pay = useMutation({
    mutationFn: () => {
      const cents = parseMoneyInput(payAmount)
      if (!cents || !payId) throw new Error('amount')
      return api.payDebt(payId, cents)
    },
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['debts'] })
      setPayId(null)
      setPayAmount('')
    },
    onError: () => hapticError(),
  })

  return (
    <>
      <Header title="Долги" subtitle="Кредиторы и выплаты" />
      <div className="space-y-4 px-4 pb-6">
        <div className="flex justify-end">
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus size={16} className="mr-1" />
            Долг
          </Button>
        </div>

        {isLoading && <Skeleton className="h-20 w-full" />}
        {isError && <QueryError message="Не удалось загрузить долги" onRetry={() => void refetch()} />}
        {!isLoading && !isError && (data ?? []).length === 0 && (
          <EmptyState title="Долгов нет" description="Отлично" />
        )}

        <div className="space-y-2">
          {(data ?? []).map((d) => (
            <div
              key={d.id}
              className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4"
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="font-medium">{d.creditor}</p>
                  <p className="mt-1 text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
                    Остаток {formatMoneyPlain(d.remaining_cents, d.currency)}
                    {d.due_date ? ` · до ${d.due_date}` : ''}
                  </p>
                </div>
                {d.remaining_cents > 0 && (
                  <Button size="sm" variant="secondary" onClick={() => setPayId(d.id)}>
                    Платеж
                  </Button>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      <Sheet open={createOpen} onClose={() => setCreateOpen(false)} title="Новый долг">
        <input
          value={creditor}
          onChange={(e) => setCreditor(e.target.value)}
          placeholder="Кому"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <input
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          inputMode="decimal"
          placeholder="Сумма, ₽"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <input
          type="date"
          value={due}
          onChange={(e) => setDue(e.target.value)}
          className="mb-4 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <Button
          className="w-full"
          disabled={!creditor.trim() || !parseMoneyInput(amount) || create.isPending}
          onClick={() => create.mutate()}
        >
          Создать
        </Button>
      </Sheet>

      <Sheet open={Boolean(payId)} onClose={() => setPayId(null)} title="Платёж по долгу">
        <input
          value={payAmount}
          onChange={(e) => setPayAmount(e.target.value)}
          inputMode="decimal"
          placeholder="Сумма, ₽"
          className="mb-4 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        <Button
          className="w-full"
          disabled={!parseMoneyInput(payAmount) || pay.isPending}
          onClick={() => pay.mutate()}
        >
          Записать платёж
        </Button>
      </Sheet>
    </>
  )
}
