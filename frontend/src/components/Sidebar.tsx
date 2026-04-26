import { useState } from 'react'
import { NavLink } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import './Sidebar.css'

const navItems = [
  { to: '/dashboard', label: 'Dashboard', icon: '📊' },
  { to: '/expenses', label: 'Expenses', icon: '💸' },
  { to: '/income', label: 'Income', icon: '💰' },
  { to: '/invoices', label: 'Invoices', icon: '📄' },
  { to: '/categories', label: 'Categories', icon: '🏷️' },
  { to: '/settings', label: 'Settings', icon: '⚙️' },
]

export function Sidebar() {
  const [mobileOpen, setMobileOpen] = useState(false)
  const { logout } = useAuth()

  return (
    <div className="sidebar-root">
      <button
        className="sidebar-toggle"
        onClick={() => setMobileOpen(!mobileOpen)}
        aria-label="Toggle navigation"
      >
        ☰
      </button>
      <aside className={`sidebar ${mobileOpen ? 'sidebar--open' : ''}`}>
        <div className="sidebar__brand">MoneyApp</div>
        <nav className="sidebar__nav">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `sidebar__link ${isActive ? 'sidebar__link--active' : ''}`
              }
              onClick={() => setMobileOpen(false)}
            >
              <span className="sidebar__icon">{item.icon}</span>
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>
        <div className="sidebar__footer">
          <button className="sidebar__logout" onClick={logout}>
            Logout
          </button>
        </div>
      </aside>
      {mobileOpen && (
        <div className="sidebar-overlay" onClick={() => setMobileOpen(false)} />
      )}
    </div>
  )
}
