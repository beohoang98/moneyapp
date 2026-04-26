import { apiClient } from './client'
import type { ApiResponse } from '../types/api'
import type { Income, CreateIncomePayload } from '../types/income'

interface IncomeListResponse {
  data: Income[]
  total: number
  total_amount: number
  page: number
  per_page: number
}

interface IncomeListParams {
  page?: number
  per_page?: number
  date_from?: string
  date_to?: string
  category_id?: number
  category_ids?: number[]
}

export async function getIncomes(params: IncomeListParams = {}): Promise<IncomeListResponse> {
  const searchParams = new URLSearchParams()
  if (params.page) searchParams.set('page', String(params.page))
  if (params.per_page) searchParams.set('per_page', String(params.per_page))
  if (params.date_from) searchParams.set('date_from', params.date_from)
  if (params.date_to) searchParams.set('date_to', params.date_to)
  if (params.category_id) searchParams.set('category_id', String(params.category_id))
  if (params.category_ids?.length) searchParams.set('category_ids', params.category_ids.join(','))

  const qs = searchParams.toString()
  return apiClient.get<IncomeListResponse>(`/incomes${qs ? '?' + qs : ''}`)
}

export async function getIncome(id: number): Promise<Income> {
  const res = await apiClient.get<ApiResponse<Income>>(`/incomes/${id}`)
  return res.data
}

export async function createIncome(data: CreateIncomePayload): Promise<Income> {
  const res = await apiClient.post<ApiResponse<Income>>('/incomes', data)
  return res.data
}

export async function updateIncome(id: number, data: CreateIncomePayload): Promise<Income> {
  const res = await apiClient.put<ApiResponse<Income>>(`/incomes/${id}`, data)
  return res.data
}

export async function deleteIncome(id: number): Promise<void> {
  await apiClient.delete(`/incomes/${id}`)
}
