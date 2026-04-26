import { useCallback, useSyncExternalStore } from 'react'
import { TOKEN_KEY } from '../api/client'
import { login as apiLogin, logout as apiLogout } from '../api/auth'

function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

function subscribe(callback: () => void): () => void {
  const handler = (e: StorageEvent) => {
    if (e.key === TOKEN_KEY) callback()
  }
  window.addEventListener('storage', handler)
  return () => window.removeEventListener('storage', handler)
}

export function useAuth() {
  const token = useSyncExternalStore(subscribe, getToken, () => null)
  const isAuthenticated = token !== null

  const login = useCallback(async (username: string, password: string) => {
    await apiLogin(username, password)
  }, [])

  const logout = useCallback(() => {
    apiLogout()
  }, [])

  return { isAuthenticated, token, login, logout }
}
