import { useState, useEffect, useRef } from 'react'
import { scanInvoice, deleteTempScan } from '../../api/scanning'
import { createInvoice } from '../../api/invoices'
import { formatAmount } from '../../utils/format'
import type { ScanResult } from '../../types/scanning'
import { ApiClientError } from '../../api/client'

interface Props {
  onClose: () => void
  onInvoiceCreated: () => void
}

type Step = 'pick' | 'scanning' | 'review' | 'error'

const ERROR_MESSAGES: Record<string, string> = {
  scanning_disabled: 'Document scanning is not configured. Please go to Settings to enable it.',
  extraction_failed: 'Could not extract data from this image. Please check the image is a clear photo of a receipt or invoice, and try again.',
  scan_timeout: 'Scanning timed out. Make sure Ollama is running and the model is loaded (`ollama run qwen3-vl:4b`).',
  too_many_scans: 'Another scan is in progress. Please wait a moment and try again.',
}

export function ScanModal({ onClose, onInvoiceCreated }: Props) {
  const [step, setStep] = useState<Step>('pick')
  const [scanResult, setScanResult] = useState<ScanResult | null>(null)
  const [tempStorageKey, setTempStorageKey] = useState('')
  const [errorMessage, setErrorMessage] = useState('')
  const [elapsedSeconds, setElapsedSeconds] = useState(0)
  const [saving, setSaving] = useState(false)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const abortRef = useRef<AbortController | null>(null)

  const [vendor, setVendor] = useState('')
  const [date, setDate] = useState('')
  const [totalAmount, setTotalAmount] = useState('')
  const [currency, setCurrency] = useState('VND')
  const [description, setDescription] = useState('')

  const handleCancel = async () => {
    if (abortRef.current) {
      abortRef.current.abort()
      abortRef.current = null
    }
    if (tempStorageKey) {
      try {
        await deleteTempScan(tempStorageKey)
      } catch { /* best effort */ }
    }
    onClose()
  }

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') handleCancel()
    }
    window.addEventListener('keydown', onKey)
    return () => {
      window.removeEventListener('keydown', onKey)
    }
  })

  useEffect(() => {
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [])

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    const validTypes = ['image/jpeg', 'image/png', 'image/webp']
    if (!validTypes.includes(file.type)) {
      setErrorMessage('Unsupported file type. Please use JPEG, PNG, or WebP.')
      setStep('error')
      return
    }
    if (file.size > 10 * 1024 * 1024) {
      setErrorMessage('File is too large. Maximum size is 10 MB.')
      setStep('error')
      return
    }

    setStep('scanning')
    setElapsedSeconds(0)
    timerRef.current = setInterval(() => setElapsedSeconds((s) => s + 1), 1000)

    const controller = new AbortController()
    abortRef.current = controller

    try {
      const response = await scanInvoice(file, controller.signal)
      if (timerRef.current) clearInterval(timerRef.current)

      setScanResult(response.scan_result)
      setTempStorageKey(response.temp_storage_key)
      setVendor(response.scan_result.vendor || '')
      setDate(response.scan_result.date || '')
      setTotalAmount(response.scan_result.total_amount ? String(response.scan_result.total_amount) : '')
      setCurrency(response.scan_result.currency || 'VND')
      setStep('review')
    } catch (err) {
      if (timerRef.current) clearInterval(timerRef.current)

      if (err instanceof DOMException && err.name === 'AbortError') {
        return
      }

      let msg = 'Could not reach the server. Check your connection.'
      if (err instanceof ApiClientError) {
        try {
          const body = JSON.parse(err.message)
          const code = body.error || err.message
          msg = ERROR_MESSAGES[code] || err.message
        } catch {
          msg = ERROR_MESSAGES[err.message] || err.message
        }
      }
      setErrorMessage(msg)
      setStep('error')
    }
  }

  const handleSave = async () => {
    if (!vendor || !date || !totalAmount) return
    setSaving(true)
    try {
      const invoice = await createInvoice({
        vendor_name: vendor,
        amount: parseInt(totalAmount, 10),
        issue_date: date,
        due_date: date,
        status: 'unpaid',
        description,
      })

      if (tempStorageKey) {
        const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api'
        const token = localStorage.getItem('moneyapp_token')
        const formData = new FormData()
        formData.append('entity_type', 'invoice')
        formData.append('entity_id', String(invoice.id))
        formData.append('source_storage_key', tempStorageKey)
        await fetch(`${BASE_URL}/attachments`, {
          method: 'POST',
          headers: token ? { Authorization: `Bearer ${token}` } : {},
          body: formData,
        })
      }

      onInvoiceCreated()
    } catch {
      setErrorMessage('Failed to create invoice. Please try again.')
      setSaving(false)
    }
  }

  const canSave = vendor.trim() !== '' && date.trim() !== '' && totalAmount.trim() !== '' && parseInt(totalAmount, 10) > 0

  const confidenceIcon = (field: string) => {
    if (!scanResult?.confidence?.[field]) return null
    if (scanResult.confidence[field] === 'low') {
      return <span title="Low confidence — please verify" style={{ color: '#f59e0b', marginLeft: 4 }}>⚠</span>
    }
    return null
  }

  return (
    <div className="form-modal-overlay" onClick={handleCancel}>
      <div className="form-modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 540 }}>
        {step === 'pick' && (
          <>
            <h2>Scan Invoice</h2>
            <p style={{ color: 'var(--text)', fontSize: 14, marginBottom: 16 }}>
              Select a receipt or invoice image (JPEG, PNG, or WebP, max 10 MB).
            </p>
            <input
              type="file"
              accept="image/jpeg,image/png,image/webp"
              onChange={handleFileSelect}
              data-testid="scan-file-input"
            />
            <div className="form-actions">
              <button className="btn btn-sm" onClick={handleCancel}>Cancel</button>
            </div>
          </>
        )}

        {step === 'scanning' && (
          <div style={{ textAlign: 'center', padding: '40px 20px' }}>
            <h2>Scanning image...</h2>
            <p style={{ color: 'var(--text)', fontSize: 14 }}>
              Elapsed: {elapsedSeconds}s
            </p>
            <div className="loading">Processing...</div>
          </div>
        )}

        {step === 'error' && (
          <>
            <h2>Scan Error</h2>
            <p style={{ color: '#ef4444', marginBottom: 16 }}>{errorMessage}</p>
            <input
              type="file"
              accept="image/jpeg,image/png,image/webp"
              onChange={handleFileSelect}
              data-testid="scan-file-input"
            />
            <div className="form-actions">
              <button className="btn btn-sm" onClick={handleCancel}>Cancel</button>
            </div>
          </>
        )}

        {step === 'review' && scanResult && (
          <div data-testid="scan-review-form">
            <h2>Review Scanned Data</h2>

            <div className="settings-field">
              <label>Vendor {confidenceIcon('vendor')}</label>
              <input type="text" value={vendor} onChange={(e) => setVendor(e.target.value)} />
            </div>

            <div className="settings-field">
              <label>Date {confidenceIcon('date')}</label>
              <input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
            </div>

            <div className="settings-field">
              <label>Total Amount {confidenceIcon('total_amount')}</label>
              <input
                type="number"
                value={totalAmount}
                onChange={(e) => setTotalAmount(e.target.value)}
              />
              {totalAmount && <span style={{ color: 'var(--text)', fontSize: 12 }}>{formatAmount(parseInt(totalAmount, 10) || 0)}</span>}
            </div>

            <div className="settings-field">
              <label>Currency</label>
              <input type="text" value={currency} onChange={(e) => setCurrency(e.target.value)} />
            </div>

            <div className="settings-field">
              <label>Notes</label>
              <input type="text" value={description} onChange={(e) => setDescription(e.target.value)} />
            </div>

            {scanResult.line_items && scanResult.line_items.length > 0 && (
              <div style={{ marginTop: 16 }}>
                <label style={{ fontWeight: 500, fontSize: 13 }}>Line Items</label>
                <table className="data-table" style={{ marginTop: 8 }}>
                  <thead>
                    <tr>
                      <th>Description</th>
                      <th style={{ textAlign: 'right' }}>Amount</th>
                    </tr>
                  </thead>
                  <tbody>
                    {scanResult.line_items.map((item, i) => (
                      <tr key={i}>
                        <td>{item.description}</td>
                        <td className="amount">{formatAmount(item.amount)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            <div className="form-actions">
              <button className="btn btn-sm" onClick={handleCancel}>Cancel</button>
              <button
                className="btn btn-primary btn-sm"
                onClick={handleSave}
                disabled={!canSave || saving}
              >
                {saving ? 'Creating...' : 'Create Invoice'}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
