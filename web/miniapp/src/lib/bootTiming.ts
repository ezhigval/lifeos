/** Marks from navigation start (performance.now). Inline HTML records the early ones. */

declare global {
  interface Window {
    __LIFEOS_T?: Array<[string, number]>
    __LIFEOS_MARK?: (name: string) => void
  }
}

const reported = new Set<string>()

export function bootMark(name: string) {
  if (typeof window.__LIFEOS_MARK === 'function') {
    window.__LIFEOS_MARK(name)
    return
  }
  const rows = (window.__LIFEOS_T ??= [])
  if (rows.some(([n]) => n === name)) return
  rows.push([name, Math.round(performance.now())])
}

/** Rewrite `<html data-boot>` and log once per label. Times are ms from navigation start. */
export function bootReport(label: string) {
  const text = bootSnapshot()
  document.documentElement.setAttribute('data-boot', text)
  if (reported.has(label)) return
  reported.add(label)
  console.info(`[lifeos boot] ${label} ${text}`)
}

function bootSnapshot(): string {
  const rows = window.__LIFEOS_T ?? []
  const marks = rows
    .map(([name, t], i) => {
      const delta = t - (i > 0 ? rows[i - 1][1] : 0)
      return `${name}@${t}+${delta}`
    })
    .join(' ')

  const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming | undefined
  const navBit = nav
    ? ` nav ttfb=${Math.round(nav.responseStart)} resp=${Math.round(nav.responseEnd)} dcl=${Math.round(nav.domContentLoadedEventEnd)}`
    : ''

  const resources = performance
    .getEntriesByType('resource')
    .filter((e): e is PerformanceResourceTiming => e.entryType === 'resource')
    .filter((e) => /\/app\/.+\.(js|css)(\?|$)/.test(e.name) || e.name.includes('telegram-web-app'))
    .map((e) => {
      const file = (e.name.split('/').pop() ?? e.name).split('?')[0]
      return `${file}:${Math.round(e.duration)}`
    })
    .join(',')

  return `${marks}${navBit}${resources ? ` res ${resources}` : ''}`
}
