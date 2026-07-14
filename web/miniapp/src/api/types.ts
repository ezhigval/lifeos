export type Task = {
  id: string
  title: string
  description?: string
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

export type HabitDay = {
  id: string
  name: string
  today_completed: boolean
  streak: number
}

export type CalendarEvent = {
  id: string
  title: string
  starts_at: string
}

export type TimeOfDay = { hour: number; minute: number }

export type UserSettings = {
  morning_review_at: TimeOfDay
  evening_review_at: TimeOfDay
  weekly_review_at: TimeOfDay
  monthly_review_at: TimeOfDay
  quiet_hours_start: TimeOfDay | null
  quiet_hours_end: TimeOfDay | null
  language: string
}

export type AnalyticsSummary = {
  period_label: string
  tasks_created: number
  tasks_completed: number
  completion_rate: number
  open_tasks: number
  habit_consistency: number
  habit_completions: number
  habit_count: number
  projects: { Title?: string; title?: string; Percent?: string; percent?: string }[]
}

export type Note = {
  id: string
  body: string
  tags: string[]
  created_at: string
}

export type Debt = {
  id: string
  creditor: string
  amount_cents: number
  paid_cents: number
  remaining_cents: number
  currency: string
  due_date?: string
}

export type Reminder = {
  id: string
  message: string
  fire_at: string
  status: string
}

export type Contact = {
  id: string
  name: string
  company: string
  role: string
  notes: string
  created_at: string
}

export type Skill = {
  id: string
  name: string
  level: string
  created_at: string
}

export type WeightLog = {
  id: string
  weight_kg: number
  logged_at: string
}

export type StepLog = {
  id: string
  steps: number
  logged_at: string
}

export type SleepLog = {
  id: string
  duration_minutes: number
  duration_hours: number
  logged_at: string
}

export type TokenResponse = {
  access_token: string
  expires_in: number
  token_type: string
}

export type ApiError = {
  error: string
}
