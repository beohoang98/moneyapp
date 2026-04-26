import { apiClient } from './client'
import type { ApiResponse } from '../types/api'
import type { Invoice, CreateInvoicePayload } from '../types/invoice'

interface InvoiceListResponse {
  data: Invoice[]
  total: number
  total_amount: number
  page: number
  per_page: number
}

interface InvoiceListParams {
  page?: number
  per_page?: number
  status?: string
  date_from?: string
  date_to?: string
  date_field?: string
}

interface InvoiceStats {
  total_outstanding: number
  unpaid_count: number
  unpaid_amount: number
  overdue_count: number
  overdue_amount: number
}

export async function getInvoices(params: InvoiceListParams = {}): Promise<InvoiceListResponse> {
  const searchParams = new URLSearchParams()
  if (params.page) searchParams.set('page', String(params.page))
  if (params.per_page) searchParams.set('per_page', String(params.per_page))
  if (params.status) searchParams.set('status', params.status)
  if (params.date_from) searchParams.set('date_from', params.date_from)
  if (params.date_to) searchParams.set('date_to', params.date_to)
  if (params.date_field) searchParams.set('date_field', params.date_field)

  const qs = searchParams.toString()
  return apiClient.get<InvoiceListResponse>(`/invoices${qs ? '?' + qs : ''}`)
}

export async function getInvoice(id: number): Promise<Invoice> {
  const res = await apiClient.get<ApiResponse<Invoice>>(`/invoices/${id}`)
  return res.data
}

export async function createInvoice(data: CreateInvoicePayload): Promise<Invoice> {
  const res = await apiClient.post<ApiResponse<Invoice>>('/invoices', data)
  return res.data
}

export async function updateInvoice(id: number, data: CreateInvoicePayload): Promise<Invoice> {
  const res = await apiClient.put<ApiResponse<Invoice>>(`/invoices/${id}`, data)
  return res.data
}

export async function deleteInvoice(id: number): Promise<void> {
  await apiClient.delete(`/invoices/${id}`)
}

export async function markInvoiceAsPaid(id: number): Promise<Invoice> {
  const res = await apiClient.patch<ApiResponse<Invoice>>(`/invoices/${id}/status`, { status: 'paid' })
  return res.data
}

export async function getInvoiceStats(): Promise<InvoiceStats> {
  const res = await apiClient.get<ApiResponse<InvoiceStats>>('/invoices/stats')
  return res.data
}
