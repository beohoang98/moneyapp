import './Filters.css'

interface DateRangeFilterProps {
  dateFrom: string
  dateTo: string
  onChange: (dateFrom: string, dateTo: string) => void
  onClear: () => void
}

export function DateRangeFilter({ dateFrom, dateTo, onChange, onClear }: DateRangeFilterProps) {
  return (
    <div className="filter-group">
      <div className="filter-field">
        <label>From</label>
        <input
          type="date"
          value={dateFrom}
          onChange={(e) => onChange(e.target.value, dateTo)}
        />
      </div>
      <div className="filter-field">
        <label>To</label>
        <input
          type="date"
          value={dateTo}
          onChange={(e) => onChange(dateFrom, e.target.value)}
        />
      </div>
      {(dateFrom || dateTo) && (
        <button className="btn btn-sm filter-clear" onClick={onClear}>
          Clear
        </button>
      )}
    </div>
  )
}
