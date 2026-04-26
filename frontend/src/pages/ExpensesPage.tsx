import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { getExpenses, createExpense, updateExpense, deleteExpense } from '../api/expenses'
import { ExpenseForm } from '../components/expenses/ExpenseForm'
import { DateRangeFilter } from '../components/filters/DateRangeFilter'
import { CategoryFilter } from '../components/filters/CategoryFilter'
import { ConfirmDialog } from '../components/ConfirmDialog'
import { useToast } from '../hooks/useToast'
import { formatAmount, formatDisplayDate } from '../utils/format'
import type { Expense } from '../types/expense'
import './Pages.css'

export function ExpensesPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [total, setTotal] = useState(0)
  const [totalAmount, setTotalAmount] = useState(0)
  const [page, setPage] = useState(Number(searchParams.get('page')) || 1)
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editingExpense, setEditingExpense] = useState<Expense | undefined>()
  const [deleteTarget, setDeleteTarget] = useState<Expense | null>(null)
  const { addToast } = useToast()

  const dateFrom = searchParams.get('date_from') ?? ''
  const dateTo = searchParams.get('date_to') ?? ''
  const categoryId = searchParams.get('category_id') ? Number(searchParams.get('category_id')) : undefined
  const perPage = 20

  const [refreshKey, setRefreshKey] = useState(0)

  useEffect(() => {
    let cancelled = false
    getExpenses({ page, per_page: perPage, date_from: dateFrom || undefined, date_to: dateTo || undefined, category_id: categoryId })
      .then((result) => {
        if (cancelled) return
        setExpenses(result.data)
        setTotal(result.total)
        setTotalAmount(result.total_amount)
      })
      .catch(() => {
        if (!cancelled) addToast('Failed to load expenses', 'error')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => { cancelled = true }
  }, [page, dateFrom, dateTo, categoryId, refreshKey, addToast])

  const refreshList = () => { setLoading(true); setRefreshKey((k) => k + 1) }

  const updateParams = (updates: Record<string, string>) => {
    const params = new URLSearchParams(searchParams)
    Object.entries(updates).forEach(([k, v]) => {
      if (v) params.set(k, v)
      else params.delete(k)
    })
    setSearchParams(params)
  }

  const handleDateChange = (from: string, to: string) => {
    setLoading(true); setPage(1)
    updateParams({ date_from: from, date_to: to, page: '' })
  }

  const handleCategoryChange = (catId: number | undefined) => {
    setLoading(true); setPage(1)
    updateParams({ category_id: catId ? String(catId) : '', page: '' })
  }

  const handlePageChange = (newPage: number) => {
    setLoading(true); setPage(newPage)
    updateParams({ page: String(newPage) })
  }

  const handleCreate = async (data: { amount: number; date: string; category_id: number; description: string }) => {
    await createExpense(data)
    setShowForm(false)
    addToast('Expense added', 'success')
    refreshList()
  }

  const handleUpdate = async (data: { amount: number; date: string; category_id: number; description: string }) => {
    if (!editingExpense) return
    await updateExpense(editingExpense.id, data)
    setEditingExpense(undefined)
    setShowForm(false)
    addToast('Expense updated', 'success')
    refreshList()
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    try {
      await deleteExpense(deleteTarget.id)
      setDeleteTarget(null)
      addToast('Expense deleted', 'success')
      refreshList()
    } catch {
      addToast('Failed to delete expense', 'error')
    }
  }

  const totalPages = Math.ceil(total / perPage)

  return (
    <div>
      <div className="page-header">
        <h1>Expenses</h1>
        <button className="btn btn-primary btn-sm" onClick={() => { setEditingExpense(undefined); setShowForm(true) }}>
          + Add Expense
        </button>
      </div>

      <div className="filter-bar">
        <DateRangeFilter
          dateFrom={dateFrom}
          dateTo={dateTo}
          onChange={handleDateChange}
          onClear={() => handleDateChange('', '')}
        />
        <CategoryFilter type="expense" value={categoryId} onChange={handleCategoryChange} />
      </div>

      {totalAmount > 0 && (
        <div className="summary-bar">
          <div className="summary-stat">
            <div className="summary-stat__label">Total</div>
            <div className="summary-stat__value">{formatAmount(totalAmount)}</div>
          </div>
        </div>
      )}

      {loading ? (
        <div className="loading">Loading...</div>
      ) : expenses.length === 0 ? (
        <div className="empty-state"><p>No expenses recorded yet.</p></div>
      ) : (
        <>
          <table className="data-table">
            <thead>
              <tr>
                <th>Date</th>
                <th>Category</th>
                <th>Description</th>
                <th style={{ textAlign: 'right' }}>Amount</th>
                <th style={{ textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {expenses.map((e) => (
                <tr key={e.id}>
                  <td>{formatDisplayDate(e.date)}</td>
                  <td>{e.category_name}</td>
                  <td>{e.description}</td>
                  <td className="amount">{formatAmount(e.amount)}</td>
                  <td className="actions">
                    <button className="btn btn-sm btn-icon" onClick={() => { setEditingExpense(e); setShowForm(true) }} title="Edit">✏️</button>
                    <button className="btn btn-sm btn-icon" onClick={() => setDeleteTarget(e)} title="Delete">🗑️</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {totalPages > 1 && (
            <div className="pagination">
              <button className="btn btn-sm" disabled={page <= 1} onClick={() => handlePageChange(page - 1)}>Previous</button>
              <span className="pagination__info">Page {page} of {totalPages}</span>
              <button className="btn btn-sm" disabled={page >= totalPages} onClick={() => handlePageChange(page + 1)}>Next</button>
            </div>
          )}
        </>
      )}

      {showForm && (
        <div className="form-modal-overlay" onClick={() => { setShowForm(false); setEditingExpense(undefined) }}>
          <div className="form-modal" onClick={(e) => e.stopPropagation()}>
            <h2>{editingExpense ? 'Edit Expense' : 'Add Expense'}</h2>
            <ExpenseForm
              key={editingExpense?.id ?? 'new'}
              expense={editingExpense}
              onSubmit={editingExpense ? handleUpdate : handleCreate}
              onCancel={() => { setShowForm(false); setEditingExpense(undefined) }}
            />
          </div>
        </div>
      )}

      <ConfirmDialog
        open={!!deleteTarget}
        title="Delete Expense"
        message={`Are you sure you want to delete this expense?`}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        onCancel={() => setDeleteTarget(null)}
        danger
      />
    </div>
  )
}
