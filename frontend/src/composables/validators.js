/** Телефон: +7 и ровно 10 цифр после кода страны */
export function formatPhoneInput(value) {
  let digits = String(value ?? '').replace(/\D/g, '')
  if (digits.startsWith('8')) digits = digits.slice(1)
  if (digits.startsWith('7')) digits = digits.slice(1)
  digits = digits.slice(0, 10)
  return '+7' + digits
}

export function isValidPhone(phone) {
  return /^\+7\d{10}$/.test(phone)
}

/** Только положительные целые; при max — ограничение сверху */
export function parsePositiveInt(value, min = 1, max = null) {
  const digits = String(value ?? '').replace(/\D/g, '')
  let n = digits === '' ? min : parseInt(digits, 10)
  if (Number.isNaN(n) || n < min) n = min
  if (max != null && n > max) n = max
  return n
}

export function parseMoneyLimit(value, max = 2147483647) {
  const normalized = String(value ?? '').replace(',', '.').replace(/[^\d.]/g, '')
  const parts = normalized.split('.')
  const compact = parts.length > 1 ? `${parts[0]}.${parts.slice(1).join('').slice(0, 2)}` : parts[0]
  if (compact === '') return ''
  const n = Number(compact)
  if (!Number.isFinite(n) || n < 0) return ''
  return String(Math.min(n, max))
}

export function blockNonDigitKey(e) {
  const allowed = ['Backspace', 'Delete', 'Tab', 'ArrowLeft', 'ArrowRight', 'Home', 'End']
  if (allowed.includes(e.key)) return
  if (e.ctrlKey || e.metaKey) return
  if (!/^\d$/.test(e.key)) e.preventDefault()
}

export function blockNonDigitDecimalKey(e) {
  const allowed = ['Backspace', 'Delete', 'Tab', 'ArrowLeft', 'ArrowRight', 'Home', 'End']
  if (allowed.includes(e.key)) return
  if (e.ctrlKey || e.metaKey) return
  if (!/^[\d.,]$/.test(e.key)) e.preventDefault()
}

export function isStrongPassword(value) {
  const v = String(value ?? '')
  return /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?`~])[A-Za-z\d!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?`~]{8,72}$/.test(v)
}
