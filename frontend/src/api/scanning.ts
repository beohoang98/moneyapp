import { apiClient, ApiClientError } from './client'
import type { ScanningSettingsResponse, ScanningSettingsUpdate, ScanResponse } from '../types/scanning'

export async function getScanningSettings(): Promise<ScanningSettingsResponse> {
  return apiClient.get<ScanningSettingsResponse>('/settings/scanning')
}

export async function updateScanningSettings(data: ScanningSettingsUpdate): Promise<ScanningSettingsResponse> {
  return apiClient.put<ScanningSettingsResponse>('/settings/scanning', data)
}

export async function testScanningConnection(data: {
  base_url: string
  model: string
  api_key?: string
}): Promise<{ ok: boolean; message: string }> {
  return apiClient.post<{ ok: boolean; message: string }>('/settings/scanning/test', data)
}

export async function getScanningHealth(): Promise<{ ok: boolean; message: string }> {
  return apiClient.get<{ ok: boolean; message: string }>('/scanning/health')
}

export async function scanInvoice(file: File, signal?: AbortSignal): Promise<ScanResponse> {
  const formData = new FormData()
  formData.append('image', file)

  const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api'
  const token = localStorage.getItem('moneyapp_token')
  const headers: Record<string, string> = {}
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${BASE_URL}/scanning/invoice`, {
    method: 'POST',
    headers,
    body: formData,
    signal,
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: 'Unknown error' }))
    throw new ApiClientError(res.status, body.error)
  }

  return res.json() as Promise<ScanResponse>
}

export async function deleteTempScan(storageKey: string): Promise<void> {
  const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api'
  const token = localStorage.getItem('moneyapp_token')
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`

  await fetch(`${BASE_URL}/scanning/temp`, {
    method: 'DELETE',
    headers,
    body: JSON.stringify({ storage_key: storageKey }),
  })
}
