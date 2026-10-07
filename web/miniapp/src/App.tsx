import { lazy, type ComponentType } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { AppShell } from '@/components/layout/AppShell'
import { HomePage } from '@/pages/HomePage'

function lazyPage<T extends Record<string, ComponentType>>(
  loader: () => Promise<T>,
  name: keyof T & string,
) {
  return lazy(() => loader().then((m) => ({ default: m[name] })))
}

const TaskDetailPage = lazyPage(() => import('@/pages/TaskDetailPage'), 'TaskDetailPage')
const SpheresPage = lazyPage(() => import('@/pages/SpheresPage'), 'SpheresPage')
const MorePage = lazyPage(() => import('@/pages/MorePage'), 'MorePage')
const SettingsPage = lazyPage(() => import('@/pages/SettingsPage'), 'SettingsPage')
const HabitsPage = lazyPage(() => import('@/pages/HabitsPage'), 'HabitsPage')
const CalendarPage = lazyPage(() => import('@/pages/CalendarPage'), 'CalendarPage')
const RemindersPage = lazyPage(() => import('@/pages/RemindersPage'), 'RemindersPage')
const AnalyticsPage = lazyPage(() => import('@/pages/AnalyticsPage'), 'AnalyticsPage')
const NotesPage = lazyPage(() => import('@/pages/NotesPage'), 'NotesPage')
const HealthPage = lazyPage(() => import('@/pages/HealthPage'), 'HealthPage')
const DebtsPage = lazyPage(() => import('@/pages/DebtsPage'), 'DebtsPage')
const TriagePage = lazyPage(() => import('@/pages/TriagePage'), 'TriagePage')

export default function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<HomePage />} />
        <Route path="tasks/:taskId" element={<TaskDetailPage />} />
        <Route path="spheres" element={<SpheresPage />} />
        <Route path="spheres/:sphereId" element={<SpheresPage />} />
        <Route path="spheres/:sphereId/projects/:projectId" element={<SpheresPage />} />
        <Route path="more" element={<MorePage />} />
        <Route path="more/settings" element={<SettingsPage />} />
        <Route path="more/habits" element={<HabitsPage />} />
        <Route path="more/calendar" element={<CalendarPage />} />
        <Route path="more/reminders" element={<RemindersPage />} />
        <Route path="more/analytics" element={<AnalyticsPage />} />
        <Route path="more/notes" element={<NotesPage />} />
        <Route path="more/health" element={<HealthPage />} />
        <Route path="more/debts" element={<DebtsPage />} />
        <Route path="more/triage" element={<TriagePage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
