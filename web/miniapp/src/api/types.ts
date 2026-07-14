export type Task = {
  id: string
  title: string
  status: string
  priority: string
  due_date?: string
  project_ids?: string[]
}

export type PriorityItem = {
  kind: string
  title: string
  score: number
  detail: string
}

export type Sphere = {
  id: string
  name: string
  sort_order: number
  created_at: string
}

export type Project = {
  id: string
  name: string
  outcome?: string
  status: string
  target_value?: string
  current_value?: string
  unit?: string
  target_date?: string
  sphere_ids: string[]
}

export type FinanceCategory = {
  name: string
  amount_cents: number
  percent: number
  color_hint?: string
}

export type FinanceOverview = {
  period_label: string
  income_cents: number
  expense_cents: number
  net_cents: number
  currency: string
  categories: FinanceCategory[]
}

export type TokenResponse = {
  access_token: string
  expires_in: number
  token_type: string
}

export type ApiError = {
  error: string
}
