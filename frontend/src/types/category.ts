export interface Category {
  id: number
  name: string
  type: 'expense' | 'income'
  is_default: boolean
  color?: string
}
