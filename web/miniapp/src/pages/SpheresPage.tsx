import { useParams } from 'react-router-dom'
import { Header } from '@/components/layout/Header'
import {
  SphereTree,
  SphereDetailPage,
  ProjectDetailPage,
} from '@/components/spheres/SphereViews'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'

export function SpheresPage() {
  const { sphereId, projectId } = useParams()

  const { data: spheres } = useQuery({
    queryKey: ['spheres'],
    queryFn: async () => {
      const res = await api.spheres()
      return Array.isArray(res.spheres) ? res.spheres : []
    },
  })

  const sphere = spheres?.find((s) => s.id === sphereId)
  const { data: projects } = useQuery({
    queryKey: ['projects', sphereId],
    queryFn: async () => {
      const res = await api.projects(sphereId!)
      return Array.isArray(res.projects) ? res.projects : []
    },
    enabled: Boolean(sphereId),
  })
  const project = projects?.find((p) => p.id === projectId)

  let title = 'Сферы'
  let subtitle: string | undefined = 'Сфера → Проект → Задача'
  if (project) {
    title = project.name
    subtitle = sphere?.name
  } else if (sphere) {
    title = sphere.name
    subtitle = 'Проекты сферы'
  }

  return (
    <>
      <Header title={title} subtitle={subtitle} />
      {!sphereId && <SphereTree />}
      {sphereId && !projectId && <SphereDetailPage />}
      {sphereId && projectId && <ProjectDetailPage />}
    </>
  )
}
