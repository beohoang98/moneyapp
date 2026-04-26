import { apiClient } from './client'
import type { Category } from '../types/category'

interface CategoriesResponse {
  data: Category[]
}

export async function getCategories(type: 'expense' | 'income'): Promise<Category[]> {
  const res = await apiClient.get<CategoriesResponse>(`/categories?type=${type}`)
  return res.data
}
