const MESSAGES = {
  'invalid login or password': 'Неверный логин или пароль',
  'invalid email or password': 'Неверный email или пароль',
  'not found': 'Запись не найдена',
  'product not found': 'Товар не найден'
}

export function mapApiError(err, fallback = 'Ошибка запроса') {
  const raw = err?.response?.data?.error
  if (!raw || typeof raw !== 'string') return fallback
  return MESSAGES[raw] || raw
}
