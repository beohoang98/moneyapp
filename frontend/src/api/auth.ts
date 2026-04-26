import { apiClient, TOKEN_KEY } from './client'
import type { ApiResponse } from '../types/api'

interface LoginResult {
  token: string
  expires_at: string
}

export async function login(
  username: string,
  password: string,
): Promise<string> {
  const res = await apiClient.post<ApiResponse<LoginResult>>('/auth/login', {
    username,
    password,
  })
  localStorage.setItem(TOKEN_KEY, res.data.token)
  return res.data.token
}

export function logout(): void {
  localStorage.removeItem(TOKEN_KEY)
  window.location.href = '/login'
}
