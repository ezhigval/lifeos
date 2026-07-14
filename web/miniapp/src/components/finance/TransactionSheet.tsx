import { useState } from 'react'
import { Sheet } from '@/components/ui/Sheet'
import { Button } from '@/components/ui/Button'
import { parseMoneyInput } from '@/lib/money'

const EXPENSE_CATEGORIES = [
  'Еда',
  'Транспорт',
  'Дом',
  'Здоровье',
  'Развлечения',
  'Прочее',
]

type Props = {
  type: 'income' | 'expense' | null
  open: boolean
  onClose: () => void
  onSubmit: (amountCents: number, extra?: string) => void
  loading?: boolean
}

export function TransactionSheet({ type, open, onClose, onSubmit, loading }: Props) {
  const [amount, setAmount] = useState('')
  const [description, setDescription] = useState('')
  const [category, setCategory] = useState('Прочее')

  const title = type === 'income' ? 'Добавить доход' : 'Добавить расход'

  function handleSubmit() {
    const cents = parseMoneyInput(amount)
    if (!cents) return
    if (type === 'income') {
      onSubmit(cents, description || undefined)
    } else {
      onSubmit(cents, category)
    }
    setAmount('')
    setDescription('')
    setCategory('Прочее')
  }

  const appendDigit = (d: string) => setAmount((a) => (a + d).replace(/^0+/, '') || d)

  return (
    <Sheet open={open} onClose={onClose} title={title}>
      <div className="space-y-4">
        <div className="rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-6 text-center">
          <input
            type="text"
            inputMode="decimal"
            value={amount}
            onChange={(e) => setAmount(e.target.value.replace(/[^\d.,]/g, ''))}
            placeholder="0"
            className="w-full bg-transparent text-center text-4xl font-bold tabular-nums outline-none"
          />
          <span className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">₽</span>
        </div>

        {type === 'expense' && (
          <div className="flex flex-wrap gap-2">
            {EXPENSE_CATEGORIES.map((c) => (
              <button
                key={c}
                type="button"
                onClick={() => setCategory(c)}
                className={
                  category === c
                    ? 'rounded-full bg-[var(--tg-theme-button-color,#22c55e)] px-3 py-1.5 text-sm text-white'
                    : 'rounded-full bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-3 py-1.5 text-sm'
                }
              >
                {c}
              </button>
            ))}
          </div>
        )}

        {type === 'income' && (
          <input
            type="text"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Описание (необязательно)"
            className="w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          />
        )}

        <div className="grid grid-cols-3 gap-2">
          {['1', '2', '3', '4', '5', '6', '7', '8', '9', ',', '0', '⌫'].map((key) => (
            <button
              key={key}
              type="button"
              onClick={() => {
                if (key === '⌫') setAmount((a) => a.slice(0, -1))
                else if (key === ',') setAmount((a) => (a.includes(',') ? a : a + ','))
                else appendDigit(key)
              }}
              className="h-12 rounded-xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] text-lg font-medium"
            >
              {key}
            </button>
          ))}
        </div>

        <Button
          className="w-full"
          size="lg"
          disabled={!parseMoneyInput(amount) || loading}
          onClick={handleSubmit}
        >
          {loading ? 'Сохраняю…' : 'Сохранить'}
        </Button>
      </div>
    </Sheet>
  )
}
