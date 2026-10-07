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

/** Empty in the Mini App (same origin). Desktop sets VITE_API_BASE when it cannot proxy. */
export function apiUrl(path: string): string {
  const base = import.meta.env.VITE_API_BASE
  if (!base) return path
  return String(base).replace(/\/$/, '') + path
}

const REQUEST_TIMEOUT_MS = 15_000

function isAbortError(err: unknown): boolean {
  return (
    (err instanceof DOMException && err.name === 'AbortError') ||
    (err instanceof Error && err.name === 'AbortError')
  )
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

  const ctrl = new AbortController()
  const timer = setTimeout(() => ctrl.abort(), REQUEST_TIMEOUT_MS)
  const parent = init.signal
  if (parent) {
    if (parent.aborted) ctrl.abort()
    else parent.addEventListener('abort', () => ctrl.abort(), { once: true })
  }

  let res: Response
  try {
    res = await fetch(apiUrl(path), { ...init, headers, signal: ctrl.signal })
  } catch (err) {
    if (isAbortError(err)) throw new ApiClientError(0, 'Сервер не отвечает')
    throw err
  } finally {
    clearTimeout(timer)
  }
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
      throw new Error('На этом сервере нет входа Mini App (/api/v1/auth/telegram-webapp)')
    }
    throw e
  }
}

export async function requestTelegramLoginCode(username: string): Promise<void> {
  await request(
    '/api/v1/auth/telegram-login/request',
    {
      method: 'POST',
      body: JSON.stringify({ username }),
    },
    false,
  )
}

export async function verifyTelegramLoginCode(username: string, code: string): Promise<AuthResult> {
  const data = await request<{
    access_token: string
    expires_in?: number
    telegram_id?: number
  }>(
    '/api/v1/auth/telegram-login/verify',
    {
      method: 'POST',
      body: JSON.stringify({ username, code }),
    },
    false,
  )
  return {
    accessToken: data.access_token,
    expiresIn: data.expires_in ?? 0,
    telegramId:
      typeof data.telegram_id === 'number' && data.telegram_id > 0 ? data.telegram_id : undefined,
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

  reopenTask: (id: string) =>
    request<import('@/api/types').Task>(`/api/v1/tasks/${id}/reopen`, { method: 'POST' }),

  getTask: (id: string) => request<import('@/api/types').Task>(`/api/v1/tasks/${id}`),

  tasksDueBetween: (from: string, to: string) =>
    request<{ tasks: import('@/api/types').Task[] }>(
      `/api/v1/tasks?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
    ),

  updateTask: (id: string, body: {
    title?: string
    priority?: string
    kind?: 'task' | 'reminder' | 'meeting'
    due_date?: string
    clear_due_date?: boolean
    address?: string
    clear_address?: boolean
    note_id?: string
    clear_note_id?: boolean
    description?: string
    clear_description?: boolean
    duration_minutes?: number
    clear_duration?: boolean
    tags?: string[]
    project_ids?: string[]
    sphere_ids?: string[]
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
    kind?: 'task' | 'reminder' | 'meeting'
    due_date?: string
    address?: string
    note_id?: string
    clear_address?: boolean
    clear_note_id?: boolean
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

  createHabit: (input: { name: string; start_date?: string | null; end_date?: string | null }) =>
    request<import('@/api/types').Habit>('/api/v1/habits', {
      method: 'POST',
      body: JSON.stringify(input),
    }),

  updateHabit: (
    id: string,
    input: { name?: string; start_date?: string | null; end_date?: string | null },
  ) =>
    request<import('@/api/types').Habit>(`/api/v1/habits/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
    }),

  deleteHabit: (id: string) =>
    request<void>(`/api/v1/habits/${id}`, {
      method: 'DELETE',
    }),

  trackHabit: (id: string) =>
    request<{ name: string; streak: number }>(`/api/v1/habits/${id}/track`, {
      method: 'POST',
    }),

  calendarToday: (day?: string) =>
    request<{ events: import('@/api/types').CalendarEvent[] }>(
      `/api/v1/calendar/today${day ? `?day=${day}` : ''}`,
    ),

  calendarAgenda: (params: {
    view: 'day' | 'week' | 'month'
    from: string
    to: string
    types?: string
    projects?: string
    spheres?: string
  }) => {
    const q = new URLSearchParams({ view: params.view, from: params.from, to: params.to })
    if (params.types) q.set('types', params.types)
    if (params.projects) q.set('projects', params.projects)
    if (params.spheres) q.set('spheres', params.spheres)
    return request<import('@/api/types').AgendaResponse>(`/api/v1/calendar/agenda?${q.toString()}`)
  },

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

  // MA-C6: overloaded-day triage (proposal + defer low-priority tasks to tomorrow).
  triageProposal: () =>
    request<{ text: string; low_priority_ids: string[] }>('/api/v1/planning/triage'),

  triageDefer: (task_ids: string[]) =>
    request<{ moved: number }>('/api/v1/planning/triage/defer', {
      method: 'POST',
      body: JSON.stringify({ task_ids }),
    }),

  analyticsSummary: () =>
    request<import('@/api/types').AnalyticsSummary>('/api/v1/analytics/summary'),

  notes: (q?: string) => {
    const qs = q ? `?q=${encodeURIComponent(q)}` : ''
    return request<{ notes: import('@/api/types').Note[] }>(`/api/v1/notes${qs}`)
  },

  /** TASK-011 item 5: reverse sync — notes linked to a task/event/reminder */
  notesByTarget: (targetType: 'task' | 'event' | 'reminder', targetId: string) =>
    request<{ notes: import('@/api/types').Note[] }>(
      `/api/v1/notes?target_type=${targetType}&target_id=${encodeURIComponent(targetId)}`,
    ),

  createNote: (body: string, tags?: string[], target?: { type: 'task' | 'event' | 'reminder'; id: string }) =>
    request<import('@/api/types').Note>('/api/v1/notes', {
      method: 'POST',
      body: JSON.stringify({
        body,
        tags: tags ?? [],
        target_type: target?.type ?? '',
        target_id: target?.id ?? '',
      }),
    }),

  getNote: (id: string) =>
    request<import('@/api/types').Note>(`/api/v1/notes/${id}`),

  updateNote: (id: string, body: { body?: string; tags?: string[] }) =>
    request<import('@/api/types').Note>(`/api/v1/notes/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),

  deleteNote: (id: string) =>
    request<import('@/api/types').Note>(`/api/v1/notes/${id}`, { method: 'DELETE' }),

  debts: () => request<{ debts: import('@/api/types').Debt[] }>('/api/v1/finance/debts'),

  createDebt: (
    creditor: string,
    amount_cents: number,
    due_date?: string,
    opts?: {
      installment_cents?: number
      installment_interval?: 'none' | 'weekly' | 'monthly'
      next_payment_date?: string
    },
  ) =>
    request<import('@/api/types').Debt>('/api/v1/finance/debts', {
      method: 'POST',
      body: JSON.stringify({ creditor, amount_cents, due_date, ...opts }),
    }),

  payDebt: (id: string, amount_cents: number, regular?: boolean) =>
    request(`/api/v1/finance/debts/${id}/pay`, {
      method: 'POST',
      body: JSON.stringify({ amount_cents, ...(regular != null ? { regular } : {}) }),
    }),

  financePlan: () =>
    request<import('@/api/types').FinancePlan>('/api/v1/finance/plan'),

  createFinancePlan: (body: {
    kind: 'income' | 'expense'
    title: string
    amount_cents: number
    currency?: string
    interval: string
    next_date: string
  }) =>
    request<import('@/api/types').FinancePlanItem>('/api/v1/finance/plan', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  deleteFinancePlan: (id: string) =>
    request<void>(`/api/v1/finance/plan/${id}`, { method: 'DELETE' }),

  completeFinancePlan: (id: string) =>
    request<{
      deleted: boolean
      posted?: boolean
      posted_cents?: number
      posted_kind?: string
      item?: import('@/api/types').FinancePlanItem
    }>(`/api/v1/finance/plan/${id}/complete`, { method: 'POST' }),

  reminders: () =>
    request<{ reminders: import('@/api/types').Reminder[] }>('/api/v1/reminders'),

  createReminder: (message: string, fire_at: string) =>
    request<import('@/api/types').Reminder>('/api/v1/reminders', {
      method: 'POST',
      body: JSON.stringify({ message, fire_at }),
    }),

  cancelReminder: (id: string) =>
    request(`/api/v1/reminders/${id}`, { method: 'DELETE' }),

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
