import { Link } from 'react-router-dom'
import { formatAmount } from '../../utils/format'
import './Dashboard.css'

interface InvoiceSummaryCardProps {
  unpaidCount: number
  unpaidAmount: number
  overdueCount: number
  overdueAmount: number
  totalOutstanding: number
}

export function InvoiceSummaryCard({ unpaidCount, overdueCount, totalOutstanding }: InvoiceSummaryCardProps) {
  const totalCount = unpaidCount + overdueCount

  if (totalCount === 0) {
    return (
      <div className="invoice-summary-card">
        <div className="invoice-summary-card__title">Outstanding Invoices</div>
        <p className="invoice-summary-card__empty">No outstanding invoices</p>
      </div>
    )
  }

  return (
    <div className="invoice-summary-card">
      <div className="invoice-summary-card__title">Outstanding Invoices</div>
      <div className="invoice-summary-card__total">{formatAmount(totalOutstanding)}</div>
      <div className="invoice-summary-card__details">
        {totalCount} outstanding ({unpaidCount} unpaid, {overdueCount} overdue)
      </div>
      <Link to="/invoices?status=unpaid" className="invoice-summary-card__link">
        View invoices →
      </Link>
    </div>
  )
}
