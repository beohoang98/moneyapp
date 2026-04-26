export function formatAmount(minorUnits: number, currency = 'VND'): string {
  if (currency === 'VND') {
    return new Intl.NumberFormat('vi-VN').format(minorUnits) + ' ₫'
  }
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
  }).format(minorUnits / 100)
}

export function toMinorUnits(displayAmount: number, currency = 'VND'): number {
  if (currency === 'VND') return Math.round(displayAmount)
  return Math.round(displayAmount * 100)
}

export function toDisplayAmount(minorUnits: number, currency = 'VND'): number {
  if (currency === 'VND') return minorUnits
  return minorUnits / 100
}

export function todayString(): string {
  return new Date().toISOString().split('T')[0]
}

/** First calendar day (YYYY-MM-DD) from API date or RFC3339 datetime — for `<input type="date">`. */
export function toInputDate(value: string | undefined): string {
  if (!value) return ''
  if (/^\d{4}-\d{2}-\d{2}$/.test(value)) return value
  const head = value.slice(0, 10)
  return /^\d{4}-\d{2}-\d{2}$/.test(head) ? head : ''
}

/** Human-readable date for tables (calendar day in local timezone). */
export function formatDisplayDate(value: string | undefined): string {
  const ymd = toInputDate(value)
  if (!ymd) return value ?? ''
  const [y, m, d] = ymd.split('-').map(Number)
  return new Date(y, m - 1, d).toLocaleDateString()
}
