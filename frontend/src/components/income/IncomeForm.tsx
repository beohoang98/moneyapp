import { useState, useEffect, type FormEvent } from 'react'
import { getCategories } from '../../api/categories'
import type { Category } from '../../types/category'
import type { Income } from '../../types/income'
import { todayString, toDisplayAmount, toInputDate, toMinorUnits } from '../../utils/format'

interface IncomeFormProps {
  income?: Income
  onSubmit: (data: { amount: number; date: string; category_id: number; description: string }) => Promise<void>
  onCancel: () => void
}

export function IncomeForm({ income, onSubmit, onCancel }: IncomeFormProps) {
  const [amount, setAmount] = useState(income ? String(toDisplayAmount(income.amount)) : '')
  const [date, setDate] = useState(
    () => (income ? toInputDate(income.date) || todayString() : todayString()),
  )
  const [categoryId, setCategoryId] = useState(income?.category_id ?? 0)
  const [description, setDescription] = useState(income?.description ?? '')
  const [categories, setCategories] = useState<Category[]>([])
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    getCategories('income').then((cats) => {
      setCategories(cats)
      if (!income && cats.length > 0 && categoryId === 0) {
        setCategoryId(cats[0].id)
      }
    })
  }, [])  // eslint-disable-line react-hooks/exhaustive-deps

  const validate = () => {
    const errs: Record<string, string> = {}
    const numAmount = Number(amount)
    if (!amount || numAmount <= 0) errs.amount = 'Amount must be greater than zero'
    if (!date) errs.date = 'Date is required'
    if (!categoryId) errs.category_id = 'Category is required'
    setErrors(errs)
    return Object.keys(errs).length === 0
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!validate()) return
    setLoading(true)
    try {
      await onSubmit({
        amount: toMinorUnits(Number(amount)),
        date,
        category_id: categoryId,
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
        <label>Amount</label>
        <input type="number" value={amount} onChange={(e) => setAmount(e.target.value)} min="1" step="1" autoFocus />
        {errors.amount && <span className="field-error">{errors.amount}</span>}
      </div>
      <div className="form-field">
        <label>Date</label>
        <input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
        {errors.date && <span className="field-error">{errors.date}</span>}
      </div>
      <div className="form-field">
        <label>Category</label>
        <select value={categoryId} onChange={(e) => setCategoryId(Number(e.target.value))}>
          <option value={0} disabled>Select category</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>{c.name}</option>
          ))}
        </select>
        {errors.category_id && <span className="field-error">{errors.category_id}</span>}
      </div>
      <div className="form-field">
        <label>Description</label>
        <textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={3} placeholder="Optional" />
      </div>
      <div className="form-actions">
        <button type="button" className="btn btn-sm" onClick={onCancel}>Cancel</button>
        <button type="submit" className="btn btn-sm btn-primary" disabled={loading}>
          {loading ? 'Saving...' : income ? 'Update' : 'Add Income'}
        </button>
      </div>
    </form>
  )
}
