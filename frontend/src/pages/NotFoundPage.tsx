import { Link } from 'react-router-dom'

export function NotFoundPage() {
  return (
    <div style={{ textAlign: 'center', padding: '80px 16px' }}>
      <h1>404</h1>
      <p>Page not found.</p>
      <Link to="/" style={{ color: 'var(--accent)' }}>
        Go to Dashboard
      </Link>
    </div>
  )
}
