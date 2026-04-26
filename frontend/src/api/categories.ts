import { apiClient } from './client'
import type { ApiResponse } from '../types/api'
import type { Category } from '../types/category'

interface CategoriesResponse {
  data: Category[]
}

export async function getCategories(type: 'expense' | 'income'): Promise<Category[]> {
  const res = await apiClient.get<CategoriesResponse>(`/categories?type=${type}`)
  return res.data
}

export async function getAllCategories(): Promise<Category[]> {
  const res = await apiClient.get<CategoriesResponse>('/categories')
  return res.data
}

export async function createCategory(name: string, type: 'expense' | 'income'): Promise<Category> {
  const res = await apiClient.post<ApiResponse<Category>>('/categories', { name, type })
  return res.data
}

export async function updateCategory(id: number, name: string): Promise<Category> {
  const res = await apiClient.put<ApiResponse<Category>>(`/categories/${id}`, { name })
  return res.data
}

export async function deleteCategory(id: number): Promise<void> {
  await apiClient.delete(`/categories/${id}`)
}
