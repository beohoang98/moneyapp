export interface ApiResponse<T> {
  data: T
}

export interface ApiListResponse<T> {
  data: T[]
  total: number
  page: number
  per_page: number
}

export interface ApiError {
  error: string
}
