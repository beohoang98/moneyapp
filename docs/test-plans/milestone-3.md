# Milestone 3 Test Plan

**Feature scope**: M3 — Dashboard charts & period selector (E6-S2–E6-S4), CSV export (E6-S5), document scanning (E8-S1–E8-S5)  
**Last updated**: 2026-04-27  
**Author**: QA/UX Tester agent  
**TS numbering**: continues from M2 (last used: TS-40); M3 range **TS-41 – TS-50**

**Related docs**: [`docs/tickets.md`](../tickets.md), [`docs/milestone-03/implementation-plan.md`](../milestone-03/implementation-plan.md), [`docs/milestone-03/invoice-scan-user-flows.md`](../milestone-03/invoice-scan-user-flows.md).

---

## 1. Scope

### In scope (M3)

| Ticket | Area | Priority |
|---|---|---|
| E6-S2 | Dashboard period switcher (presets + custom range + URL persistence) | Should |
| E6-S3 | Income vs. Expenses bar chart (Recharts, monthly trend API) | Should |
| E6-S4 | Expense breakdown by category donut chart (Recharts, category API) | Should |
| E6-S5 | CSV export for Expenses and Income pages | Should |
| E8-S1 | Scanning settings UI + Test Connection endpoint | Must |
| E8-S2 | Backend scan endpoint (`POST /api/scanning/invoice`) + temp-delete | Must |
| E8-S3 | Invoice "Scan Invoice" button with health gate | Must |
| E8-S4 | Scan review form: confirm-to-save flow + cancel cleanup | Must |
| E8-S5 | Scanning robustness: timeout, concurrency limit, auth, logging | Must |

### Out of scope (do not test in M3)

- Batch / multi-image scanning (SC-16 — Won't have)
- On-device OCR / Tesseract.js (SC-15 — Won't have)
- Camera capture via `<input capture>` (E8-S8 — deferred M4)
- Scan audit log / `scan_results` table (SC-13 — deferred)
- Category auto-suggestion from vendor name (SC-14 — deferred)
- Multi-user / shared scanning credentials
- Dashboard customisation (widget reorder, pin) — not in M3 tickets
- Any M4 tickets

---

## 2. Traceability Table

| Ticket ID | Area | Priority | Test Suite | Test Case Count | Notes |
|---|---|---|---|---|---|
| E6-S2 | Dashboard Period Switcher | Should | TS-41 | 8 | Presets, custom range, URL persistence, recalculation |
| E6-S3 | Income vs. Expenses Chart | Should | TS-42 | 7 | Bar chart render, empty state, tooltip, no divide-by-100 |
| E6-S4 | Category Breakdown Chart | Should | TS-43 | 6 | Donut chart, proportions, empty state, legend |
| E6-S5 | CSV Export | Should | TS-44 | 8 | BOM, encoding, filters, raw int amounts, auth |
| E8-S1 | Scanning Settings & Test Connection | Must | TS-45 | 10 | CRUD, defaults, api_key redaction, test-connection paths |
| E8-S2 | Scan API & Temp Delete | Must | TS-46 | 9 | Happy path, timeout, bad JSON, bad MIME, path-traversal guard, auth |
| E8-S3 | Invoice Health Gate UI | Must | TS-47 | 6 | Greyed states, tooltip copy, focus re-fetch |
| E8-S4 | Scan Review Form & Confirm/Cancel | Must | TS-48 | 10 | Pre-fill, edits, line items, confidence badge, save, cancel, ESC/backdrop |
| E8-S5 | Scanning Robustness | Must | TS-49 | 8 | Schema validation, 429 concurrency, per-error UI copy, logs, auth sweep |
| Regression | M1 + M2 smoke | — | TS-50 | 6 | Login, expense CRUD, attachment, categories |

**Total test cases: ~78**

---

## 3. Environment Prerequisites

### 3.1 Base environment (required for all suites)

```
# Backend
cd backend && go run ./cmd/server
# env vars:
STORAGE_TYPE=local
LOCAL_STORAGE_PATH=./uploads   # created on startup
PORT=8080

# Frontend
cd frontend && npm run dev     # default port 5173
```

- SQLite DB auto-migrated on startup (migrations through `009_create_scanning_settings.up.sql` must be present for E8-S1+).
- A valid JWT token for all authenticated calls — obtain via `POST /api/auth/login`.

### 3.2 Environments

| Env tag | Extra config | Used by |
|---|---|---|
| `m3-base` | Local storage, no Ollama | E6-S2–E6-S5, E8-S1 (save/get), E8-S3 (disabled gate) |
| `m3-ollama` | Ollama running at `http://localhost:11434`, `qwen3-vl:4b` pulled | E8-S1 (test-connection success), E8-S2 (happy path), E8-S4 |
| `m3-ollama-missing-model` | Ollama running, model **not** pulled | E8-S1 (test-connection model-not-found path) |
| `m3-no-ollama` | Backend running, Ollama not started | E8-S1 (test-connection unreachable), E8-S3 (unhealthy gate) |
| `m3-minio` | `STORAGE_TYPE=s3`, MinIO at `localhost:9000` | E8-S2 (temp key + delete in S3), optional |

### 3.3 Ollama setup (for `m3-ollama`)

```bash
ollama serve            # starts API at :11434
ollama pull qwen3-vl:4b
# Verify:
curl http://localhost:11434/v1/models   # should include "qwen3-vl:4b"
```

### 3.4 Test image assets

Place in `e2e/fixtures/`:

| File | Use |
|---|---|
| `receipt_clear.jpg` | Happy-path scan (clear JPEG, ≤ 2 MB) |
| `receipt_blurry.jpg` | Partial-result / low-confidence scan |
| `unsupported.gif` | MIME-type rejection |
| `oversized.jpg` | > 10 MB file (size rejection) |
| `unicode_desc.csv` | Reference CSV with Vietnamese/accented chars (for manual comparison) |

---

## 4. Test Suites

---

### Test Suite TS-41 — Dashboard Period Switcher (E6-S2)

**Prerequisite**: M3 backend and frontend running. At least 5 expenses and 3 incomes seeded across two calendar months (e.g., March and April 2026).

| Test ID | Type | Description | Steps | Expected Result | Ref |
|---|---|---|---|---|---|
| TS-41-01 | Smoke | "This Month" preset recalculates summary | 1. Open `/dashboard`. 2. Click "This Month". | All summary card values reflect only the current calendar month. API request to `/api/dashboard/summary` includes `date_from=<first-of-month>&date_to=<today>`. | AC1 (month) |
| TS-41-02 | Smoke | "Last Month" preset recalculates summary | Click "Last Month". | Values match the previous calendar month. Verify via direct API call for the same date range. | AC1 (last month) |
| TS-41-03 | Smoke | "This Year" preset — Jan 1 to today | Click "This Year". | `date_from=<year>-01-01`, `date_to=<today>`. Summary reflects full year-to-date data. | AC1 (year) |
| TS-41-04 | Smoke | Custom date range — Apply recalculates within 2 s | 1. Enter `date_from=2026-01-01`, `date_to=2026-02-28`. 2. Click Apply. | Dashboard updates in ≤ 2 s (NF-02). Summary values match only Jan–Feb 2026 data. | AC1 (custom) |
| TS-41-05 | Regression | URL reflects selected period | 1. Select "Last Month". 2. Copy the URL. 3. Open it in a new tab. | New tab shows the same "Last Month" period without any additional interaction. `date_from` and `date_to` are present in the URL query params. | AC (URL persist) |
| TS-41-06 | Regression | Custom range — invalid (from > to) shows error | Enter `date_from=2026-04-30`, `date_to=2026-01-01` and Apply. | Inline validation error shown; no API call made. | NF-09 |
| TS-41-07 | Regression | Charts also respond to period change | Select "Last Month". | Both charts (TS-42, TS-43) re-fetch and show data for the new period. No stale chart from a previous period remains. | AC1 (integration) |
| TS-41-08 | Regression | Period selector visible and accessible | Navigate to `/dashboard` with keyboard only (Tab to `PeriodSelector`). | All preset buttons and date inputs are reachable via Tab. Selected preset has `aria-pressed="true"` or equivalent. | NF-14 (a11y) |

---

### Test Suite TS-42 — Income vs. Expenses Chart (E6-S3)

**Prerequisite**: Data for ≥ 2 months in the selected period. `recharts` installed.

| Test ID | Type | Description | Steps | Expected Result | Ref |
|---|---|---|---|---|---|
| TS-42-01 | Smoke | Chart renders with data for selected period | Open `/dashboard` with "This Year" selected; data exists for Jan–Apr. | A bar chart appears below the summary cards showing 4 month groups. Income bars (green) and Expense bars (red) are visible and labelled. | AC1 |
| TS-42-02 | Smoke | Month with no data shows zero bars | The period includes a month (e.g., Feb) with zero expenses and income. | Feb bar group appears in the X-axis with height 0; the chart is not broken or missing that month entirely. | AC2 |
| TS-42-03 | Smoke | Tooltip shows exact amount on hover | Hover over an expense bar for a month where you know the total is 300,000. | Tooltip displays `300000` (or `300,000` if formatted with a separator) — **not** `3000.00` or `3,000.00` (i.e., no divide-by-100). | AC3, NF-amounts |
| TS-42-04 | Regression | Y-axis values are in major units — no divide-by-100 | Compare Y-axis tick labels against known DB amounts (`SELECT SUM(amount) FROM expenses WHERE strftime('%Y-%m', date) = '2026-01'`). | Y-axis tick label matches the DB `SUM` value directly. If DB sum is `500000`, axis shows `500000` (or formatted `500,000`), never `5000`. | E6-S3 note |
| TS-42-05 | Regression | Chart is responsive — renders at 375px viewport | Resize browser window to 375px wide. | Chart remains visible, no horizontal overflow. `ResponsiveContainer` collapses bars to fit; no JS errors in console. | NF-13 |
| TS-42-06 | Regression | Changing period updates chart data | Switch period from "This Year" to "Last Month". | Chart re-fetches `GET /api/dashboard/monthly-trend` with new date range and updates. Old bars are not shown. | E6-S2 integration |
| TS-42-07 | Regression | Empty state — no data for selected period | Select a period (e.g., year 2020) for which no expenses or income exist. | Chart shows an empty state message (e.g., "No data for this period") rather than an empty Recharts canvas with no visible bars. | E6-S3 AC (empty) |

---

### Test Suite TS-43 — Category Breakdown Chart (E6-S4)

**Prerequisite**: At least 3 different category expenses in the selected period. Verify `GET /api/dashboard/expense-by-category` is implemented.

| Test ID | Type | Description | Steps | Expected Result | Ref |
|---|---|---|---|---|---|
| TS-43-01 | Smoke | Donut chart renders with correct segments | Select "This Month". DB has Food=300000, Transport=200000, Housing=500000. Open `/dashboard`. | Donut chart shows 3 segments. Food segment ≈ 30%, Transport ≈ 20%, Housing ≈ 50% of the total arc. | AC1 |
| TS-43-02 | Smoke | Legend shows category name and amount | Inspect chart legend. | Each legend entry shows `<category name>` and the amount. Amount is the raw int value from DB (500000), not divided by 100. | AC1 |
| TS-43-03 | Smoke | Tooltip on hover shows category, amount, and percentage | Hover over a donut segment. | Tooltip shows: category name, amount value, and percentage. Percentage sums to 100% across all segments. | AC1 |
| TS-43-04 | Regression | Empty state when no expenses for period | Select a period with zero expenses. | Chart shows an empty state message (e.g., "No expense data for this period"). No empty Recharts canvas rendered without context. | AC2 |
| TS-43-05 | Regression | Category color from DB used when available | A category "Food" has a custom color `#FF5733`. Reload dashboard. | The Food segment in the donut chart is colored `#FF5733`. | E6-S4 frontend note |
| TS-43-06 | Regression | Period change updates donut chart | Switch from "This Month" to "Last Month". | Chart re-fetches and updates segments. A category with zero spend in "Last Month" disappears from the chart. | E6-S4 integration |

---

### Test Suite TS-44 — CSV Export (E6-S5)

**Prerequisite**: 5+ expenses and 5+ incomes, at least one with a UTF-8 multi-byte character in the description (e.g., "Ăn sáng"). Apply at least one filter (e.g., by category) on the expenses page before testing filter-respected export.

| Test ID | Type | Description | Steps | Expected Result | Ref |
|---|---|---|---|---|---|
| TS-44-01 | Smoke | Export button on Expenses page triggers CSV download | 1. Navigate to `/expenses`. 2. Click "Export CSV". | Browser downloads a file named `expenses_<date>.csv`. `Content-Disposition: attachment; filename="expenses_<date>.csv"`. `Content-Type: text/csv; charset=utf-8`. | AC1 |
| TS-44-02 | Smoke | CSV opens in Excel without import wizard (BOM present) | Download the CSV. Open it in Microsoft Excel (Windows) or LibreOffice. | File opens with correct column headers and UTF-8 characters displayed without any manual encoding selection. This verifies BOM (`\xEF\xBB\xBF`) is present as first 3 bytes. Verify: `xxd expenses_*.csv \| head -1` → first bytes are `ef bb bf`. | AC2 |
| TS-44-03 | Smoke | Column headers and structure are correct | Inspect CSV rows. | Row 1 is `date,type,category,description,amount`. Subsequent rows have 5 columns per line. No extra columns or missing columns. | AC3 |
| TS-44-04 | Smoke | Amount column contains raw integer (no division) | Export an expense whose DB `amount` is `150000`. Check the `amount` column. | Value is `150000` — **not** `1500`, `1500.00`, or `150,000`. | AC4 |
| TS-44-05 | Smoke | UTF-8 multi-byte description survives export | Export expenses; one description is "Ăn sáng phở". Open CSV in a text editor set to UTF-8. | The description column contains "Ăn sáng phở" correctly. No mojibake or garbled characters. | AC5 |
| TS-44-06 | Regression | Active filters are respected in export | 1. On `/expenses`, filter by category "Food". 2. List shows only Food expenses. 3. Click "Export CSV". | Exported CSV contains **only** the rows matching the active filter (Food category). Rows from other categories are absent. | AC1 (filters) |
| TS-44-07 | Regression | **Negative**: Unauthenticated export returns 401 | `curl -s "http://localhost:8080/api/export/transactions?type=expense"` (no auth header). | HTTP 401. No CSV bytes in response body. | AC6, NF-07 |
| TS-44-08 | Regression | Export button absent from Dashboard page | Navigate to `/dashboard`. Inspect all buttons. | No "Export CSV" button is present anywhere on the Dashboard page. | E6-S5 scope note |

---

### Test Suite TS-45 — Scanning Settings & Test Connection (E8-S1)

**Prerequisite**: Migration `009_create_scanning_settings.up.sql` applied. JWT token available.

| Test ID | Type | Description | Steps | Expected Result | Ref |
|---|---|---|---|---|---|
| TS-45-01 | Smoke | Fresh DB returns default settings | `curl -H "Authorization: Bearer <token>" http://localhost:8080/api/settings/scanning` | HTTP 200; body: `{ "enabled": false, "base_url": "http://localhost:11434/v1", "model": "qwen3-vl:4b", "api_key_set": false }`. | AC1 |
| TS-45-02 | Smoke | Save settings — persisted across restart | 1. `PUT /api/settings/scanning` body `{ "enabled": true, "base_url": "http://myollama:11434/v1", "model": "qwen3-vl:4b" }`. 2. Restart backend. 3. `GET /api/settings/scanning`. | `GET` returns `{ "enabled": true, "base_url": "http://myollama:11434/v1", ... }`. Values survived restart (persisted in DB). | AC2 |
| TS-45-03 | Smoke | API key stored — GET returns `api_key_set: true`, not raw value | 1. `PUT /api/settings/scanning` body `{ "api_key": "sk-secret123" }`. 2. `GET /api/settings/scanning`. | Response contains `"api_key_set": true`. The string `"sk-secret123"` does **not** appear anywhere in the response body. | AC3, NF-07 |
| TS-45-04 | Smoke | Test Connection — Ollama healthy, model present (`m3-ollama`) | 1. Enable scanning, set `base_url=http://localhost:11434/v1`, `model=qwen3-vl:4b`. 2. Click "Test Connection" in UI (or `POST /api/settings/scanning/test`). | Response: `{ "ok": true, "message": "Connected" }` (or similar). UI shows success badge. Response arrives within 5 s. | AC4 |
| TS-45-05 | Regression | Test Connection — Ollama unreachable (`m3-no-ollama`) | Set `base_url=http://localhost:11434/v1`. Stop Ollama. Click "Test Connection". | HTTP 200 from backend; body: `{ "ok": false, "message": "Cannot reach http://localhost:11434/v1" }` (or similar). UI shows error badge. No crash. | AC5 |
| TS-45-06 | Regression | Test Connection — model not found (`m3-ollama-missing-model`) | Ollama running; model not pulled. Click "Test Connection". | `{ "ok": false, "message": "Model qwen3-vl:4b not found …" }`. UI surfaces the specific error message. | Flow doc §1.2 edge case 2 |
| TS-45-07 | Regression | Test Connection — invalid API key (401 from Ollama) | Set `api_key` to an invalid value against an Ollama instance that requires auth. Click "Test Connection". | `{ "ok": false, "message": "Invalid API key" }`. | Flow doc §1.3 |
| TS-45-08 | Regression | **Negative**: Invalid `base_url` (not http/https) | `PUT /api/settings/scanning` with `{ "base_url": "ftp://bad" }`. | HTTP 400. Body contains validation error referencing `base_url`. Settings not saved. | E8-S1 backend task (URL validation) |
| TS-45-09 | Regression | Disable scanning — toggle OFF | `PUT /api/settings/scanning` with `{ "enabled": false }`. | `GET` returns `"enabled": false`. Scan affordance is disabled app-wide (verified via TS-47). | AC2, SC-11 |
| TS-45-10 | Regression | **Negative**: Auth required — all three endpoints return 401 unauthenticated | `GET`, `PUT`, and `POST /api/settings/scanning/test` without `Authorization` header. | All three return HTTP 401. No settings data exposed. | AC6, NF-07 |

---

### Test Suite TS-46 — Scan API & Temp Delete (E8-S2)

**Prerequisite**: Scanning enabled in DB. `m3-ollama` environment for TS-46-01. All others can use `m3-base` with a mocked or offline Ollama.

| Test ID | Type | Description | Steps | Expected Result | Ref |
|---|---|---|---|---|---|
| TS-46-01 | Smoke | Happy path — JPEG → structured JSON response (`m3-ollama`) | `curl -H "Authorization: Bearer <token>" -F "image=@receipt_clear.jpg" http://localhost:8080/api/scanning/invoice` | HTTP 200. Body contains `{ "scan_result": { "vendor": "...", "date": "YYYY-MM-DD", "total_amount": <int>, "currency": "...", "line_items": [...], "confidence": {...} }, "temp_storage_key": "scan-tmp/..." }`. `temp_storage_key` starts with `scan-tmp/`. File present in `LOCAL_STORAGE_PATH/scan-tmp/`. | AC1 |
| TS-46-02 | Smoke | Scanning disabled → 503 | Set `enabled=false`. `POST /api/scanning/invoice` with valid JPEG. | HTTP 503; body `{ "error": "scanning_disabled" }`. No file stored. | AC2 |
| TS-46-03 | Smoke | Unsupported file type (GIF) → 400 | `POST /api/scanning/invoice` with `@unsupported.gif`. | HTTP 400; body contains `{ "error": … }` referencing unsupported type. No file stored in `scan-tmp/`. | AC3 |
| TS-46-04 | Smoke | Vision API timeout → 504 | Configure Ollama `base_url` to a slow endpoint or use a mock that delays > 60 s. `POST /api/scanning/invoice`. | HTTP 504; `{ "error": "scan_timeout" }`. The temp file is **deleted** from ObjectStore (no orphan). | AC4, E8-S5 timeout |
| TS-46-05 | Regression | Malformed JSON from vision model → 422 | Mock Ollama to return `model response: "not json at all"` as assistant content. `POST /api/scanning/invoice`. | HTTP 422; `{ "error": "extraction_failed", "detail": "..." }`. No 500. | AC5 |
| TS-46-06 | Regression | Partial JSON (missing `total_amount`) → 422 | Mock Ollama to return `{ "vendor": "Shop" }` (no `total_amount`). | HTTP 422; `{ "error": "extraction_failed" }`. | Flow doc §3 edge case 5 |
| TS-46-07 | Regression | `DELETE /api/scanning/temp` — valid `scan-tmp/` key cleans up | 1. After TS-46-01, note `temp_storage_key`. 2. `DELETE /api/scanning/temp` body `{ "storage_key": "scan-tmp/<uuid>_receipt_clear.jpg" }`. | HTTP 204. File no longer present in `LOCAL_STORAGE_PATH/scan-tmp/`. | E8-S2 AC (delete) |
| TS-46-08 | Regression | **Negative**: `DELETE /api/scanning/temp` with non-`scan-tmp/` key → 400 (path-traversal guard) | `DELETE /api/scanning/temp` body `{ "storage_key": "attachments/expense/1/somefile.jpg" }`. | HTTP 400. Backend rejects key that does not start with `scan-tmp/`. No file deleted. | AC6, NF security |
| TS-46-09 | Regression | **Negative**: Auth on both endpoints | `POST /api/scanning/invoice` and `DELETE /api/scanning/temp` without Authorization header. | Both return HTTP 401. | AC7, NF-07 |

---

### Test Suite TS-47 — Invoice Health Gate UI (E8-S3)

**Prerequisite**: M3 frontend running.

| Test ID | Type | Description | Steps | Expected Result | Ref |
|---|---|---|---|---|---|
| TS-47-01 | Smoke | Button disabled when scanning disabled | Set `enabled=false` (via API or Settings UI). Navigate to `/invoices`. | "Scan Invoice" button is visible but disabled (greyed-out). Cursor shows `not-allowed`. | AC1 |
| TS-47-02 | Smoke | Disabled button tooltip explains why | Inspect disabled button (hover or inspect `title`/`aria-describedby`). | Tooltip/accessible label contains a human-readable message referencing Settings (e.g., "Document scanning is not configured — go to Settings to enable it"). | AC1, E8-S3 frontend task |
| TS-47-03 | Smoke | Button enabled when scanning healthy (`m3-ollama`) | Set `enabled=true`, Ollama healthy. Navigate to `/invoices`. | "Scan Invoice" button is **enabled** (not greyed). Clicking opens the scan modal (E8-S4). | AC2 |
| TS-47-04 | Regression | Button disabled when scanning enabled but Ollama unhealthy (`m3-no-ollama`) | `enabled=true`, Ollama not running. Navigate to `/invoices`. | Button is disabled. Tooltip reflects the health check failure message (e.g., "Scanning service unreachable. Check Settings → Document Scanning"). | AC2 (unhealthy variant) |
| TS-47-05 | Regression | Focus re-fetch on tab return updates state | 1. `/invoices` page open; scanning disabled (button greyed). 2. Open new tab → `/settings`, enable scanning. 3. Return (focus) to `/invoices` tab. | "Scan Invoice" button transitions to enabled state **without** manual page refresh. (`useScanningHealth` re-fetches on `window focus` event.) | AC3 |
| TS-47-06 | Regression | **Negative**: `GET /api/scanning/health` unauthenticated → 401 | `curl http://localhost:8080/api/scanning/health` (no auth). | HTTP 401. | AC5, NF-07 |

---

### Test Suite TS-48 — Scan Review Form & Confirm/Cancel (E8-S4)

**Prerequisite**: `m3-ollama` environment. Scanning enabled, Ollama healthy, `receipt_clear.jpg` available.

| Test ID | Type | Description | Steps | Expected Result | Ref |
|---|---|---|---|---|---|
| TS-48-01 | Smoke | Review form pre-filled after scan | 1. Click "Scan Invoice". 2. Select `receipt_clear.jpg`. 3. Wait for scan. | Review form appears with `vendor`, `date`, and `total_amount` pre-filled from scan result. Spinner/loading overlay shown during scan, dismissed on completion. | AC1 |
| TS-48-02 | Smoke | Line items table rendered (SC-07) | Scan result includes `line_items`. | A table below the main fields shows each line item with `description` and `amount` columns. | AC2, SC-07 |
| TS-48-03 | Smoke | Low-confidence field shows warning icon (SC-09) | Scan result returns `confidence: { "total_amount": "low" }`. | A ⚠ warning icon appears next to the `total_amount` field. Tooltip: "Low confidence — please verify". | AC3, SC-09 |
| TS-48-04 | Smoke | Confirm save with edited vendor → invoice uses edited value | 1. Scan an image. 2. Change "Vendor" field to "My Edited Vendor". 3. Click "Create Invoice". | `POST /api/invoices` body contains `title` or `vendor` = "My Edited Vendor" (not the raw LLM output). Invoice record in DB uses the edited value. | AC4 |
| TS-48-05 | Smoke | Confirm save → invoice visible in list + image attached | 1. Complete TS-48-04. 2. Close modal. 3. Check `/invoices` list. 4. Open invoice detail. | Invoice appears in list. In the detail view, the scanned image appears in the attachments section (same `AttachmentList` UI from E7-S4). | AC5, SC-10 |
| TS-48-06 | Regression | "Create Invoice" disabled until required fields filled | 1. Open scan modal. 2. After scan, clear the `total_amount` field. | "Create Invoice" button is disabled. Button re-enables when `title`, `date`, and `total_amount` are all non-empty. | E8-S4 frontend task |
| TS-48-07 | Regression | Cancel after scan — temp image deleted, no record created | 1. Scan image (note `temp_storage_key` from network tab). 2. Click "Cancel". | `DELETE /api/scanning/temp?storage_key=…` (or request body) sent. `temp_storage_key` file absent from `LOCAL_STORAGE_PATH/scan-tmp/`. No invoice row added to DB. Modal closes. | AC6, SC-06, OQ9 |
| TS-48-08 | Regression | ESC key triggers same cleanup as Cancel | 1. Scan image. 2. Press `Escape`. | Same cleanup as TS-48-07. `DELETE /api/scanning/temp` is called. No DB record created. | AC6, SC-06 |
| TS-48-09 | Regression | Backdrop click triggers same cleanup as Cancel | 1. Scan image. 2. Click outside the modal. | Same cleanup as TS-48-07. | AC6 |
| TS-48-10 | Regression | Scan error (422) — modal stays open, user can retry | Mock backend to return 422 `{ "error": "extraction_failed" }` for the scan call. Click "Scan". | Modal shows error message ("Could not extract data…"). File picker remains accessible. User can select a different image and click Scan again without reopening the modal. | AC7, E8-S4 frontend task |

---

### Test Suite TS-49 — Scanning Robustness, Security & Concurrency (E8-S5)

**Prerequisite**: Combination of `m3-ollama` and `m3-base`. Some tests require sending concurrent requests; use `curl --parallel` or a simple Go test harness.

| Test ID | Type | Description | Steps | Expected Result | Ref |
|---|---|---|---|---|---|
| TS-49-01 | Smoke | Schema validation: missing `total_amount` → 422 not 500 | Mock Ollama to return `{ "vendor": "Shop", "date": "2026-04-27", "currency": "VND" }` (no `total_amount`). `POST /api/scanning/invoice`. | HTTP 422 `{ "error": "extraction_failed", "detail": "..." }`. HTTP 500 is a failure. | AC1, E8-S5 backend |
| TS-49-02 | Smoke | Schema validation: invalid date format → 422 | Mock Ollama to return `{ "total_amount": 100, "date": "27/04/2026" }` (wrong format). | HTTP 422 `{ "error": "extraction_failed" }`. | AC1 |
| TS-49-03 | Smoke | Concurrency limit: 3rd simultaneous scan → 429 | 1. Send 2 `POST /api/scanning/invoice` requests simultaneously (slow mock — delay each > 2 s). 2. While both in flight, send a 3rd. | The 3rd request returns HTTP 429 `{ "error": "too_many_scans" }`. The first two may succeed or fail based on mock, but the 3rd is rejected before Ollama is called. | AC2, E8-S5 backend (semaphore 2) |
| TS-49-04 | Regression | UI: 504 timeout → specific Ollama troubleshooting message | Trigger a scan timeout (mock or real). | Frontend shows: "Scanning timed out. Make sure Ollama is running and the model is loaded (`ollama run qwen3-vl:4b`)." — **not** a generic "Something went wrong". | AC3, E8-S5 frontend |
| TS-49-05 | Regression | UI: 429 → "another scan in progress" message | Trigger 429 from backend. | Frontend shows: "Another scan is in progress. Please wait a moment and try again." | AC3, E8-S5 frontend |
| TS-49-06 | Regression | No image bytes in server logs | Run a successful scan. Inspect server stdout/log file. | No `base64` string, no raw file bytes, no JPEG/PNG binary data in any log line. Log entry for the scan contains only `file_size_bytes` and `content_type` fields. | AC5, E8-S5 backend (NF-27) |
| TS-49-07 | Regression | Auth sweep — all 6 scanning routes require JWT | Send unauthenticated requests to: `GET /api/settings/scanning`, `PUT /api/settings/scanning`, `POST /api/settings/scanning/test`, `GET /api/scanning/health`, `POST /api/scanning/invoice`, `DELETE /api/scanning/temp`. | All 6 return HTTP 401. | AC6, NF-07 |
| TS-49-08 | Regression | Error stays in modal — user can retry | After any error (422, 504, 429, network), verify the modal remains open and the file input is re-enabled. | Modal does not close on error. File picker is interactive. User selects a different image and re-scans without reopening the modal. | AC4, E8-S5 frontend |

---

### Test Suite TS-50 — Regression Smoke (M1 + M2)

Run before M3 sign-off to confirm no regressions were introduced.

| Test ID | Type | Description | Steps | Expected Result |
|---|---|---|---|---|
| TS-50-01 | Smoke | Login still works | `POST /api/auth/login` with valid credentials. | HTTP 200 with JWT token. |
| TS-50-02 | Smoke | Expense CRUD | Create expense → verify in list → edit description → delete → verify gone. | All operations complete without error. List refreshes correctly. |
| TS-50-03 | Smoke | Invoice list loads with existing data | Navigate to `/invoices`. | Invoice list renders. "Scan Invoice" button visible (state depends on scanning config — see TS-47). |
| TS-50-04 | Smoke | Attachment upload on expense still works | Upload `receipt_clear.jpg` to an expense via the FileUpload component. | HTTP 201. File appears in `AttachmentList`. Download returns correct bytes. |
| TS-50-05 | Smoke | Category management intact | Create a new category "Utilities". Assign it to a new expense. | Category appears in dropdown. Expense created with correct category. Dashboard category chart updates if applicable. |
| TS-50-06 | Smoke | M2 sorting/filtering — expense sort by amount still works | On `/expenses`, sort by amount descending. | Expenses ordered highest to lowest. No regression from M3 chart/export code changes. |

---

## 5. Given / When / Then — Critical Path Scenarios

The following critical paths have formal BDD-style definitions for automation or traceability review.

### CP-1: Dashboard Period → Chart Data (E6-S2 + E6-S3 + E6-S4 integrated)

**Given** the dashboard has expense data for Jan–Apr 2026  
**When** the user selects "This Year"  
**Then** both the monthly trend bar chart and the category donut chart re-fetch and show data for Jan 1 – today, all amounts in major units (no divide-by-100), and the URL contains `date_from=2026-01-01`.

### CP-2: CSV Export with Filters (E6-S5)

**Given** the user is on `/expenses` with a "Food" category filter active, showing 12 matching expenses  
**When** the user clicks "Export CSV"  
**Then** the browser downloads `expenses_<date>.csv`; the file starts with BOM bytes `\xEF\xBB\xBF`; it contains exactly 13 rows (1 header + 12 data); all `amount` values are raw integers matching the DB; all rows have `type=expense` and `category=Food`.

### CP-3: Scan Happy Path (E8-S2 + E8-S4 integrated)

**Given** scanning is enabled and Ollama is healthy with `qwen3-vl:4b`  
**When** the user opens the scan modal, selects `receipt_clear.jpg`, and clicks Scan  
**Then** the review form appears pre-filled within 60 s; the user reviews the values; clicks "Create Invoice";
a new invoice record is created; the scanned image is promoted from `scan-tmp/` to a permanent attachment key; the invoice list shows the new row; the image is viewable in the attachment section; no `scan-tmp/` file remains in ObjectStore.

### CP-4: Cancel After Scan — No Orphan (E8-S4 + E8-S5)

**Given** scanning completed and the review form is open with `temp_storage_key=scan-tmp/abc123_receipt.jpg`  
**When** the user clicks "Cancel" (or presses Escape, or clicks the backdrop)  
**Then** `DELETE /api/scanning/temp` is called with `storage_key=scan-tmp/abc123_receipt.jpg`; backend responds 204; the file is absent from ObjectStore; no invoice or attachment row is inserted in the database; the modal closes.

### CP-5: Scanning Disabled — Full UI Gate (E8-S1 + E8-S3)

**Given** `PUT /api/settings/scanning` was called with `{ "enabled": false }`  
**When** the user navigates to `/invoices`  
**Then** `GET /api/scanning/health` returns `{ "ok": false, "message": "Scanning is disabled. Enable it in Settings." }`; the "Scan Invoice" button is visible but disabled and grey; its accessible tooltip contains the message; clicking the button has no effect (no modal opens).

---

## 6. Playwright Automation — Suggested Hooks

> This section is **advisory**: it identifies which test cases are highest-value for automation and provides starter patterns. Implementation lives under `e2e/`.

### Priority order

| Priority | Suite | Test IDs | Rationale |
|---|---|---|---|
| P1 | TS-44 CSV export | TS-44-01, TS-44-02, TS-44-04 | BOM check and amount-division guard are regression-prone and hard to check manually |
| P1 | TS-47 Health gate | TS-47-01, TS-47-03, TS-47-05 | Toggle-based UI state is brittle; focus-refresh behaviour needs reliable automation |
| P1 | TS-48 Cancel cleanup | TS-48-07, TS-48-08, TS-48-09 | Orphan-file prevention is a data-integrity invariant — automate first |
| P2 | TS-41 Period switcher | TS-41-05 (URL persist) | URL state is easy to assert; covers full E6-S2 integration |
| P2 | TS-45 Settings | TS-45-03 (api_key_set) | Security property — verify raw key never returned |
| P3 | TS-42/TS-43 Charts | TS-42-04 (no divide) | Chart rendering is harder to assert; focus on data source value check via network intercept |

### Suggested file layout

```
e2e/
  specs/
    dashboard-period.spec.ts     # TS-41
    charts.spec.ts               # TS-42, TS-43
    csv-export.spec.ts           # TS-44
    scanning-settings.spec.ts    # TS-45
    scanning-gate.spec.ts        # TS-47
    scan-review.spec.ts          # TS-48
  fixtures/
    receipt_clear.jpg
    unsupported.gif
  pages/
    DashboardPage.ts
    ScanModalPage.ts
    SettingsPage.ts
```

### Starter pattern — CSV BOM check (TS-44-02)

```typescript
// e2e/specs/csv-export.spec.ts
import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

test.describe('CSV Export', () => {
  test.beforeEach(async ({ page }) => {
    // log in and navigate to expenses page
    await page.goto('/login');
    await page.getByTestId('username').fill('admin');
    await page.getByTestId('password').fill('password');
    await page.getByRole('button', { name: 'Login' }).click();
    await page.waitForURL('/dashboard');
    await page.goto('/expenses');
  });

  test('TS-44-02: CSV download starts with UTF-8 BOM', async ({ page }) => {
    const [download] = await Promise.all([
      page.waitForEvent('download'),
      page.getByRole('button', { name: 'Export CSV' }).click(),
    ]);
    const filePath = path.join('/tmp', download.suggestedFilename());
    await download.saveAs(filePath);

    const buf = fs.readFileSync(filePath);
    // BOM: EF BB BF
    expect(buf[0]).toBe(0xEF);
    expect(buf[1]).toBe(0xBB);
    expect(buf[2]).toBe(0xBF);
  });

  test('TS-44-04: amount column contains raw integer, not divided', async ({ page }) => {
    // Assumes an expense with known amount=150000 exists via API setup
    const [download] = await Promise.all([
      page.waitForEvent('download'),
      page.getByRole('button', { name: 'Export CSV' }).click(),
    ]);
    const filePath = path.join('/tmp', download.suggestedFilename());
    await download.saveAs(filePath);

    const content = fs.readFileSync(filePath, 'utf-8');
    const lines = content.split('\n').filter(Boolean);
    const dataLines = lines.slice(1); // skip header
    const amountColIndex = lines[0].replace(/^\xEF\xBB\xBF/, '').split(',').indexOf('amount');

    for (const line of dataLines) {
      const cols = line.split(',');
      const amount = cols[amountColIndex];
      // Amount must be an integer string — no decimal point
      expect(amount).toMatch(/^\d+$/);
    }
  });
});
```

### Starter pattern — Scan cancel / orphan check (TS-48-07)

```typescript
// e2e/specs/scan-review.spec.ts
import { test, expect } from '@playwright/test';
import path from 'path';

test.describe('Scan Review Flow', () => {
  test('TS-48-07: cancel after scan calls DELETE temp and closes modal', async ({ page }) => {
    await page.goto('/invoices');

    // Track the DELETE temp request
    let deleteTempCalled = false;
    let deletedKey = '';
    page.on('request', req => {
      if (req.method() === 'DELETE' && req.url().includes('/api/scanning/temp')) {
        deleteTempCalled = true;
        try {
          deletedKey = JSON.parse(req.postData() ?? '{}').storage_key ?? '';
        } catch { /* ignore */ }
      }
    });

    await page.getByRole('button', { name: 'Scan Invoice' }).click();

    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(path.join(__dirname, '../fixtures/receipt_clear.jpg'));

    // Wait for review form to appear (scan complete)
    await expect(page.getByTestId('scan-review-form')).toBeVisible({ timeout: 20_000 });

    await page.getByRole('button', { name: 'Cancel' }).click();

    expect(deleteTempCalled).toBe(true);
    expect(deletedKey).toMatch(/^scan-tmp\//);

    // Modal must be closed
    await expect(page.getByTestId('scan-review-form')).not.toBeVisible();
  });
});
```

### Chart data-source intercept pattern (TS-42-04 — no divide-by-100)

Rather than reading pixel heights, intercept the API response and assert on the data passed to Recharts:

```typescript
test('TS-42-04: monthly trend values are raw ints, not divided by 100', async ({ page }) => {
  const [response] = await Promise.all([
    page.waitForResponse(r => r.url().includes('/api/dashboard/monthly-trend') && r.status() === 200),
    page.goto('/dashboard'),
  ]);
  const body = await response.json();
  // Every total_income and total_expenses must be a whole number > 0
  for (const row of body.data) {
    expect(Number.isInteger(row.total_income)).toBe(true);
    expect(Number.isInteger(row.total_expenses)).toBe(true);
    // If we know a specific month's expected sum from the DB fixture:
    // e.g., Jan 2026 expenses = 500000
    // expect(row.total_expenses).toBeGreaterThanOrEqual(100); // sanity: clearly not cents
  }
});
```

---

## 7. Coverage Gaps & Notes

| Gap | Ticket | Detail |
|---|---|---|
| **Temp image leaked on server crash mid-scan** | E8-S2 / E8-S5 | If the server process is killed between the temp upload and the LLM call, the `scan-tmp/` object is never deleted. No cleanup job is specified in M3 scope — document as known limitation; a periodic `scan-tmp/` sweep may be addressed in M4. |
| **`POST /api/attachments` with `source_storage_key` — move vs copy+delete** | E8-S4 | The ticket allows implement "move" as copy+delete. If the copy succeeds but delete fails, the temp object is orphaned. Test TS-48-05 covers the happy path, but a storage-failure mid-move is not explicitly tested — add to M4 robustness backlog. |
| **Chart tooltip accessibility** | E6-S3/E6-S4 | Recharts `<Tooltip>` is not keyboard-accessible by default. WCAG SC 1.4.3 (contrast) and SC 2.1.1 (keyboard) may not be satisfied. Flag as an M4 accessibility debt item. |
| **CSV injection** | E6-S5 | If a description starts with `=`, `-`, `+`, or `@`, some spreadsheet apps treat it as a formula. Go's `encoding/csv` does not sanitise formula prefixes. The ticket does not require this; document as a known limitation for a finance app with self-hosted, trusted data. |
| **Scan concurrency counter reset on restart** | E8-S5 | The semaphore is an in-process buffered channel. Restarting the server resets the counter — any scan that was in-flight at restart leaves the counter at 0. The temp image from the aborted scan is orphaned. Document as known; no AC covers it. |
| **`api_key_set` vs empty string edge case** | E8-S1 | If `api_key` is saved as `""` (empty string), `api_key_set` should be `false`. Verify the backend does not set `api_key_set: true` for a blank key. No explicit AC; add to TS-45 extended coverage. |

---

*End of Milestone 3 Test Plan*
