import { lazy, Suspense, type ReactNode } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { AppShell } from '@/components/layout/AppShell'
import { HomePage } from '@/pages/HomePage'

const SpheresPage = lazy(() => import('@/pages/SpheresPage').then((m) => ({ default: m.SpheresPage })))
const MorePage = lazy(() => import('@/pages/MorePage').then((m) => ({ default: m.MorePage })))
const SettingsPage = lazy(() => import('@/pages/SettingsPage').then((m) => ({ default: m.SettingsPage })))
const HabitsPage = lazy(() => import('@/pages/HabitsPage').then((m) => ({ default: m.HabitsPage })))
const CalendarPage = lazy(() => import('@/pages/CalendarPage').then((m) => ({ default: m.CalendarPage })))
const RemindersPage = lazy(() => import('@/pages/RemindersPage').then((m) => ({ default: m.RemindersPage })))
const TaskDetailPage = lazy(() => import('@/pages/TaskDetailPage').then((m) => ({ default: m.TaskDetailPage })))
const AnalyticsPage = lazy(() => import('@/pages/AnalyticsPage').then((m) => ({ default: m.AnalyticsPage })))
const NotesPage = lazy(() => import('@/pages/NotesPage').then((m) => ({ default: m.NotesPage })))
const HealthPage = lazy(() => import('@/pages/HealthPage').then((m) => ({ default: m.HealthPage })))
const DebtsPage = lazy(() => import('@/pages/DebtsPage').then((m) => ({ default: m.DebtsPage })))
const TriagePage = lazy(() => import('@/pages/TriagePage').then((m) => ({ default: m.TriagePage })))

function PageFallback() {
  return (
    <div style={{ padding: '2rem 1.25rem', color: '#94a3b8', fontSize: 15 }}>Загрузка…</div>
  )
}

function whenOpen(page: ReactNode) {
  return <Suspense fallback={<PageFallback />}>{page}</Suspense>
}

export default function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<HomePage />} />
        <Route path="tasks/:taskId" element={whenOpen(<TaskDetailPage />)} />
        <Route path="spheres" element={whenOpen(<SpheresPage />)} />
        <Route path="spheres/:sphereId" element={whenOpen(<SpheresPage />)} />
        <Route path="spheres/:sphereId/projects/:projectId" element={whenOpen(<SpheresPage />)} />
        <Route path="more" element={whenOpen(<MorePage />)} />
        <Route path="more/settings" element={whenOpen(<SettingsPage />)} />
        <Route path="more/habits" element={whenOpen(<HabitsPage />)} />
        <Route path="more/calendar" element={whenOpen(<CalendarPage />)} />
        <Route path="more/reminders" element={whenOpen(<RemindersPage />)} />
        <Route path="more/analytics" element={whenOpen(<AnalyticsPage />)} />
        <Route path="more/notes" element={whenOpen(<NotesPage />)} />
        <Route path="more/health" element={whenOpen(<HealthPage />)} />
        <Route path="more/debts" element={whenOpen(<DebtsPage />)} />
        <Route path="more/triage" element={whenOpen(<TriagePage />)} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
