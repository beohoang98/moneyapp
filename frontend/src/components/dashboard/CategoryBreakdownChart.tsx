import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'
import { formatAmount } from '../../utils/format'
import type { CategoryBreakdownItem } from '../../types/dashboard'

interface Props {
  data: CategoryBreakdownItem[]
}

const COLORS = [
  '#3b82f6', '#ef4444', '#22c55e', '#f59e0b', '#8b5cf6',
  '#ec4899', '#14b8a6', '#f97316', '#6366f1', '#06b6d4',
]

export function CategoryBreakdownChart({ data }: Props) {
  if (data.length === 0) {
    return (
      <div className="chart-empty">
        <p>No expense data for this period</p>
      </div>
    )
  }

  const total = data.reduce((sum, d) => sum + d.total, 0)

  return (
    <div className="chart-container">
      <h3 className="chart-title">Expenses by Category</h3>
      <ResponsiveContainer width="100%" height={300}>
        <PieChart>
          <Pie
            data={data}
            dataKey="total"
            nameKey="category_name"
            cx="50%"
            cy="50%"
            innerRadius={60}
            outerRadius={100}
            paddingAngle={2}
          >
            {data.map((_, index) => (
              <Cell key={index} fill={COLORS[index % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip
            formatter={(value, name) => {
              const v = Number(value)
              const pct = total > 0 ? ((v / total) * 100).toFixed(1) : '0'
              return [`${formatAmount(v)} (${pct}%)`, String(name)]
            }}
          />
          <Legend
            formatter={(value: string) => {
              const item = data.find((d) => d.category_name === value)
              if (!item) return value
              const pct = total > 0 ? ((item.total / total) * 100).toFixed(1) : '0'
              return `${value}: ${formatAmount(item.total)} (${pct}%)`
            }}
          />
        </PieChart>
      </ResponsiveContainer>
    </div>
  )
}
