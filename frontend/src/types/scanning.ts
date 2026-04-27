export interface ScanningSettingsResponse {
  enabled: boolean
  base_url: string
  model: string
  api_key_set: boolean
}

export interface ScanningSettingsUpdate {
  enabled: boolean
  base_url: string
  model: string
  api_key?: string
}

export interface ScanResponse {
  scan_result: ScanResult
  temp_storage_key: string
}

export interface ScanResult {
  vendor: string
  date: string
  currency: string
  total_amount: number
  line_items: LineItem[]
  confidence: Record<string, 'low' | 'medium' | 'high'>
}

export interface LineItem {
  description: string
  amount: number
}

export interface ScanError {
  error: string
  detail?: string
}
