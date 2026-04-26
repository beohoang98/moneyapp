import { createContext } from 'react'

type ToastType = 'success' | 'error' | 'info'

export interface ToastContextValue {
  addToast: (message: string, type?: ToastType, duration?: number) => void
}

export const ToastContext = createContext<ToastContextValue | null>(null)
