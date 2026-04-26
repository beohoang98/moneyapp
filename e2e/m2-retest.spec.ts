import { test, expect } from '@playwright/test'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const e2eDir = path.dirname(fileURLToPath(import.meta.url))
const root = path.join(e2eDir, '..')
const screenshotsDir = path.join(e2eDir, 'test-results/screenshots')
const fixtureJpeg = path.join(root, 'testing/files/test-invoice.jpeg')
const API_BASE = 'http://localhost:8080/api'

test.beforeAll(() => {
  fs.mkdirSync(screenshotsDir, { recursive: true })
  if (!fs.existsSync(fixtureJpeg)) {
    throw new Error(`Missing fixture: ${fixtureJpeg}`)
  }
})

test.describe('M2 QA re-test (Playwright + screenshots)', () => {
  test('login, core pages, expense JPEG upload', async ({ page }) => {
    await page.goto('/login')
    await page.screenshot({ path: path.join(screenshotsDir, '01-login.png'), fullPage: true })

    await page.locator('#username').fill('admin')
    await page.locator('#password').fill('changeme')
    await page.getByRole('button', { name: /sign in/i }).click()
    await page.waitForURL('**/dashboard', { timeout: 30_000 })
    await page.screenshot({ path: path.join(screenshotsDir, '02-dashboard.png'), fullPage: true })

    const loginRes = await page.request.post(`${API_BASE}/auth/login`, {
      data: { username: 'admin', password: 'changeme' },
    })
    const { token } = await loginRes.json()

    await page.request.post(`${API_BASE}/expenses`, {
      headers: { Authorization: `Bearer ${token}` },
      data: { amount: 1000, date: '2026-01-15', category_id: 1, description: 'E2E seed expense' },
    })

    await page.goto('/expenses')
    await expect(page.getByRole('heading', { name: 'Expenses' })).toBeVisible()
    await page.screenshot({ path: path.join(screenshotsDir, '03-expenses.png'), fullPage: true })

    await page.getByRole('button', { name: 'Files' }).first().click()
    await expect(page.getByRole('heading', { name: 'Expense Attachments' })).toBeVisible()
    await page.screenshot({ path: path.join(screenshotsDir, '04-attachments-modal-before-upload.png'), fullPage: true })

    await page.locator('.form-modal input[type="file"]').setInputFiles(fixtureJpeg)
    await expect(page.getByText('test-invoice.jpeg').first()).toBeVisible({ timeout: 120_000 })
    await expect(page.getByText(/Attachments \(\d+\)/)).toBeVisible()
    await page.screenshot({ path: path.join(screenshotsDir, '05-attachments-after-upload.png'), fullPage: true })

    await page.goto('/invoices')
    await expect(page.getByRole('heading', { name: 'Invoices' })).toBeVisible()
    await page.screenshot({ path: path.join(screenshotsDir, '06-invoices.png'), fullPage: true })

    await page.goto('/categories')
    await expect(page.getByRole('heading', { name: 'Categories', exact: true })).toBeVisible()
    await page.screenshot({ path: path.join(screenshotsDir, '07-categories.png'), fullPage: true })
  })
})
