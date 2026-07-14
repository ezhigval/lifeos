/** Map API / network errors to short RU copy for sheets and banners. */
export function ruApiError(err: unknown, fallback = 'Не удалось сохранить'): string {
  if (!(err instanceof Error) || !err.message) return fallback
  const m = err.message.toLowerCase()
  if (m.includes('linked projects') || m.includes('has linked')) {
    return 'Сфера связана с проектами — сначала отвяжи их'
  }
  if (m.includes('fire_at must be in the future')) {
    return 'Время напоминания должно быть в будущем'
  }
  if (m.includes('invalid fire_at') || m.includes('rfc3339')) {
    return 'Некорректное время'
  }
  if (m.includes('invalid starts_at')) {
    return 'Некорректное время события'
  }
  if (m.includes('payment exceeds') || m.includes('overpayment')) {
    return 'Сумма больше остатка долга'
  }
  if (m.includes('debt is not open')) {
    return 'Долг уже закрыт'
  }
  if (m.includes('invalid due_date')) {
    return 'Дата в формате ГГГГ-ММ-ДД'
  }
  if (m.includes('invalid weight')) return 'Вес: реалистичное значение в кг'
  if (m.includes('invalid steps')) return 'Укажи число шагов'
  if (m.includes('invalid sleep') || m.includes('duration_hours')) {
    return 'Сон: часы, например 7.5 (до 24)'
  }
  if (m.includes('invalid health')) return 'Введи положительное число'
  if (m.includes('required')) return 'Заполни обязательные поля'
  if (m.includes('not found')) return 'Не найдено'
  if (m.includes('unauthorized') || m.includes('401')) {
    return 'Сессия истекла — открой Mini App заново'
  }
  // Prefer known short backend messages; skip huge stack-like strings.
  if (err.message.length <= 120 && !err.message.includes('\n')) {
    return err.message
  }
  return fallback
}
