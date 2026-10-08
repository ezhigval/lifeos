import { useState, type ReactNode } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { api } from '@/api/client'
import type { CareerContact, CareerSkill } from '@/api/types'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { EmptyState } from '@/components/ui/EmptyState'
import { QueryError } from '@/components/ui/QueryError'
import { Sheet } from '@/components/ui/Sheet'
import { Skeleton } from '@/components/ui/Skeleton'
import { ruApiError } from '@/lib/apiError'
import { confirmAction, hapticError, hapticSuccess, hapticWarning } from '@/lib/telegram'

type SheetKind = 'contact' | 'skill' | null

const fieldClass =
  'w-full rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] px-4 py-3 text-base outline-none placeholder:text-[var(--tg-theme-hint-color,#94a3b8)]'

export function CareerPage() {
  const queryClient = useQueryClient()
  const [sheet, setSheet] = useState<SheetKind>(null)
  const [name, setName] = useState('')
  const [company, setCompany] = useState('')
  const [role, setRole] = useState('')
  const [notes, setNotes] = useState('')
  const [level, setLevel] = useState('')
  const [formError, setFormError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const contacts = useQuery({
    queryKey: ['career', 'contacts'],
    queryFn: async () => {
      const res = await api.careerContacts()
      return Array.isArray(res.contacts) ? res.contacts : []
    },
  })
  const skills = useQuery({
    queryKey: ['career', 'skills'],
    queryFn: async () => {
      const res = await api.careerSkills()
      return Array.isArray(res.skills) ? res.skills : []
    },
  })

  const create = useMutation({
    mutationFn: async () => {
      const trimmed = name.trim()
      if (!trimmed) throw new Error('empty name')
      if (sheet === 'contact') {
        return api.createCareerContact({
          name: trimmed,
          company: company.trim(),
          role: role.trim(),
          notes: notes.trim(),
        })
      }
      return api.createCareerSkill({ name: trimmed, level: level.trim() })
    },
    onSuccess: () => {
      hapticSuccess()
      void queryClient.invalidateQueries({ queryKey: ['career'] })
      closeSheet()
    },
    onError: (err) => {
      hapticError()
      setFormError(ruApiError(err, 'Не удалось сохранить'))
    },
  })

  const removeContact = useMutation({
    mutationFn: (id: string) => api.deleteCareerContact(id),
    onSuccess: () => {
      hapticSuccess()
      setActionError(null)
      void queryClient.invalidateQueries({ queryKey: ['career', 'contacts'] })
    },
    onError: (err) => {
      hapticError()
      setActionError(ruApiError(err, 'Не удалось удалить контакт'))
    },
  })

  const removeSkill = useMutation({
    mutationFn: (id: string) => api.deleteCareerSkill(id),
    onSuccess: () => {
      hapticSuccess()
      setActionError(null)
      void queryClient.invalidateQueries({ queryKey: ['career', 'skills'] })
    },
    onError: (err) => {
      hapticError()
      setActionError(ruApiError(err, 'Не удалось удалить навык'))
    },
  })

  function closeSheet() {
    setSheet(null)
    setName('')
    setCompany('')
    setRole('')
    setNotes('')
    setLevel('')
    setFormError(null)
  }

  function openSheet(kind: Exclude<SheetKind, null>) {
    setFormError(null)
    setName('')
    setCompany('')
    setRole('')
    setNotes('')
    setLevel('')
    setSheet(kind)
  }

  const contactItems = contacts.data ?? []
  const skillItems = skills.data ?? []
  const loading = contacts.isLoading || skills.isLoading

  return (
    <>
      <Header title="Карьера" subtitle="Контакты и навыки" />
      <div className="space-y-6 px-4 pb-6">
        {actionError && (
          <p className="text-sm text-rose-400" role="alert">
            {actionError}
          </p>
        )}

        {loading && (
          <div className="space-y-2">
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-20 w-full" />
          </div>
        )}

        {(contacts.isError || skills.isError) && (
          <QueryError
            message="Не удалось загрузить карьеру"
            onRetry={() => {
              void contacts.refetch()
              void skills.refetch()
            }}
          />
        )}

        {!loading && !contacts.isError && (
          <Section
            title="Контакты"
            emptyTitle="Нет контактов"
            emptyDescription="Имя, компания и роль. Бот пишет в тот же список."
            actionLabel="Контакт"
            items={contactItems}
            onAdd={() => openSheet('contact')}
            render={(item: CareerContact) => (
              <Row
                key={item.id}
                title={item.name}
                meta={[item.role, item.company].filter(Boolean).join(' · ')}
                note={item.notes}
                deleting={removeContact.isPending && removeContact.variables === item.id}
                onDelete={async () => {
                  hapticWarning()
                  if (await confirmAction('Удалить контакт?')) removeContact.mutate(item.id)
                }}
              />
            )}
          />
        )}

        {!loading && !skills.isError && (
          <Section
            title="Навыки"
            emptyTitle="Нет навыков"
            emptyDescription="Название и уровень. Тот же список, что у бота."
            actionLabel="Навык"
            items={skillItems}
            onAdd={() => openSheet('skill')}
            render={(item: CareerSkill) => (
              <Row
                key={item.id}
                title={item.name}
                meta={item.level}
                deleting={removeSkill.isPending && removeSkill.variables === item.id}
                onDelete={async () => {
                  hapticWarning()
                  if (await confirmAction('Удалить навык?')) removeSkill.mutate(item.id)
                }}
              />
            )}
          />
        )}
      </div>

      <Sheet
        open={sheet !== null}
        onClose={closeSheet}
        title={sheet === 'skill' ? 'Новый навык' : 'Новый контакт'}
      >
        <div className="space-y-3">
          <input
            className={fieldClass}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={sheet === 'skill' ? 'Go' : 'Имя'}
            autoFocus
          />
          {sheet === 'contact' ? (
            <>
              <input
                className={fieldClass}
                value={role}
                onChange={(e) => setRole(e.target.value)}
                placeholder="Роль"
              />
              <input
                className={fieldClass}
                value={company}
                onChange={(e) => setCompany(e.target.value)}
                placeholder="Компания"
              />
              <input
                className={fieldClass}
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                placeholder="Заметка"
              />
            </>
          ) : (
            <input
              className={fieldClass}
              value={level}
              onChange={(e) => setLevel(e.target.value)}
              placeholder="Уровень, например рабочий"
            />
          )}
          {formError && (
            <p className="text-sm text-rose-400" role="alert">
              {formError}
            </p>
          )}
          <Button className="w-full" disabled={!name.trim() || create.isPending} onClick={() => create.mutate()}>
            {create.isPending ? 'Сохраняем…' : 'Добавить'}
          </Button>
        </div>
      </Sheet>
    </>
  )
}

function Section<T extends { id: string }>({
  title,
  emptyTitle,
  emptyDescription,
  actionLabel,
  items,
  onAdd,
  render,
}: {
  title: string
  emptyTitle: string
  emptyDescription: string
  actionLabel: string
  items: T[]
  onAdd: () => void
  render: (item: T) => ReactNode
}) {
  return (
    <section className="space-y-2">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-base font-semibold">{title}</h2>
        {items.length > 0 && (
          <Button size="sm" onClick={onAdd}>
            <Plus size={16} className="mr-1" />
            {actionLabel}
          </Button>
        )}
      </div>
      {items.length === 0 ? (
        <EmptyState title={emptyTitle} description={emptyDescription} actionLabel="Добавить" onAction={onAdd} />
      ) : (
        <div className="space-y-2">{items.map(render)}</div>
      )}
    </section>
  )
}

function Row({
  title,
  meta,
  note,
  deleting,
  onDelete,
}: {
  title: string
  meta?: string
  note?: string
  deleting: boolean
  onDelete: () => void
}) {
  return (
    <div className="flex items-start gap-2 rounded-2xl bg-[var(--tg-theme-secondary-bg-color,#1e293b)] p-4">
      <div className="min-w-0 flex-1">
        <p className="font-medium">{title}</p>
        {meta ? (
          <p className="mt-1 text-xs text-[var(--tg-theme-hint-color,#94a3b8)]">{meta}</p>
        ) : null}
        {note ? <p className="mt-1 text-sm">{note}</p> : null}
      </div>
      <button
        type="button"
        className="rounded-full p-2 text-rose-400 disabled:opacity-50"
        aria-label="Удалить"
        disabled={deleting}
        onClick={onDelete}
      >
        <Trash2 size={16} />
      </button>
    </div>
  )
}
