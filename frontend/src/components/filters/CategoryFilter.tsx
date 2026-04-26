import { useEffect, useState } from 'react'
import { getCategories } from '../../api/categories'
import type { Category } from '../../types/category'
import './Filters.css'

interface CategoryFilterProps {
  type: 'expense' | 'income'
  value: number | undefined
  onChange: (categoryId: number | undefined) => void
}

export function CategoryFilter({ type, value, onChange }: CategoryFilterProps) {
  const [categories, setCategories] = useState<Category[]>([])

  useEffect(() => {
    getCategories(type).then(setCategories).catch(() => {})
  }, [type])

  return (
    <div className="filter-field">
      <label>Category</label>
      <select
        value={value ?? ''}
        onChange={(e) => onChange(e.target.value ? Number(e.target.value) : undefined)}
      >
        <option value="">All Categories</option>
        {categories.map((c) => (
          <option key={c.id} value={c.id}>
            {c.name}
          </option>
        ))}
      </select>
    </div>
  )
}
