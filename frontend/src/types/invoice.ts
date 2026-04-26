export interface Invoice {
  id: number
  vendor_name: string
  amount: number
  issue_date: string
  due_date: string
  status: 'unpaid' | 'paid' | 'overdue'
  description: string
  created_at: string
  updated_at: string
}

export interface CreateInvoicePayload {
  vendor_name: string
  amount: number
  issue_date: string
  due_date: string
  status: string
  description: string
}
