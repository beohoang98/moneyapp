import { useState } from 'react'

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api'
const TOKEN_KEY = 'moneyapp_token'

interface Props {
  type: 'expense' | 'income'
  dateFrom?: string
  dateTo?: string
  categoryId?: number
}

export function ExportButton({ type, dateFrom, dateTo, categoryId }: Props) {
  const [loading, setLoading] = useState(false)

  const handleExport = async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams()
      params.set('type', type)
      if (dateFrom) params.set('date_from', dateFrom)
      if (dateTo) params.set('date_to', dateTo)
      if (categoryId) params.set('category_id', String(categoryId))

      const token = localStorage.getItem(TOKEN_KEY)
      const res = await fetch(`${BASE_URL}/export/transactions?${params.toString()}`, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      })

      if (!res.ok) {
        throw new Error('Export failed')
      }

      const disposition = res.headers.get('Content-Disposition') ?? ''
      const match = disposition.match(/filename="?([^"]+)"?/)
      const filename = match?.[1] ?? `${type}s_export.csv`

      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch {
      // silently fail — the toast is in the caller if needed
    } finally {
      setLoading(false)
    }
  }

  return (
    <button className="btn btn-sm" onClick={handleExport} disabled={loading}>
      {loading ? 'Exporting...' : 'Export CSV'}
    </button>
  )
}
