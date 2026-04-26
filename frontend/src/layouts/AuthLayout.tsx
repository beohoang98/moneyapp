import { Outlet } from 'react-router-dom'
import './AuthLayout.css'

export function AuthLayout() {
  return (
    <div className="auth-layout">
      <div className="auth-layout__card">
        <h1 className="auth-layout__brand">MoneyApp</h1>
        <Outlet />
      </div>
    </div>
  )
}
