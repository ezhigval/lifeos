import { Routes, Route } from 'react-router-dom'
import { AppShell } from '@/components/layout/AppShell'
import { HomePage } from '@/pages/HomePage'
import { SpheresPage } from '@/pages/SpheresPage'
import { MorePage } from '@/pages/MorePage'
import { SettingsPage } from '@/pages/SettingsPage'
import { ComingSoonPage } from '@/pages/ComingSoonPage'

export default function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<HomePage />} />
        <Route path="spheres" element={<SpheresPage />} />
        <Route path="spheres/:sphereId" element={<SpheresPage />} />
        <Route path="spheres/:sphereId/projects/:projectId" element={<SpheresPage />} />
        <Route path="more" element={<MorePage />} />
        <Route path="more/settings" element={<SettingsPage />} />
        <Route
          path="more/habits"
          element={
            <ComingSoonPage
              title="Привычки"
              description="Трекинг привычек"
              hint="MA-B4 — список today + track в 2 тапа"
            />
          }
        />
        <Route
          path="more/calendar"
          element={
            <ComingSoonPage
              title="Календарь"
              description="События на сегодня"
              hint="MA-B5 — list + create sheet"
            />
          }
        />
        <Route
          path="more/analytics"
          element={
            <ComingSoonPage
              title="Аналитика"
              description="Обзоры и статистика"
              hint="MA-C1 — reviews + summary"
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
