import { useEffect, useState } from 'react'
import { getDashboardSummary } from '../api/dashboard'
import { SummaryCard } from '../components/dashboard/SummaryCard'
import { InvoiceSummaryCard } from '../components/dashboard/InvoiceSummaryCard'
import { formatAmount } from '../utils/format'
import type { DashboardSummary } from '../types/dashboard'
import './Pages.css'

export function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getDashboardSummary()
      .then(setSummary)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div>
        <h1>Dashboard</h1>
        <div className="dashboard-skeleton">
          <div className="skeleton-card" />
          <div className="skeleton-card" />
          <div className="skeleton-card" />
        </div>
      </div>
    )
  }

  if (!summary) {
    return (
      <div>
        <h1>Dashboard</h1>
        <p>Failed to load dashboard data.</p>
      </div>
    )
  }

  return (
    <div>
      <h1>Dashboard</h1>
      <p style={{ color: 'var(--text)', marginBottom: 20 }}>
        {summary.date_from} to {summary.date_to}
      </p>

      <div className="dashboard-cards">
        <SummaryCard
          title="Total Income"
          amount={summary.total_income}
          formatFn={formatAmount}
          variant="positive"
        />
        <SummaryCard
          title="Total Expenses"
          amount={summary.total_expenses}
          formatFn={formatAmount}
          variant="negative"
        />
        <SummaryCard
          title="Net Balance"
          amount={summary.net_balance}
          formatFn={formatAmount}
          variant={summary.net_balance >= 0 ? 'positive' : 'negative'}
        />
      </div>

      <InvoiceSummaryCard
        unpaidCount={summary.unpaid_count}
        unpaidAmount={summary.unpaid_amount}
        overdueCount={summary.overdue_count}
        overdueAmount={summary.overdue_amount}
        totalOutstanding={summary.total_outstanding}
      />
    </div>
  )
}
