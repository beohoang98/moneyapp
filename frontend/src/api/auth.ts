import { apiClient, TOKEN_KEY } from './client'

interface LoginResponse {
  token: string
}

export async function login(
  username: string,
  password: string,
): Promise<string> {
  const res = await apiClient.post<LoginResponse>('/auth/login', {
    username,
    password,
  })
  localStorage.setItem(TOKEN_KEY, res.token)
  return res.token
}

export function logout(): void {
  localStorage.removeItem(TOKEN_KEY)
  window.location.href = '/login'
}
