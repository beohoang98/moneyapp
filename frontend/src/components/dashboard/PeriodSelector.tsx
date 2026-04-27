import { useState } from 'react'

interface Props {
  dateFrom: string
  dateTo: string
  onChange: (dateFrom: string, dateTo: string) => void
}

type Preset = 'this_month' | 'last_month' | 'this_year' | 'custom'

function getPresetDates(preset: Preset): { from: string; to: string } | null {
  const now = new Date()
  const y = now.getFullYear()
  const m = now.getMonth()

  switch (preset) {
    case 'this_month': {
      const from = new Date(y, m, 1)
      const to = new Date(y, m + 1, 0)
      return { from: fmt(from), to: fmt(to) }
    }
    case 'last_month': {
      const from = new Date(y, m - 1, 1)
      const to = new Date(y, m, 0)
      return { from: fmt(from), to: fmt(to) }
    }
    case 'this_year': {
      const from = new Date(y, 0, 1)
      return { from: fmt(from), to: fmt(now) }
    }
    default:
      return null
  }
}

function fmt(d: Date): string {
  // Use local calendar date (avoid UTC offset shifting day).
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function detectPreset(dateFrom: string, dateTo: string): Preset {
  for (const p of ['this_month', 'last_month', 'this_year'] as const) {
    const dates = getPresetDates(p)
    if (dates && dates.from === dateFrom && dates.to === dateTo) return p
  }
  if (dateFrom || dateTo) return 'custom'
  return 'this_month'
}

export function PeriodSelector({ dateFrom, dateTo, onChange }: Props) {
  const activePreset = detectPreset(dateFrom, dateTo)
  const [customFrom, setCustomFrom] = useState(dateFrom)
  const [customTo, setCustomTo] = useState(dateTo)
  const [validationError, setValidationError] = useState('')

  const handlePreset = (preset: Preset) => {
    const dates = getPresetDates(preset)
    if (dates) {
      setValidationError('')
      onChange(dates.from, dates.to)
    }
  }

  const handleApplyCustom = () => {
    if (customFrom && customTo && customFrom > customTo) {
      setValidationError('Start date must be before end date')
      return
    }
    setValidationError('')
    onChange(customFrom, customTo)
  }

  const handleClear = () => {
    setCustomFrom('')
    setCustomTo('')
    setValidationError('')
    const dates = getPresetDates('this_month')!
    onChange(dates.from, dates.to)
  }

  const presets: { key: Preset; label: string }[] = [
    { key: 'this_month', label: 'This Month' },
    { key: 'last_month', label: 'Last Month' },
    { key: 'this_year', label: 'This Year' },
  ]

  return (
    <div className="period-selector">
      <div className="period-presets">
        {presets.map(({ key, label }) => (
          <button
            key={key}
            className={`btn btn-sm ${activePreset === key ? 'btn-primary' : ''}`}
            aria-pressed={activePreset === key}
            onClick={() => handlePreset(key)}
          >
            {label}
          </button>
        ))}
      </div>
      <div className="period-custom">
        <input
          type="date"
          value={customFrom}
          onChange={(e) => setCustomFrom(e.target.value)}
          aria-label="Start date"
        />
        <span style={{ color: 'var(--text)' }}>to</span>
        <input
          type="date"
          value={customTo}
          onChange={(e) => setCustomTo(e.target.value)}
          aria-label="End date"
        />
        <button className="btn btn-sm" onClick={handleApplyCustom}>Apply</button>
        <button className="btn btn-sm" onClick={handleClear}>Clear</button>
      </div>
      {validationError && (
        <div style={{ color: '#ef4444', fontSize: 13, marginTop: 4 }}>{validationError}</div>
      )}
    </div>
  )
}
