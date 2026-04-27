import { useEffect, useState, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { getDashboardSummary, getMonthlyTrend, getExpenseByCategory } from '../api/dashboard'
import { SummaryCard } from '../components/dashboard/SummaryCard'
import { InvoiceSummaryCard } from '../components/dashboard/InvoiceSummaryCard'
import { PeriodSelector } from '../components/dashboard/PeriodSelector'
import { MonthlyTrendChart } from '../components/dashboard/MonthlyTrendChart'
import { CategoryBreakdownChart } from '../components/dashboard/CategoryBreakdownChart'
import { formatAmount } from '../utils/format'
import type { DashboardSummary, MonthlyTrendItem, CategoryBreakdownItem } from '../types/dashboard'
import './Pages.css'

function defaultDateRange(): { from: string; to: string } {
  const now = new Date()
  const y = now.getFullYear()
  const m = now.getMonth()
  const fmt = (d: Date) => {
    const yy = d.getFullYear()
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    return `${yy}-${mm}-${dd}`
  }
  const from = fmt(new Date(y, m, 1))
  const to = fmt(new Date(y, m + 1, 0))
  return { from, to }
}

export function DashboardPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [trend, setTrend] = useState<MonthlyTrendItem[]>([])
  const [categoryBreakdown, setCategoryBreakdown] = useState<CategoryBreakdownItem[]>([])
  const [loading, setLoading] = useState(true)

  const defaults = defaultDateRange()
  const dateFrom = searchParams.get('date_from') || defaults.from
  const dateTo = searchParams.get('date_to') || defaults.to

  const handlePeriodChange = useCallback(
    (from: string, to: string) => {
      const params = new URLSearchParams(searchParams)
      if (from) params.set('date_from', from)
      else params.delete('date_from')
      if (to) params.set('date_to', to)
      else params.delete('date_to')
      setSearchParams(params)
    },
    [searchParams, setSearchParams],
  )

  useEffect(() => {
    let cancelled = false
    const fetchAll = async () => {
      try {
        const [s, t, c] = await Promise.all([
          getDashboardSummary(dateFrom, dateTo),
          getMonthlyTrend(dateFrom, dateTo),
          getExpenseByCategory(dateFrom, dateTo),
        ])
        if (!cancelled) {
          setSummary(s)
          setTrend(t)
          setCategoryBreakdown(c)
        }
      } catch {
        // ignore
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    fetchAll()
    return () => {
      cancelled = true
    }
  }, [dateFrom, dateTo])

  if (loading) {
    return (
      <div>
        <h1>Dashboard</h1>
        <PeriodSelector dateFrom={dateFrom} dateTo={dateTo} onChange={handlePeriodChange} />
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
        <PeriodSelector dateFrom={dateFrom} dateTo={dateTo} onChange={handlePeriodChange} />
        <p>Failed to load dashboard data.</p>
      </div>
    )
  }

  return (
    <div>
      <h1>Dashboard</h1>

      <PeriodSelector dateFrom={dateFrom} dateTo={dateTo} onChange={handlePeriodChange} />

      <p style={{ color: 'var(--text)', marginBottom: 20 }}>
        {summary.date_from} to {summary.date_to}
      </p>

      <div className="dashboard-cards">
        <SummaryCard title="Total Income" amount={summary.total_income} formatFn={formatAmount} variant="positive" />
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

      <div className="dashboard-charts">
        <MonthlyTrendChart data={trend} />
        <CategoryBreakdownChart data={categoryBreakdown} />
      </div>
    </div>
  )
}
