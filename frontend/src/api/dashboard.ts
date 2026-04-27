import { apiClient } from './client'
import type { ApiResponse } from '../types/api'
import type { DashboardSummary, MonthlyTrendItem, CategoryBreakdownItem } from '../types/dashboard'

function buildDateParams(dateFrom?: string, dateTo?: string): string {
  const params = new URLSearchParams()
  if (dateFrom) params.set('date_from', dateFrom)
  if (dateTo) params.set('date_to', dateTo)
  const qs = params.toString()
  return qs ? '?' + qs : ''
}

export async function getDashboardSummary(dateFrom?: string, dateTo?: string): Promise<DashboardSummary> {
  const res = await apiClient.get<ApiResponse<DashboardSummary>>(`/dashboard/summary${buildDateParams(dateFrom, dateTo)}`)
  return res.data
}

export async function getMonthlyTrend(dateFrom?: string, dateTo?: string): Promise<MonthlyTrendItem[]> {
  const res = await apiClient.get<{ data: MonthlyTrendItem[] }>(`/dashboard/monthly-trend${buildDateParams(dateFrom, dateTo)}`)
  return res.data
}

export async function getExpenseByCategory(dateFrom?: string, dateTo?: string): Promise<CategoryBreakdownItem[]> {
  const res = await apiClient.get<{ data: CategoryBreakdownItem[] }>(`/dashboard/expense-by-category${buildDateParams(dateFrom, dateTo)}`)
  return res.data ?? []
}
