import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { getIncomes, createIncome, updateIncome, deleteIncome } from '../api/incomes'
import { IncomeForm } from '../components/income/IncomeForm'
import { DateRangeFilter } from '../components/filters/DateRangeFilter'
import { CategoryFilter } from '../components/filters/CategoryFilter'
import { ConfirmDialog } from '../components/ConfirmDialog'
import { FileUpload } from '../components/attachments/FileUpload'
import { AttachmentList } from '../components/attachments/AttachmentList'
import { useToast } from '../hooks/useToast'
import { formatAmount, formatDisplayDate } from '../utils/format'
import type { Income } from '../types/income'
import './Pages.css'

export function IncomePage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [incomes, setIncomes] = useState<Income[]>([])
  const [total, setTotal] = useState(0)
  const [totalAmount, setTotalAmount] = useState(0)
  const [page, setPage] = useState(Number(searchParams.get('page')) || 1)
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editingIncome, setEditingIncome] = useState<Income | undefined>()
  const [deleteTarget, setDeleteTarget] = useState<Income | null>(null)
  const [detailTarget, setDetailTarget] = useState<Income | null>(null)
  const [attachmentRefresh, setAttachmentRefresh] = useState(0)
  const { addToast } = useToast()

  const dateFrom = searchParams.get('date_from') ?? ''
  const dateTo = searchParams.get('date_to') ?? ''
  const categoryId = searchParams.get('category_id') ? Number(searchParams.get('category_id')) : undefined
  const perPage = 20

  const [refreshKey, setRefreshKey] = useState(0)

  useEffect(() => {
    let cancelled = false
    getIncomes({ page, per_page: perPage, date_from: dateFrom || undefined, date_to: dateTo || undefined, category_id: categoryId })
      .then((result) => {
        if (cancelled) return
        setIncomes(result.data)
        setTotal(result.total)
        setTotalAmount(result.total_amount)
      })
      .catch(() => {
        if (!cancelled) addToast('Failed to load income records', 'error')
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
    await createIncome(data)
    setShowForm(false)
    addToast('Income added', 'success')
    refreshList()
  }

  const handleUpdate = async (data: { amount: number; date: string; category_id: number; description: string }) => {
    if (!editingIncome) return
    await updateIncome(editingIncome.id, data)
    setEditingIncome(undefined)
    setShowForm(false)
    addToast('Income updated', 'success')
    refreshList()
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    try {
      await deleteIncome(deleteTarget.id)
      setDeleteTarget(null)
      addToast('Income deleted', 'success')
      refreshList()
    } catch {
      addToast('Failed to delete income', 'error')
    }
  }

  const totalPages = Math.ceil(total / perPage)

  return (
    <div>
      <div className="page-header">
        <h1>Income</h1>
        <button className="btn btn-primary btn-sm" onClick={() => { setEditingIncome(undefined); setShowForm(true) }}>
          + Add Income
        </button>
      </div>

      <div className="filter-bar">
        <DateRangeFilter dateFrom={dateFrom} dateTo={dateTo} onChange={handleDateChange} onClear={() => handleDateChange('', '')} />
        <CategoryFilter type="income" value={categoryId} onChange={handleCategoryChange} />
      </div>

      <div className="summary-bar">
        <div className="summary-stat">
          <div className="summary-stat__label">Total</div>
          <div className="summary-stat__value summary-stat__value--positive">{formatAmount(totalAmount)}</div>
        </div>
      </div>

      {loading ? (
        <div className="loading">Loading...</div>
      ) : incomes.length === 0 ? (
        <div className="empty-state"><p>No income records yet.</p></div>
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
              {incomes.map((inc) => (
                <tr key={inc.id}>
                  <td>{formatDisplayDate(inc.date)}</td>
                  <td>{inc.category_name}</td>
                  <td>{inc.description}</td>
                  <td className="amount">{formatAmount(inc.amount)}</td>
                  <td className="actions">
                    <button className="btn btn-sm" onClick={() => setDetailTarget(inc)} title="Attachments">Files</button>
                    <button className="btn btn-sm btn-icon" onClick={() => { setEditingIncome(inc); setShowForm(true) }} title="Edit">Edit</button>
                    <button className="btn btn-sm btn-icon" onClick={() => setDeleteTarget(inc)} title="Delete">Delete</button>
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
        <div className="form-modal-overlay" onClick={() => { setShowForm(false); setEditingIncome(undefined) }}>
          <div className="form-modal" onClick={(e) => e.stopPropagation()}>
            <h2>{editingIncome ? 'Edit Income' : 'Add Income'}</h2>
            <IncomeForm
              key={editingIncome?.id ?? 'new'}
              income={editingIncome}
              onSubmit={editingIncome ? handleUpdate : handleCreate}
              onCancel={() => { setShowForm(false); setEditingIncome(undefined) }}
            />
          </div>
        </div>
      )}

      {detailTarget && (
        <div className="form-modal-overlay" onClick={() => setDetailTarget(null)}>
          <div className="form-modal" onClick={(e) => e.stopPropagation()}>
            <h2>Income Attachments</h2>
            <p style={{ fontSize: 14, color: 'var(--text)' }}>
              {formatDisplayDate(detailTarget.date)} &mdash; {formatAmount(detailTarget.amount)}
            </p>
            <FileUpload
              entityType="income"
              entityId={detailTarget.id}
              onUploaded={() => setAttachmentRefresh((k) => k + 1)}
            />
            <AttachmentList
              entityType="income"
              entityId={detailTarget.id}
              refreshKey={attachmentRefresh}
            />
          </div>
        </div>
      )}

      <ConfirmDialog open={!!deleteTarget} title="Delete Income" message="Are you sure you want to delete this income record?" confirmLabel="Delete" onConfirm={handleDelete} onCancel={() => setDeleteTarget(null)} danger />
    </div>
  )
}
