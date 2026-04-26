export interface Income {
  id: number
  amount: number
  date: string
  category_id: number
  category_name?: string
  description: string
  created_at: string
  updated_at: string
}

export interface CreateIncomePayload {
  amount: number
  date: string
  category_id: number
  description: string
}
