import { Routes, Route } from 'react-router-dom'
import { AppShell } from '@/components/layout/AppShell'
import { HomePage } from '@/pages/HomePage'
import { SpheresPage } from '@/pages/SpheresPage'
import { MorePage } from '@/pages/MorePage'
import { SettingsPage } from '@/pages/SettingsPage'
import { HabitsPage } from '@/pages/HabitsPage'
import { CalendarPage } from '@/pages/CalendarPage'
import { TaskDetailPage } from '@/pages/TaskDetailPage'
import { ComingSoonPage } from '@/pages/ComingSoonPage'

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
        <Route
          path="more/analytics"
          element={
            <ComingSoonPage
              title="Аналитика"
              description="Обзоры и статистика"
              hint="MA-C1 — после MVP"
            />
          }
        />
        <Route
          path="more/health"
          element={
            <ComingSoonPage
              title="Здоровье"
              description="Вес, шаги, сон"
              hint="MA-C3"
            />
          }
        />
        <Route
          path="more/career"
          element={
            <ComingSoonPage
              title="Карьера"
              description="Контакты и навыки"
              hint="MA-C4"
            />
          }
        />
        <Route
          path="more/notes"
          element={
            <ComingSoonPage
              title="Заметки"
              description="Inbox заметок"
              hint="MA-C2"
            />
          }
        />
      </Route>
    </Routes>
  )
}
