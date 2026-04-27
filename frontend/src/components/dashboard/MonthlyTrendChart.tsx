import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'
import { formatAmount } from '../../utils/format'
import type { MonthlyTrendItem } from '../../types/dashboard'

interface Props {
  data: MonthlyTrendItem[]
}

export function MonthlyTrendChart({ data }: Props) {
  if (data.length === 0) {
    return (
      <div className="chart-empty">
        <p>No data for this period</p>
      </div>
    )
  }

  return (
    <div className="chart-container">
      <h3 className="chart-title">Income vs. Expenses</h3>
      <ResponsiveContainer width="100%" height={300}>
        <BarChart data={data} margin={{ top: 5, right: 20, bottom: 5, left: 20 }}>
          <XAxis dataKey="month" tick={{ fontSize: 12 }} />
          <YAxis tickFormatter={(v: number) => formatAmount(v)} tick={{ fontSize: 11 }} width={100} />
          <Tooltip
            formatter={(value, name) => [
              formatAmount(Number(value)),
              String(name) === 'total_income' ? 'Income' : 'Expenses',
            ]}
            labelFormatter={(label) => `Month: ${label}`}
          />
          <Legend formatter={(value: string) => (value === 'total_income' ? 'Income' : 'Expenses')} />
          <Bar dataKey="total_income" fill="#22c55e" name="total_income" radius={[4, 4, 0, 0]} />
          <Bar dataKey="total_expenses" fill="#ef4444" name="total_expenses" radius={[4, 4, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}
