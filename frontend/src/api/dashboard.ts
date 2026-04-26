import { apiClient } from './client'
import type { ApiResponse } from '../types/api'
import type { DashboardSummary } from '../types/dashboard'

export async function getDashboardSummary(dateFrom?: string, dateTo?: string): Promise<DashboardSummary> {
  const params = new URLSearchParams()
  if (dateFrom) params.set('date_from', dateFrom)
  if (dateTo) params.set('date_to', dateTo)
  const qs = params.toString()
  const res = await apiClient.get<ApiResponse<DashboardSummary>>(`/dashboard/summary${qs ? '?' + qs : ''}`)
  return res.data
}
