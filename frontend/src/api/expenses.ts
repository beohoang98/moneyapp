import { apiClient } from './client'
import type { ApiResponse } from '../types/api'
import type { Expense, CreateExpensePayload } from '../types/expense'

interface ExpenseListResponse {
  data: Expense[]
  total: number
  total_amount: number
  page: number
  per_page: number
}

interface ExpenseListParams {
  page?: number
  per_page?: number
  date_from?: string
  date_to?: string
  category_id?: number
  category_ids?: number[]
  sort_by?: string
  sort_order?: string
}

export async function getExpenses(params: ExpenseListParams = {}): Promise<ExpenseListResponse> {
  const searchParams = new URLSearchParams()
  if (params.page) searchParams.set('page', String(params.page))
  if (params.per_page) searchParams.set('per_page', String(params.per_page))
  if (params.date_from) searchParams.set('date_from', params.date_from)
  if (params.date_to) searchParams.set('date_to', params.date_to)
  if (params.category_id) searchParams.set('category_id', String(params.category_id))
  if (params.category_ids?.length) searchParams.set('category_ids', params.category_ids.join(','))
  if (params.sort_by) searchParams.set('sort_by', params.sort_by)
  if (params.sort_order) searchParams.set('sort_order', params.sort_order)

  const qs = searchParams.toString()
  return apiClient.get<ExpenseListResponse>(`/expenses${qs ? '?' + qs : ''}`)
}

export async function getExpense(id: number): Promise<Expense> {
  const res = await apiClient.get<ApiResponse<Expense>>(`/expenses/${id}`)
  return res.data
}

export async function createExpense(data: CreateExpensePayload): Promise<Expense> {
  const res = await apiClient.post<ApiResponse<Expense>>('/expenses', data)
  return res.data
}

export async function updateExpense(id: number, data: CreateExpensePayload): Promise<Expense> {
  const res = await apiClient.put<ApiResponse<Expense>>(`/expenses/${id}`, data)
  return res.data
}

export async function deleteExpense(id: number): Promise<void> {
  await apiClient.delete(`/expenses/${id}`)
}
