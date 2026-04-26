import './Dashboard.css'

interface SummaryCardProps {
  title: string
  amount: number
  formatFn: (n: number) => string
  variant?: 'positive' | 'negative' | 'neutral'
}

export function SummaryCard({ title, amount, formatFn, variant = 'neutral' }: SummaryCardProps) {
  return (
    <div className={`summary-card summary-card--${variant}`}>
      <div className="summary-card__title">{title}</div>
      <div className="summary-card__amount">{formatFn(amount)}</div>
    </div>
  )
}
