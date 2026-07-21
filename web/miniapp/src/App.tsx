import { Routes, Route, Navigate } from 'react-router-dom'
import { AppShell } from '@/components/layout/AppShell'
import { HomePage } from '@/pages/HomePage'
import { SpheresPage } from '@/pages/SpheresPage'
import { MorePage } from '@/pages/MorePage'
import { SettingsPage } from '@/pages/SettingsPage'
import { HabitsPage } from '@/pages/HabitsPage'
import { CalendarPage } from '@/pages/CalendarPage'
import { RemindersPage } from '@/pages/RemindersPage'
import { TaskDetailPage } from '@/pages/TaskDetailPage'
import { AnalyticsPage } from '@/pages/AnalyticsPage'
import { NotesPage } from '@/pages/NotesPage'
import { HealthPage } from '@/pages/HealthPage'
// import { CareerPage } from '@/pages/CareerPage'
import { DebtsPage } from '@/pages/DebtsPage'

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
        {/* <Route path="more/career" element={<CareerPage />} /> */}
        <Route path="more/career" element={<Navigate to="/more" replace />} />
        <Route path="more/debts" element={<DebtsPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
