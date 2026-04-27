import { useEffect, useState, useCallback, useRef } from 'react'
import { getScanningHealth } from '../api/scanning'

export function useScanningHealth() {
  const [isHealthy, setIsHealthy] = useState(false)
  const [message, setMessage] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const mountedRef = useRef(true)

  const fetchHealth = useCallback(() => {
    getScanningHealth()
      .then((res) => {
        if (mountedRef.current) {
          setIsHealthy(res.ok)
          setMessage(res.message)
          setIsLoading(false)
        }
      })
      .catch(() => {
        if (mountedRef.current) {
          setIsHealthy(false)
          setMessage('Could not reach the server')
          setIsLoading(false)
        }
      })
  }, [])

  useEffect(() => {
    mountedRef.current = true
    fetchHealth()

    const onFocus = () => fetchHealth()
    window.addEventListener('focus', onFocus)
    return () => {
      mountedRef.current = false
      window.removeEventListener('focus', onFocus)
    }
  }, [fetchHealth])

  return { isHealthy, message, isLoading }
}
