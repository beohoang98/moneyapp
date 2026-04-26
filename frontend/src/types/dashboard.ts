export interface DashboardSummary {
  total_income: number
  total_expenses: number
  net_balance: number
  date_from: string
  date_to: string
  unpaid_count: number
  unpaid_amount: number
  overdue_count: number
  overdue_amount: number
  total_outstanding: number
}
