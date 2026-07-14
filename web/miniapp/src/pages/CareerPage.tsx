import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { api } from '@/api/client'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { confirmAction, hapticError, hapticSuccess, hapticWarning } from '@/lib/telegram'

export function CareerPage() {
  const queryClient = useQueryClient()
  const [tab, setTab] = useState<'contacts' | 'skills'>('contacts')
  const [createOpen, setCreateOpen] = useState(false)
  const [name, setName] = useState('')
  const [company, setCompany] = useState('')
  const [role, setRole] = useState('')
  const [level, setLevel] = useState('')

  const contacts = useQuery({
    queryKey: ['career', 'contacts'],
    queryFn: async () => (await api.contacts()).contacts,
  })
  const skills = useQuery({
    queryKey: ['career', 'skills'],
    queryFn: async () => (await api.skills()).skills,
  })

  const createContact = useMutation({
    mutationFn: () => api.createContact({ name: name.trim(), company, role }),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['career', 'contacts'] })
      setCreateOpen(false)
      setName('')
      setCompany('')
      setRole('')
    },
    onError: () => hapticError(),
  })

  const createSkill = useMutation({
    mutationFn: () => api.createSkill(name.trim(), level.trim()),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['career', 'skills'] })
      setCreateOpen(false)
      setName('')
      setLevel('')
    },
    onError: () => hapticError(),
  })

  const deleteContact = useMutation({
    mutationFn: (id: string) => api.deleteContact(id),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['career', 'contacts'] })
    },
    onError: () => hapticError(),
  })

  const deleteSkill = useMutation({
    mutationFn: (id: string) => api.deleteSkill(id),
    onSuccess: () => {
      hapticSuccess()
      queryClient.invalidateQueries({ queryKey: ['career', 'skills'] })
    },
    onError: () => hapticError(),
  })

  const listLoading = tab === 'contacts' ? contacts.isLoading : skills.isLoading
  const listError = tab === 'contacts' ? contacts.isError : skills.isError

  return (
    <>
      <Header title="Карьера" subtitle="Контакты и навыки" />
      <div className="space-y-4 px-4 pb-6">
        <div className="grid grid-cols-2 gap-2">
          <Button
            size="sm"
            variant={tab === 'contacts' ? 'primary' : 'secondary'}
            onClick={() => setTab('contacts')}
          >
            Контакты
          </Button>
          <Button
            size="sm"
            variant={tab === 'skills' ? 'primary' : 'secondary'}
            onClick={() => setTab('skills')}
          >
            Навыки
          </Button>
        </div>

        <div className="flex justify-end">
          <Button
            size="sm"
            onClick={() => {
              setName('')
              setCompany('')
              setRole('')
              setLevel('')
              setCreateOpen(true)
            }}
          >
            <Plus size={16} className="mr-1" />
            {tab === 'contacts' ? 'Контакт' : 'Навык'}
          </Button>
        </div>

        {listLoading && <Skeleton className="h-20 w-full" />}
        {listError && (
          <QueryError
            message="Ошибка загрузки"
            onRetry={() => void (tab === 'contacts' ? contacts.refetch() : skills.refetch())}
          />
        )}

        {tab === 'contacts' && !contacts.isLoading && (contacts.data ?? []).length === 0 && (
          <EmptyState title="Контактов нет" />
        )}
        {tab === 'skills' && !skills.isLoading && (skills.data ?? []).length === 0 && (
          <EmptyState title="Навыков нет" />
        )}

        <div className="space-y-2">
          {tab === 'contacts' &&
            (contacts.data ?? []).map((c) => (
              <div
                key={c.id}
                className="flex items-start gap-2 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4"
              >
                <div className="min-w-0 flex-1">
                  <p className="font-medium">{c.name}</p>
                  <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">
                    {[c.role, c.company].filter(Boolean).join(' · ') || '—'}
                  </p>
                </div>
                <button
                  type="button"
                  className="rounded-full p-2 text-rose-400"
                  aria-label="Удалить"
                  onClick={async () => {
                    hapticWarning()
                    if (await confirmAction(`Удалить ${c.name}?`)) deleteContact.mutate(c.id)
                  }}
                >
                  <Trash2 size={16} />
                </button>
              </div>
            ))}

          {tab === 'skills' &&
            (skills.data ?? []).map((s) => (
              <div
                key={s.id}
                className="flex items-center gap-2 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4"
              >
                <div className="min-w-0 flex-1">
                  <p className="font-medium">{s.name}</p>
                  {s.level && (
                    <p className="text-sm text-[var(--tg-theme-hint-color,#94a3b8)]">{s.level}</p>
                  )}
                </div>
                <button
                  type="button"
                  className="rounded-full p-2 text-rose-400"
                  aria-label="Удалить"
                  onClick={async () => {
                    hapticWarning()
                    if (await confirmAction(`Удалить ${s.name}?`)) deleteSkill.mutate(s.id)
                  }}
                >
                  <Trash2 size={16} />
                </button>
              </div>
            ))}
        </div>
      </div>

      <Sheet
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        title={tab === 'contacts' ? 'Новый контакт' : 'Новый навык'}
      >
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Имя"
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
        />
        {tab === 'contacts' ? (
          <>
            <input
              value={role}
              onChange={(e) => setRole(e.target.value)}
              placeholder="Роль"
              className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
            />
            <input
              value={company}
              onChange={(e) => setCompany(e.target.value)}
              placeholder="Компания"
              className="mb-4 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
            />
            <Button
              className="w-full"
              disabled={!name.trim() || createContact.isPending}
              onClick={() => createContact.mutate()}
            >
              Создать
            </Button>
          </>
        ) : (
          <>
            <input
              value={level}
              onChange={(e) => setLevel(e.target.value)}
              placeholder="Уровень (опционально)"
              className="mb-4 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
            />
            <Button
              className="w-full"
              disabled={!name.trim() || createSkill.isPending}
              onClick={() => createSkill.mutate()}
            >
              Создать
            </Button>
          </>
        )}
      </Sheet>
    </>
  )
}
