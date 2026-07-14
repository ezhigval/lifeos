import type { FinanceOverview } from '@/api/types'
import { periodKey, periodFullLabel, currentPeriod, isSamePeriod, type Period } from '@/lib/periods'
import { majorCategories } from '@/lib/categories'
import { clearSession } from '@/lib/session'

export type { Period } from '@/lib/periods'

export type AuthResult = {
  accessToken: string
  expiresIn: number
  /** Signed Telegram user id from server (initData.user.id) */
  telegramId?: number
}

let accessToken: string | null = null
let onUnauthorized: (() => Promise<boolean>) | null = null

export function setAccessToken(token: string | null) {
  accessToken = token
}

export function getAccessToken() {
  return accessToken
}

/** Called once when a request gets 401; return true if auth was refreshed. */
export function setUnauthorizedHandler(handler: (() => Promise<boolean>) | null) {
  onUnauthorized = handler
}

class ApiClientError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(
  path: string,
  init: RequestInit = {},
  allowRefresh = true,
): Promise<T> {
  const headers = new Headers(init.headers)
  if (!headers.has('Content-Type') && init.body) {
    headers.set('Content-Type', 'application/json')
  }
  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`)
  }

  const res = await fetch(path, { ...init, headers })
  if (res.status === 401 && allowRefresh && onUnauthorized) {
    const refreshed = await onUnauthorized()
    if (refreshed) {
      return request<T>(path, init, false)
    }
    clearSession()
    setAccessToken(null)
  }
  if (!res.ok) {
    let msg = res.statusText
    try {
      const body = (await res.json()) as { error?: string }
      if (body.error) msg = body.error
    } catch {
      /* ignore */
    }
    throw new ApiClientError(res.status, msg)
  }
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

export async function authWithInitData(initData: string): Promise<AuthResult> {
  try {
    const data = await request<{
      access_token: string
      expires_in?: number
      telegram_id?: number
    }>(
      '/api/v1/auth/telegram-webapp',
      {
        method: 'POST',
        body: JSON.stringify({ init_data: initData }),
      },
      false,
    )
    return {
      accessToken: data.access_token,
      expiresIn: data.expires_in ?? 0,
      telegramId:
        typeof data.telegram_id === 'number' && data.telegram_id > 0
          ? data.telegram_id
          : undefined,
    }
  } catch (e) {
    if (e instanceof ApiClientError && e.status === 404) {
      throw new Error('auth/telegram-webapp not implemented')
    }
    throw e
  }
}

export async function authWithDevCredentials(
  apiKey: string,
  telegramId: number,
): Promise<AuthResult> {
  const data = await request<{
    access_token: string
    expires_in?: number
    telegram_id?: number
  }>(
    '/api/v1/auth/token',
    {
      method: 'POST',
      headers: { 'X-API-Key': apiKey },
      body: JSON.stringify({ telegram_id: telegramId }),
    },
    false,
  )
  return {
    accessToken: data.access_token,
    expiresIn: data.expires_in ?? 0,
    telegramId:
      typeof data.telegram_id === 'number' && data.telegram_id > 0
        ? data.telegram_id
        : telegramId,
  }
}

export const api = {
  tasksToday: () => request<{ tasks: import('@/api/types').Task[] }>('/api/v1/tasks/today'),

  priorities: () =>
    request<{ priorities: import('@/api/types').PriorityItem[] }>('/api/v1/priorities'),

  completeTask: (id: string) =>
    request<{ id: string }>(`/api/v1/tasks/${id}/complete`, { method: 'POST' }),

  getTask: (id: string) => request<import('@/api/types').Task>(`/api/v1/tasks/${id}`),

  updateTask: (id: string, body: {
    title?: string
    priority?: string
    due_date?: string
    clear_due_date?: boolean
    description?: string
    clear_description?: boolean
    duration_minutes?: number
    clear_duration?: boolean
    tags?: string[]
    project_ids?: string[]
  }) =>
    request<import('@/api/types').Task>(`/api/v1/tasks/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),

  archiveTask: (id: string) =>
    request<import('@/api/types').Task>(`/api/v1/tasks/${id}/archive`, { method: 'POST' }),

  deleteTask: (id: string) =>
    request<void>(`/api/v1/tasks/${id}`, { method: 'DELETE' }),

  createTask: (body: {
    title: string
    priority?: string
    due_date?: string
    project_ids?: string[]
  }) =>
    request<import('@/api/types').Task>('/api/v1/tasks', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  spheres: () => request<{ spheres: import('@/api/types').Sphere[] }>('/api/v1/settings/spheres'),

  projects: (sphereId?: string) => {
    const q = sphereId ? `?sphere_id=${sphereId}` : ''
    return request<{ projects: import('@/api/types').Project[] }>(`/api/v1/projects${q}`)
  },

  projectTasks: (projectId: string) =>
    request<{ tasks: import('@/api/types').Task[] }>(`/api/v1/projects/${projectId}/tasks`),

  createProject: (body: {
    name: string
    sphere_ids: string[]
    outcome?: string
    target_value?: string
    unit?: string
    target_date?: string
  }) =>
    request<import('@/api/types').Project>('/api/v1/projects', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  archiveProject: (projectId: string) =>
    request(`/api/v1/projects/${projectId}/archive`, { method: 'POST' }),

  projectProgress: (projectId: string) =>
    request<{
      percent: string
      current: string
      target: string
      has_target: boolean
    }>(`/api/v1/projects/progress?project_id=${projectId}`),

  recordIncome: (amount_cents: number, description?: string) =>
    request('/api/v1/finance/income', {
      method: 'POST',
      body: JSON.stringify({ amount_cents, description: description || 'доход' }),
    }),

  recordExpense: (amount_cents: number, category: string) =>
    request('/api/v1/finance/expense', {
      method: 'POST',
      body: JSON.stringify({ amount_cents, category }),
    }),

  habitsToday: () =>
    request<{ habits: import('@/api/types').HabitDay[] }>('/api/v1/habits/today'),

  createHabit: (name: string) =>
    request<{ id: string; name: string; frequency: string }>('/api/v1/habits', {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),

  trackHabit: (id: string) =>
    request<{ name: string; streak: number }>(`/api/v1/habits/${id}/track`, {
      method: 'POST',
    }),

  calendarToday: () =>
    request<{ events: import('@/api/types').CalendarEvent[] }>('/api/v1/calendar/today'),

  createCalendarEvent: (title: string, starts_at: string) =>
    request<import('@/api/types').CalendarEvent>('/api/v1/calendar/events', {
      method: 'POST',
      body: JSON.stringify({ title, starts_at }),
    }),

  createSphere: (name: string) =>
    request<import('@/api/types').Sphere>('/api/v1/settings/spheres', {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),

  updateSphere: (id: string, name: string, sort_order: number) =>
    request<import('@/api/types').Sphere>(`/api/v1/settings/spheres/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ name, sort_order }),
    }),

  deleteSphere: (id: string) =>
    request<import('@/api/types').Sphere>(`/api/v1/settings/spheres/${id}`, {
      method: 'DELETE',
    }),

  settings: () => request<import('@/api/types').UserSettings>('/api/v1/settings'),

  updateMorningReview: (hour: number, minute: number) =>
    request('/api/v1/settings/morning-review', {
      method: 'PUT',
      body: JSON.stringify({ hour, minute }),
    }),

  updateEveningReview: (hour: number, minute: number) =>
    request('/api/v1/settings/evening-review', {
      method: 'PUT',
      body: JSON.stringify({ hour, minute }),
    }),

  updateQuietHours: (start_hour: number, start_minute: number, end_hour: number, end_minute: number) =>
    request('/api/v1/settings/quiet-hours', {
      method: 'PUT',
      body: JSON.stringify({ start_hour, start_minute, end_hour, end_minute }),
    }),

  analyticsSummary: () =>
    request<import('@/api/types').AnalyticsSummary>('/api/v1/analytics/summary'),

  notes: (q?: string) => {
    const qs = q ? `?q=${encodeURIComponent(q)}` : ''
    return request<{ notes: import('@/api/types').Note[] }>(`/api/v1/notes${qs}`)
  },

  createNote: (body: string, tags?: string[]) =>
    request<import('@/api/types').Note>('/api/v1/notes', {
      method: 'POST',
      body: JSON.stringify({ body, tags: tags ?? [] }),
    }),

  deleteNote: (id: string) =>
    request<import('@/api/types').Note>(`/api/v1/notes/${id}`, { method: 'DELETE' }),

  debts: () => request<{ debts: import('@/api/types').Debt[] }>('/api/v1/finance/debts'),

  createDebt: (creditor: string, amount_cents: number, due_date?: string) =>
    request<import('@/api/types').Debt>('/api/v1/finance/debts', {
      method: 'POST',
      body: JSON.stringify({ creditor, amount_cents, due_date }),
    }),

  payDebt: (id: string, amount_cents: number) =>
    request(`/api/v1/finance/debts/${id}/pay`, {
      method: 'POST',
      body: JSON.stringify({ amount_cents }),
    }),

  reminders: () =>
    request<{ reminders: import('@/api/types').Reminder[] }>('/api/v1/reminders'),

  createReminder: (message: string, fire_at: string) =>
    request('/api/v1/reminders', {
      method: 'POST',
      body: JSON.stringify({ message, fire_at }),
    }),

  cancelReminder: (id: string) =>
    request(`/api/v1/reminders/${id}`, { method: 'DELETE' }),

  contacts: (q?: string) => {
    const qs = q ? `?q=${encodeURIComponent(q)}` : ''
    return request<{ contacts: import('@/api/types').Contact[] }>(`/api/v1/career/contacts${qs}`)
  },

  createContact: (body: { name: string; company?: string; role?: string; notes?: string }) =>
    request<import('@/api/types').Contact>('/api/v1/career/contacts', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  deleteContact: (id: string) =>
    request(`/api/v1/career/contacts/${id}`, { method: 'DELETE' }),

  skills: (q?: string) => {
    const qs = q ? `?q=${encodeURIComponent(q)}` : ''
    return request<{ skills: import('@/api/types').Skill[] }>(`/api/v1/career/skills${qs}`)
  },

  createSkill: (name: string, level?: string) =>
    request<import('@/api/types').Skill>('/api/v1/career/skills', {
      method: 'POST',
      body: JSON.stringify({ name, level: level || '' }),
    }),

  deleteSkill: (id: string) =>
    request(`/api/v1/career/skills/${id}`, { method: 'DELETE' }),

  latestWeight: () =>
    request<import('@/api/types').WeightLog>('/api/v1/health/weight/latest'),

  recordWeight: (weight_kg: number) =>
    request<import('@/api/types').WeightLog>('/api/v1/health/weight', {
      method: 'POST',
      body: JSON.stringify({ weight_kg }),
    }),

  latestSteps: () =>
    request<import('@/api/types').StepLog>('/api/v1/health/steps/latest'),

  recordSteps: (steps: number) =>
    request<import('@/api/types').StepLog>('/api/v1/health/steps', {
      method: 'POST',
      body: JSON.stringify({ steps }),
    }),

  latestSleep: () =>
    request<import('@/api/types').SleepLog>('/api/v1/health/sleep/latest'),

  recordSleep: (duration_hours: number) =>
    request<import('@/api/types').SleepLog>('/api/v1/health/sleep', {
      method: 'POST',
      body: JSON.stringify({ duration_hours }),
    }),

  /**
   * Prefer GET /finance/overview (categories + period).
   * Cash-flow only on 404/501 for older deploys — no “API missing” copy in UI.
   */
  financeOverview: async (period: Period): Promise<FinanceOverview> => {
    const key = periodKey(period)
    try {
      const raw = await request<FinanceOverview>(`/api/v1/finance/overview?period=${key}`)
      return normalizeFinanceOverview(raw, period)
    } catch (e) {
      if (!(e instanceof ApiClientError) || (e.status !== 404 && e.status !== 501)) {
        throw e
      }
      if (!isSamePeriod(period, currentPeriod())) {
        return {
          period_label: periodFullLabel(period),
          income_cents: 0,
          expense_cents: 0,
          net_cents: 0,
          currency: 'RUB',
          categories: [],
        }
      }
      const cf = await request<{
        income_cents: number
        expense_cents: number
        net_cents: number
        currency: string
      }>('/api/v1/finance/cash-flow')

      return {
        period_label: periodFullLabel(period),
        income_cents: cf.income_cents ?? 0,
        expense_cents: cf.expense_cents ?? 0,
        net_cents: cf.net_cents ?? 0,
        currency: cf.currency || 'RUB',
        categories: [],
      }
    }
  },
}

/** Coerce overview payload so Legend/Ring always get a categories array. */
function normalizeFinanceOverview(
  raw: FinanceOverview,
  period: Period,
): FinanceOverview {
  const categories = Array.isArray(raw.categories) ? raw.categories : []
  return {
    period_label: raw.period_label || periodFullLabel(period),
    income_cents: Number(raw.income_cents) || 0,
    expense_cents: Number(raw.expense_cents) || 0,
    net_cents: Number(raw.net_cents) || 0,
    currency: raw.currency || 'RUB',
    categories: categories.map((c) => ({
      name: c.name || 'Прочее',
      amount_cents: Number(c.amount_cents) || 0,
      percent: Number(c.percent) || 0,
      color_hint: c.color_hint,
    })),
  }
}

export function enrichFinanceCategories(overview: FinanceOverview): FinanceOverview {
  if (!overview.categories || overview.categories.length === 0) return overview
  const slices = majorCategories(overview.categories, overview.expense_cents)
  return {
    ...overview,
    categories: slices.map((s) => ({
      name: s.name,
      amount_cents: s.amountCents,
      percent: Math.round(s.percent * 10) / 10,
      color_hint: s.color,
    })),
  }
}

export { ApiClientError }
