const EXPANDED_SPHERES = 'lifeos:expandedSpheres'
const EXPANDED_PROJECTS = 'lifeos:expandedProjects'
const SELECTED_PERIOD = 'lifeos:selectedPeriod'

function readJSON<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key)
    if (!raw) return fallback
    return JSON.parse(raw) as T
  } catch {
    return fallback
  }
}

function writeJSON(key: string, value: unknown) {
  localStorage.setItem(key, JSON.stringify(value))
}

export function getExpandedSpheres(): string[] {
  return readJSON<string[]>(EXPANDED_SPHERES, [])
}

export function setExpandedSpheres(ids: string[]) {
  writeJSON(EXPANDED_SPHERES, ids)
}

export function toggleExpandedSphere(id: string): string[] {
  const set = new Set(getExpandedSpheres())
  if (set.has(id)) set.delete(id)
  else set.add(id)
  const next = [...set]
  setExpandedSpheres(next)
  return next
}

export function getExpandedProjects(): string[] {
  return readJSON<string[]>(EXPANDED_PROJECTS, [])
}

export function toggleExpandedProject(id: string): string[] {
  const set = new Set(getExpandedProjects())
  if (set.has(id)) set.delete(id)
  else set.add(id)
  const next = [...set]
  setExpandedProjects(next)
  return next
}

export function setExpandedProjects(ids: string[]) {
  writeJSON(EXPANDED_PROJECTS, ids)
}

export function getSavedPeriod(): string | null {
  return localStorage.getItem(SELECTED_PERIOD)
}

export function savePeriod(key: string) {
  localStorage.setItem(SELECTED_PERIOD, key)
}
