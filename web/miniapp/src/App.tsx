import { Routes, Route } from 'react-router-dom'
import { AppShell } from '@/components/layout/AppShell'
import { HomePage } from '@/pages/HomePage'
import { SpheresPage } from '@/pages/SpheresPage'

export default function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<HomePage />} />
        <Route path="spheres" element={<SpheresPage />} />
        <Route path="spheres/:sphereId" element={<SpheresPage />} />
        <Route path="spheres/:sphereId/projects/:projectId" element={<SpheresPage />} />
      </Route>
    </Routes>
  )
}
