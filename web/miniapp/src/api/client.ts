import type { FinanceOverview } from '@/api/types'
import { periodKey, periodFullLabel, currentPeriod, isSamePeriod, type Period } from '@/lib/periods'
import { majorCategories } from '@/lib/categories'

export type { Period } from '@/lib/periods'

let accessToken: string | null = null

export function setAccessToken(token: string | null) {
  accessToken = token
}

export function getAccessToken() {
  return accessToken
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
): Promise<T> {
  const headers = new Headers(init.headers)
  if (!headers.has('Content-Type') && init.body) {
    headers.set('Content-Type', 'application/json')
  }
  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`)
  }

  const res = await fetch(path, { ...init, headers })
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

export async function authWithInitData(initData: string): Promise<string> {
  try {
    const data = await request<{ access_token: string }>('/api/v1/auth/telegram-webapp', {
      method: 'POST',
      body: JSON.stringify({ init_data: initData }),
    })
    return data.access_token
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
): Promise<string> {
  const data = await request<{ access_token: string }>('/api/v1/auth/token', {
    method: 'POST',
    headers: { 'X-API-Key': apiKey },
    body: JSON.stringify({ telegram_id: telegramId }),
  })
  return data.access_token
}

export const api = {
  tasksToday: () => request<{ tasks: import('@/api/types').Task[] }>('/api/v1/tasks/today'),

  priorities: () =>
    request<{ priorities: import('@/api/types').PriorityItem[] }>('/api/v1/priorities'),

  completeTask: (id: string) =>
    request<{ id: string }>(`/api/v1/tasks/${id}/complete`, { method: 'POST' }),

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

  financeOverview: async (period: Period): Promise<FinanceOverview> => {
    const key = periodKey(period)
    try {
      return await request<FinanceOverview>(`/api/v1/finance/overview?period=${key}`)
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
        income_cents: cf.income_cents,
        expense_cents: cf.expense_cents,
        net_cents: cf.net_cents,
        currency: cf.currency || 'RUB',
        categories: [],
      }
    }
  },
}

export function enrichFinanceCategories(overview: FinanceOverview): FinanceOverview {
  if (overview.categories.length === 0) return overview
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
