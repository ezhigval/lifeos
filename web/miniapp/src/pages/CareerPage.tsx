import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { api } from '@/api/client'
import type { Contact, Skill } from '@/api/types'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
import { confirmAction, hapticError, hapticSuccess, hapticWarning } from '@/lib/telegram'

export function CareerPage() {
  const queryClient = useQueryClient()
  const [tab, setTab] = useState<'contacts' | 'skills'>('contacts')
  const [createOpen, setCreateOpen] = useState(false)
  const [name, setName] = useState('')
  const [company, setCompany] = useState('')
  const [role, setRole] = useState('')
  const [level, setLevel] = useState('')
  const [formError, setFormError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const contacts = useQuery({
    queryKey: ['career', 'contacts'],
    queryFn: async () => {
      const res = await api.contacts()
      return Array.isArray(res.contacts) ? res.contacts : []
    },
  })
  const skills = useQuery({
    queryKey: ['career', 'skills'],
    queryFn: async () => {
      const res = await api.skills()
      return Array.isArray(res.skills) ? res.skills : []
    },
  })

  const contactList = contacts.data ?? []
  const skillList = skills.data ?? []
  const list = tab === 'contacts' ? contactList : skillList
  const listLoading = tab === 'contacts' ? contacts.isLoading : skills.isLoading
  const listError = tab === 'contacts' ? contacts.isError : skills.isError

  const createContact = useMutation({
    mutationFn: () =>
      api.createContact({
        name: name.trim(),
        company: company.trim() || undefined,
        role: role.trim() || undefined,
      }),
    onSuccess: (c) => {
      hapticSuccess()
      queryClient.setQueryData<Contact[]>(['career', 'contacts'], (old) => [
        c,
        ...(old ?? []).filter((x) => x.id !== c.id),
      ])
      void queryClient.invalidateQueries({ queryKey: ['career', 'contacts'] })
      setCreateOpen(false)
      setName('')
      setCompany('')
      setRole('')
      setFormError(null)
    },
    onError: (err) => {
      hapticError()
      setFormError(ruApiError(err, 'Не удалось создать контакт'))
    },
  })

  const createSkill = useMutation({
    mutationFn: () => api.createSkill(name.trim(), level.trim()),
    onSuccess: (s) => {
      hapticSuccess()
      queryClient.setQueryData<Skill[]>(['career', 'skills'], (old) => [
        s,
        ...(old ?? []).filter((x) => x.id !== s.id),
      ])
      void queryClient.invalidateQueries({ queryKey: ['career', 'skills'] })
      setCreateOpen(false)
      setName('')
      setLevel('')
      setFormError(null)
    },
    onError: (err) => {
      hapticError()
      setFormError(ruApiError(err, 'Не удалось создать навык'))
    },
  })

  const deleteContact = useMutation({
    mutationFn: (id: string) => api.deleteContact(id),
    onMutate: async (id) => {
      setActionError(null)
      await queryClient.cancelQueries({ queryKey: ['career', 'contacts'] })
      const prev = queryClient.getQueryData<Contact[]>(['career', 'contacts'])
      queryClient.setQueryData<Contact[]>(['career', 'contacts'], (old) =>
        (old ?? []).filter((c) => c.id !== id),
      )
      return { prev }
    },
    onSuccess: () => hapticSuccess(),
    onError: (err, _id, ctx) => {
      hapticError()
      if (ctx?.prev) queryClient.setQueryData(['career', 'contacts'], ctx.prev)
      setActionError(ruApiError(err, 'Не удалось удалить'))
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: ['career', 'contacts'] })
    },
  })

  const deleteSkill = useMutation({
    mutationFn: (id: string) => api.deleteSkill(id),
    onMutate: async (id) => {
      setActionError(null)
      await queryClient.cancelQueries({ queryKey: ['career', 'skills'] })
      const prev = queryClient.getQueryData<Skill[]>(['career', 'skills'])
      queryClient.setQueryData<Skill[]>(['career', 'skills'], (old) =>
        (old ?? []).filter((s) => s.id !== id),
      )
      return { prev }
    },
    onSuccess: () => hapticSuccess(),
    onError: (err, _id, ctx) => {
      hapticError()
      if (ctx?.prev) queryClient.setQueryData(['career', 'skills'], ctx.prev)
      setActionError(ruApiError(err, 'Не удалось удалить'))
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: ['career', 'skills'] })
    },
  })

  const openCreate = () => {
    setFormError(null)
    setName('')
    setCompany('')
    setRole('')
    setLevel('')
    setCreateOpen(true)
  }

  const deletingContactId = deleteContact.isPending ? deleteContact.variables : null
  const deletingSkillId = deleteSkill.isPending ? deleteSkill.variables : null
  const creating = tab === 'contacts' ? createContact.isPending : createSkill.isPending

  return (
    <>
      <Header title="Карьера" subtitle="Контакты и навыки" />
      <div className="space-y-4 px-4 pb-6">
        <div className="grid grid-cols-2 gap-2">
          <Button
            size="sm"
            variant={tab === 'contacts' ? 'primary' : 'secondary'}
            onClick={() => {
              setTab('contacts')
              setActionError(null)
            }}
          >
            Контакты
          </Button>
          <Button
            size="sm"
            variant={tab === 'skills' ? 'primary' : 'secondary'}
            onClick={() => {
              setTab('skills')
              setActionError(null)
            }}
          >
            Навыки
          </Button>
        </div>

        {list.length > 0 && (
          <div className="flex justify-end">
            <Button size="sm" onClick={openCreate}>
              <Plus size={16} className="mr-1" />
              {tab === 'contacts' ? 'Контакт' : 'Навык'}
            </Button>
          </div>
        )}

        {actionError && (
          <p className="text-sm text-rose-400" role="alert">
            {actionError}
          </p>
        )}

        {listLoading && (
          <div className="space-y-2">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
        )}

        {listError && (
          <QueryError
            message="Ошибка загрузки"
            onRetry={() => void (tab === 'contacts' ? contacts.refetch() : skills.refetch())}
          />
        )}

        {!listLoading && !listError && tab === 'contacts' && contactList.length === 0 && (
          <EmptyState
            title="Контактов нет"
            description="Люди из сети — имя и роль"
            actionLabel="Добавить"
            onAction={openCreate}
          />
        )}
        {!listLoading && !listError && tab === 'skills' && skillList.length === 0 && (
          <EmptyState
            title="Навыков нет"
            description="Зафиксируй стек и уровень"
            actionLabel="Добавить"
            onAction={openCreate}
          />
        )}

        {!listLoading && !listError && tab === 'contacts' && contactList.length > 0 && (
          <div className="space-y-2">
            {contactList.map((c) => (
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
                  className="rounded-full p-2 text-rose-400 disabled:opacity-50"
                  aria-label="Удалить"
                  disabled={deletingContactId === c.id}
                  onClick={async () => {
                    hapticWarning()
                    if (await confirmAction(`Удалить ${c.name}?`)) deleteContact.mutate(c.id)
                  }}
                >
                  <Trash2 size={16} />
                </button>
              </div>
            ))}
          </div>
        )}

        {!listLoading && !listError && tab === 'skills' && skillList.length > 0 && (
          <div className="space-y-2">
            {skillList.map((s) => (
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
                  className="rounded-full p-2 text-rose-400 disabled:opacity-50"
                  aria-label="Удалить"
                  disabled={deletingSkillId === s.id}
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
        )}
      </div>

      <Sheet
        open={createOpen}
        onClose={() => {
          setCreateOpen(false)
          setFormError(null)
        }}
        title={tab === 'contacts' ? 'Новый контакт' : 'Новый навык'}
      >
        <input
          value={name}
          onChange={(e) => {
            setName(e.target.value)
            if (formError) setFormError(null)
          }}
          placeholder={tab === 'contacts' ? 'Имя' : 'Навык'}
          className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          autoFocus
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
              className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
            />
          </>
        ) : (
          <input
            value={level}
            onChange={(e) => setLevel(e.target.value)}
            placeholder="Уровень (опционально)"
            className="mb-3 w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 outline-none"
          />
        )}
        {formError && (
          <p className="mb-3 text-sm text-rose-400" role="alert">
            {formError}
          </p>
        )}
        <Button
          className="w-full"
          disabled={!name.trim() || creating}
          onClick={() => (tab === 'contacts' ? createContact.mutate() : createSkill.mutate())}
        >
          Создать
        </Button>
      </Sheet>
    </>
  )
}
