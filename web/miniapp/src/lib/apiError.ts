/** Map API / network errors to short RU copy for sheets and banners. */
export function ruApiError(err: unknown, fallback = 'Не удалось сохранить'): string {
  if (!(err instanceof Error) || !err.message) return fallback
  const m = err.message.toLowerCase()
  if (m.includes('linked projects') || m.includes('has linked')) {
    return 'Сфера связана с проектами — сначала отвяжи их'
  }
  if (m.includes('required')) return 'Заполни обязательные поля'
  if (m.includes('not found')) return 'Не найдено'
  if (m.includes('invalid starts_at') || m.includes('rfc3339')) {
    return 'Некорректное время события'
  }
  if (m.includes('unauthorized') || m.includes('401')) {
    return 'Сессия истекла — открой Mini App заново'
  }
  // Prefer known short backend messages; skip huge stack-like strings.
  if (err.message.length <= 120 && !err.message.includes('\n')) {
    return err.message
  }
  return fallback
}
