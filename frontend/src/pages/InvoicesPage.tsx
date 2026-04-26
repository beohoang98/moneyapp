import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { getInvoices, createInvoice, updateInvoice, deleteInvoice, markInvoiceAsPaid, getInvoiceStats } from '../api/invoices'
import { InvoiceForm } from '../components/invoices/InvoiceForm'
import { DateRangeFilter } from '../components/filters/DateRangeFilter'
import { ConfirmDialog } from '../components/ConfirmDialog'
import { FileUpload } from '../components/attachments/FileUpload'
import { AttachmentList } from '../components/attachments/AttachmentList'
import { useToast } from '../hooks/useToast'
import { formatAmount, formatDisplayDate } from '../utils/format'
import type { Invoice } from '../types/invoice'
import './Pages.css'

const STATUS_TABS = [
  { value: '', label: 'All' },
  { value: 'unpaid', label: 'Unpaid' },
  { value: 'paid', label: 'Paid' },
  { value: 'overdue', label: 'Overdue' },
]

export function InvoicesPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(Number(searchParams.get('page')) || 1)
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editingInvoice, setEditingInvoice] = useState<Invoice | undefined>()
  const [deleteTarget, setDeleteTarget] = useState<Invoice | null>(null)
  const [payTarget, setPayTarget] = useState<Invoice | null>(null)
  const [detailTarget, setDetailTarget] = useState<Invoice | null>(null)
  const [attachmentRefresh, setAttachmentRefresh] = useState(0)
  const [stats, setStats] = useState<{ total_outstanding: number; unpaid_count: number; overdue_count: number } | null>(null)
  const { addToast } = useToast()

  const status = searchParams.get('status') ?? ''
  const dateFrom = searchParams.get('date_from') ?? ''
  const dateTo = searchParams.get('date_to') ?? ''
  const dateField = searchParams.get('date_field') ?? 'due_date'
  const perPage = 20

  const [refreshKey, setRefreshKey] = useState(0)

  useEffect(() => {
    let cancelled = false
    Promise.all([
      getInvoices({
        page,
        per_page: perPage,
        status: status || undefined,
        date_from: dateFrom || undefined,
        date_to: dateTo || undefined,
        date_field: dateField || undefined,
      }),
      getInvoiceStats(),
    ])
      .then(([result, invoiceStats]) => {
        if (cancelled) return
        setInvoices(result.data)
        setTotal(result.total)
        setStats(invoiceStats)
      })
      .catch(() => {
        if (!cancelled) addToast('Failed to load invoices', 'error')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => { cancelled = true }
  }, [page, status, dateFrom, dateTo, dateField, refreshKey, addToast])

  const refreshList = () => { setLoading(true); setRefreshKey((k) => k + 1) }

  const updateParams = (updates: Record<string, string>) => {
    const params = new URLSearchParams(searchParams)
    Object.entries(updates).forEach(([k, v]) => {
      if (v) params.set(k, v)
      else params.delete(k)
    })
    setSearchParams(params)
  }

  const handleStatusChange = (newStatus: string) => {
    setLoading(true); setPage(1)
    updateParams({ status: newStatus, page: '' })
  }

  const handleDateChange = (from: string, to: string) => {
    setLoading(true); setPage(1)
    updateParams({ date_from: from, date_to: to, page: '' })
  }

  const handleDateFieldChange = (field: string) => {
    setLoading(true); setPage(1)
    updateParams({ date_field: field, page: '' })
  }

  const handlePageChange = (newPage: number) => {
    setLoading(true); setPage(newPage)
    updateParams({ page: String(newPage) })
  }

  const handleCreate = async (data: { vendor_name: string; amount: number; issue_date: string; due_date: string; status: string; description: string }) => {
    await createInvoice(data)
    setShowForm(false)
    addToast('Invoice added', 'success')
    refreshList()
  }

  const handleUpdate = async (data: { vendor_name: string; amount: number; issue_date: string; due_date: string; status: string; description: string }) => {
    if (!editingInvoice) return
    await updateInvoice(editingInvoice.id, data)
    setEditingInvoice(undefined)
    setShowForm(false)
    addToast('Invoice updated', 'success')
    refreshList()
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    try {
      await deleteInvoice(deleteTarget.id)
      setDeleteTarget(null)
      addToast('Invoice deleted', 'success')
      refreshList()
    } catch {
      addToast('Failed to delete invoice', 'error')
    }
  }

  const handleMarkPaid = async () => {
    if (!payTarget) return
    try {
      await markInvoiceAsPaid(payTarget.id)
      setPayTarget(null)
      addToast('Invoice marked as paid', 'success')
      refreshList()
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Failed to update', 'error')
    }
  }

  const totalPages = Math.ceil(total / perPage)

  return (
    <div>
      <div className="page-header">
        <h1>Invoices</h1>
        <button className="btn btn-primary btn-sm" onClick={() => { setEditingInvoice(undefined); setShowForm(true) }}>
          + Add Invoice
        </button>
      </div>

      {stats && stats.total_outstanding > 0 && (
        <div className="summary-bar">
          <div className="summary-stat">
            <div className="summary-stat__label">Outstanding</div>
            <div className="summary-stat__value summary-stat__value--negative">
              {formatAmount(stats.total_outstanding)}
            </div>
          </div>
          <div className="summary-stat">
            <div className="summary-stat__label">Unpaid</div>
            <div className="summary-stat__value">{stats.unpaid_count}</div>
          </div>
          <div className="summary-stat">
            <div className="summary-stat__label">Overdue</div>
            <div className="summary-stat__value summary-stat__value--negative">{stats.overdue_count}</div>
          </div>
        </div>
      )}

      <div className="status-tabs">
        {STATUS_TABS.map((tab) => (
          <button
            key={tab.value}
            className={`status-tab ${status === tab.value ? 'status-tab--active' : ''}`}
            onClick={() => handleStatusChange(tab.value)}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="filter-bar">
        <div className="filter-field">
          <label>Filter by</label>
          <select value={dateField} onChange={(e) => handleDateFieldChange(e.target.value)}>
            <option value="due_date">Due Date</option>
            <option value="issue_date">Issue Date</option>
          </select>
        </div>
        <DateRangeFilter
          dateFrom={dateFrom}
          dateTo={dateTo}
          onChange={handleDateChange}
          onClear={() => handleDateChange('', '')}
        />
      </div>

      {loading ? (
        <div className="loading">Loading...</div>
      ) : invoices.length === 0 ? (
        <div className="empty-state"><p>No invoices found.</p></div>
      ) : (
        <>
          <table className="data-table">
            <thead>
              <tr>
                <th>Vendor</th>
                <th style={{ textAlign: 'right' }}>Amount</th>
                <th>Issue Date</th>
                <th>Due Date</th>
                <th>Status</th>
                <th style={{ textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {invoices.map((inv) => (
                <tr key={inv.id}>
                  <td>{inv.vendor_name}</td>
                  <td className="amount">{formatAmount(inv.amount)}</td>
                  <td>{formatDisplayDate(inv.issue_date)}</td>
                  <td>{formatDisplayDate(inv.due_date)}</td>
                  <td>
                    <span className={`status-badge status-badge--${inv.status}`}>
                      {inv.status}
                    </span>
                  </td>
                  <td className="actions">
                    <button className="btn btn-sm" onClick={() => setDetailTarget(inv)} title="Attachments">Files</button>
                    {(inv.status === 'unpaid' || inv.status === 'overdue') && (
                      <button className="btn btn-sm" onClick={() => setPayTarget(inv)} title="Mark as Paid">Paid</button>
                    )}
                    <button className="btn btn-sm btn-icon" onClick={() => { setEditingInvoice(inv); setShowForm(true) }} title="Edit">Edit</button>
                    <button className="btn btn-sm btn-icon" onClick={() => setDeleteTarget(inv)} title="Delete">Delete</button>
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
        <div className="form-modal-overlay" onClick={() => { setShowForm(false); setEditingInvoice(undefined) }}>
          <div className="form-modal" onClick={(e) => e.stopPropagation()}>
            <h2>{editingInvoice ? 'Edit Invoice' : 'Add Invoice'}</h2>
            <InvoiceForm
              key={editingInvoice?.id ?? 'new'}
              invoice={editingInvoice}
              onSubmit={editingInvoice ? handleUpdate : handleCreate}
              onCancel={() => { setShowForm(false); setEditingInvoice(undefined) }}
            />
          </div>
        </div>
      )}

      {detailTarget && (
        <div className="form-modal-overlay" onClick={() => setDetailTarget(null)}>
          <div className="form-modal" onClick={(e) => e.stopPropagation()}>
            <h2>Invoice Attachments</h2>
            <p style={{ fontSize: 14, color: 'var(--text)' }}>
              {detailTarget.vendor_name} &mdash; {formatAmount(detailTarget.amount)}
            </p>
            <FileUpload
              entityType="invoice"
              entityId={detailTarget.id}
              onUploaded={() => setAttachmentRefresh((k) => k + 1)}
            />
            <AttachmentList
              entityType="invoice"
              entityId={detailTarget.id}
              refreshKey={attachmentRefresh}
            />
          </div>
        </div>
      )}

      <ConfirmDialog open={!!deleteTarget} title="Delete Invoice" message="Are you sure you want to delete this invoice?" confirmLabel="Delete" onConfirm={handleDelete} onCancel={() => setDeleteTarget(null)} danger />
      <ConfirmDialog open={!!payTarget} title="Mark as Paid" message={payTarget ? `Mark the invoice from "${payTarget.vendor_name}" as paid?` : ''} confirmLabel="Mark as Paid" onConfirm={handleMarkPaid} onCancel={() => setPayTarget(null)} />
    </div>
  )
}
