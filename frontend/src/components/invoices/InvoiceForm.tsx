import { useState, type FormEvent } from 'react'
import type { Invoice } from '../../types/invoice'
import { todayString, toDisplayAmount, toInputDate, toMinorUnits } from '../../utils/format'

interface InvoiceFormProps {
  invoice?: Invoice
  onSubmit: (data: { vendor_name: string; amount: number; issue_date: string; due_date: string; status: string; description: string }) => Promise<void>
  onCancel: () => void
}

export function InvoiceForm({ invoice, onSubmit, onCancel }: InvoiceFormProps) {
  const [vendorName, setVendorName] = useState(invoice?.vendor_name ?? '')
  const [amount, setAmount] = useState(invoice ? String(toDisplayAmount(invoice.amount)) : '')
  const [issueDate, setIssueDate] = useState(
    () => (invoice ? toInputDate(invoice.issue_date) || todayString() : todayString()),
  )
  const [dueDate, setDueDate] = useState(() => (invoice ? toInputDate(invoice.due_date) || '' : ''))
  const [status, setStatus] = useState<string>(invoice?.status ?? 'unpaid')
  const [description, setDescription] = useState(invoice?.description ?? '')
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(false)

  const validate = () => {
    const errs: Record<string, string> = {}
    if (!vendorName.trim()) errs.vendor_name = 'Vendor name is required'
    const numAmount = Number(amount)
    if (!amount || numAmount <= 0) errs.amount = 'Amount must be greater than zero'
    if (!issueDate) errs.issue_date = 'Issue date is required'
    if (!dueDate) errs.due_date = 'Due date is required'
    if (issueDate && dueDate && dueDate < issueDate) errs.due_date = 'Due date must be on or after issue date'
    setErrors(errs)
    return Object.keys(errs).length === 0
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!validate()) return
    setLoading(true)
    try {
      await onSubmit({
        vendor_name: vendorName.trim(),
        amount: toMinorUnits(Number(amount)),
        issue_date: issueDate,
        due_date: dueDate,
        status,
        description,
      })
    } catch (err) {
      setErrors({ form: err instanceof Error ? err.message : 'Failed to save' })
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      {errors.form && <div className="login-form__error">{errors.form}</div>}
      <div className="form-field">
        <label>Vendor Name</label>
        <input type="text" value={vendorName} onChange={(e) => setVendorName(e.target.value)} autoFocus />
        {errors.vendor_name && <span className="field-error">{errors.vendor_name}</span>}
      </div>
      <div className="form-field">
        <label>Amount</label>
        <input type="number" value={amount} onChange={(e) => setAmount(e.target.value)} min="1" step="1" />
        {errors.amount && <span className="field-error">{errors.amount}</span>}
      </div>
      <div className="form-field">
        <label>Issue Date</label>
        <input type="date" value={issueDate} onChange={(e) => setIssueDate(e.target.value)} />
        {errors.issue_date && <span className="field-error">{errors.issue_date}</span>}
      </div>
      <div className="form-field">
        <label>Due Date</label>
        <input type="date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
        {errors.due_date && <span className="field-error">{errors.due_date}</span>}
      </div>
      <div className="form-field">
        <label>Status</label>
        <select value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="unpaid">Unpaid</option>
          <option value="paid">Paid</option>
          <option value="overdue">Overdue</option>
        </select>
      </div>
      <div className="form-field">
        <label>Description</label>
        <textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={3} placeholder="Optional" />
      </div>
      <div className="form-actions">
        <button type="button" className="btn btn-sm" onClick={onCancel}>Cancel</button>
        <button type="submit" className="btn btn-sm btn-primary" disabled={loading}>
          {loading ? 'Saving...' : invoice ? 'Update' : 'Add Invoice'}
        </button>
      </div>
    </form>
  )
}
