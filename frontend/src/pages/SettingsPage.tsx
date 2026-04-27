import { useEffect, useState } from 'react'
import { getScanningSettings, updateScanningSettings, testScanningConnection } from '../api/scanning'
import { useToast } from '../hooks/useToast'
import type { ScanningSettingsResponse } from '../types/scanning'
import './Pages.css'

export function SettingsPage() {
  const [settings, setSettings] = useState<ScanningSettingsResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<{ ok: boolean; message: string } | null>(null)
  const { addToast } = useToast()

  const [enabled, setEnabled] = useState(false)
  const [baseUrl, setBaseUrl] = useState('')
  const [model, setModel] = useState('')
  const [apiKey, setApiKey] = useState('')

  useEffect(() => {
    getScanningSettings()
      .then((s) => {
        setSettings(s)
        setEnabled(s.enabled)
        setBaseUrl(s.base_url)
        setModel(s.model)
      })
      .catch(() => addToast('Failed to load scanning settings', 'error'))
      .finally(() => setLoading(false))
  }, [addToast])

  const handleSave = async () => {
    setSaving(true)
    try {
      const updated = await updateScanningSettings({
        enabled,
        base_url: baseUrl,
        model,
        ...(apiKey ? { api_key: apiKey } : {}),
      })
      setSettings(updated)
      setApiKey('')
      addToast('Scanning settings saved', 'success')
    } catch {
      addToast('Failed to save settings', 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleTest = async () => {
    setTesting(true)
    setTestResult(null)
    try {
      const result = await testScanningConnection({
        base_url: baseUrl,
        model,
        ...(apiKey ? { api_key: apiKey } : {}),
      })
      setTestResult(result)
    } catch {
      setTestResult({ ok: false, message: 'Failed to test connection' })
    } finally {
      setTesting(false)
    }
  }

  if (loading) {
    return (
      <div>
        <h1>Settings</h1>
        <div className="loading">Loading...</div>
      </div>
    )
  }

  const fieldsDisabled = !enabled

  return (
    <div>
      <h1>Settings</h1>

      <div className="settings-section">
        <h2>Document Scanning</h2>
        <p style={{ color: 'var(--text)', fontSize: 14, marginBottom: 16 }}>
          Configure an Ollama-compatible vision API to scan receipts and invoices.
        </p>

        <div className="settings-field">
          <label className="settings-toggle">
            <input
              type="checkbox"
              checked={enabled}
              onChange={(e) => setEnabled(e.target.checked)}
            />
            <span>Enable Document Scanning</span>
          </label>
        </div>

        <div className="settings-field">
          <label>Base URL</label>
          <input
            type="text"
            value={baseUrl}
            onChange={(e) => setBaseUrl(e.target.value)}
            placeholder="http://localhost:11434/v1"
            disabled={fieldsDisabled}
          />
        </div>

        <div className="settings-field">
          <label>Model</label>
          <input
            type="text"
            value={model}
            onChange={(e) => setModel(e.target.value)}
            placeholder="qwen3-vl:4b"
            disabled={fieldsDisabled}
          />
        </div>

        <div className="settings-field">
          <label>API Key (optional)</label>
          <input
            type="password"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            placeholder={settings?.api_key_set ? 'API key is saved — enter a new value to replace' : 'Leave empty if not required'}
            disabled={fieldsDisabled}
          />
        </div>

        <div className="settings-actions">
          <button
            className="btn btn-sm"
            onClick={handleTest}
            disabled={fieldsDisabled || testing}
          >
            {testing ? 'Testing...' : 'Test Connection'}
          </button>
          <button
            className="btn btn-primary btn-sm"
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? 'Saving...' : 'Save Settings'}
          </button>
        </div>

        {testResult && (
          <div
            className={`settings-test-result ${testResult.ok ? 'settings-test-result--ok' : 'settings-test-result--error'}`}
          >
            {testResult.ok ? '✓' : '✗'} {testResult.message}
          </div>
        )}
      </div>
    </div>
  )
}
