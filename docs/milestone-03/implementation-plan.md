# Milestone 3 — Implementation Plan

**Version**: 1.0  
**Date**: 2026-04-27  
**Author**: Fullstack Tech Lead  
**Status**: Ready for execution

**Related docs**: [`docs/tickets.md`](../tickets.md) (ticket AC), [`docs/test-plans/milestone-3.md`](../test-plans/milestone-3.md) (QA traceability + TS cases), [`invoice-scan-user-flows.md`](invoice-scan-user-flows.md) (UX flows).

---

## 1. Executive Summary

Milestone 3 delivers two major feature areas: **financial reporting** (dashboard period switcher, income-vs-expense bar chart, expense-by-category donut chart, and CSV export) and **document scanning** (Ollama-based vision model settings, health-gated scan button, image-to-structured-JSON extraction, review form with confidence indicators, confirm-to-save flow with temp image lifecycle, and concurrency/robustness hardening). Together these satisfy E6-S2 through E6-S5 and E8-S1 through E8-S5, unlocking the ability to visualize spending trends, export data to spreadsheets, and scan receipts/invoices with automatic field extraction via a local Ollama instance.

---

## 2. Workstream Breakdown

Implementation is split into five sequential phases with clear dependency ordering. Each phase is independently deployable and testable.

### Phase A — Dashboard Charts & Period Switcher

**Tickets**: E6-S2, E6-S3, E6-S4  
**Estimated effort**: 3–4 days  

1. Backend: add `GetMonthlyTrend()` and `GetExpenseByCategory()` to `DashboardService`; register two new handler endpoints.
2. Frontend: install Recharts, build `PeriodSelector`, `MonthlyTrendChart`, `CategoryBreakdownChart`; integrate into `DashboardPage`.

### Phase B — CSV Export

**Tickets**: E6-S5  
**Estimated effort**: 2 days  

1. Backend: new `export_handler.go` with CSV streaming, UTF-8 BOM, integer amounts.
2. Frontend: `ExportButton` component on `ExpensesPage` and `IncomePage`.

### Phase C — Scanning Settings & Health

**Tickets**: E8-S1, E8-S3  
**Estimated effort**: 3 days  

1. Backend: migrations `009_create_scanning_settings.up.sql` and `010_update_default_scanning_model.up.sql` (bumps legacy `moondream:1.8b` rows to `qwen3-vl:4b`), `ScanningSettings` model, `ScanningService` (settings CRUD + TestConnection), `ScanningHandler` (settings + health endpoints).
2. Frontend: Document Scanning section on `SettingsPage`, `useScanningHealth` hook, health-gated "Scan Invoice" button on `InvoicesPage`.

### Phase D — Scan Endpoint & Temp Storage

**Tickets**: E8-S2  
**Estimated effort**: 3–4 days  

1. Backend: `ScanImage()` in `ScanningService` (upload to `scan-tmp/`, base64 encode, call Ollama `/v1/chat/completions`, parse response), `DeleteTempScan()`, handler methods `POST /api/scanning/invoice` and `DELETE /api/scanning/temp`.
2. Unit tests with mocked HTTP client.

### Phase E — Scan UI, Review Form & Robustness

**Tickets**: E8-S4, E8-S5  
**Estimated effort**: 4–5 days  

1. Frontend: `ScanModal` (file picker → scan → review form → confirm/cancel), confidence indicators, error mapping, elapsed-time spinner.
2. Backend: extend `POST /api/attachments` to accept `source_storage_key` for temp-to-permanent promotion; add concurrency semaphore (cap 2), schema validation, auth hardening.

---

## 3. Backend Plan

### 3.1 Dashboard — Monthly Trend & Category Breakdown

#### New/updated files

| File | Action |
|---|---|
| `internal/services/dashboard_service.go` | Add `GetMonthlyTrend()` and `GetExpenseByCategory()` |
| `internal/models/models.go` | Add `MonthlyTrendItem` and `CategoryBreakdownItem` structs |
| `internal/handlers/dashboard_handler.go` | Register two new routes |

#### API contracts

**`GET /api/dashboard/monthly-trend?date_from=YYYY-MM-DD&date_to=YYYY-MM-DD`**

```json
{
  "data": [
    { "month": "2026-01", "total_income": 5000000, "total_expenses": 3200000 },
    { "month": "2026-02", "total_income": 4800000, "total_expenses": 2900000 }
  ]
}
```

Implementation: two GORM raw queries using `strftime('%Y-%m', date)` grouped by month, one for expenses and one for incomes, merged into `[]MonthlyTrendItem` in Go. Months with no data in one table get `0` via `COALESCE`.

**`GET /api/dashboard/expense-by-category?date_from=YYYY-MM-DD&date_to=YYYY-MM-DD`**

```json
{
  "data": [
    { "category_name": "Food", "total": 1500000 },
    { "category_name": "Transport", "total": 800000 }
  ]
}
```

Implementation: single GORM raw query joining `expenses` with `categories`, grouped by `categories.id`, ordered by total descending.

#### New model structs

```go
type MonthlyTrendItem struct {
    Month         string `json:"month"`
    TotalIncome   int64  `json:"total_income"`
    TotalExpenses int64  `json:"total_expenses"`
}

type CategoryBreakdownItem struct {
    CategoryName string `json:"category_name"`
    Total        int64  `json:"total"`
}
```

### 3.2 CSV Export

#### New files

| File | Action |
|---|---|
| `internal/handlers/export_handler.go` | New handler with `GET /api/export/transactions` |

No new service — the handler queries through existing `ExpenseService.List()` and `IncomeService.List()` with pagination disabled (fetch all matching rows, streamed).

#### API contract

**`GET /api/export/transactions?type=expense|income&date_from=&date_to=&category_id=`**

- Response headers:
  - `Content-Type: text/csv; charset=utf-8`
  - `Content-Disposition: attachment; filename="expenses_2026-04-27.csv"` (or `incomes_...`)
- Body: UTF-8 BOM (`\xEF\xBB\xBF`) followed by CSV using Go's `encoding/csv`.
- Columns: `date,type,category,description,amount`.
- `amount` column: raw `int64` as plain integer string. No division, no decimal point.
- Auth: JWT middleware required. Returns 401 if unauthenticated.

Implementation notes:
- The handler uses a `perPage=0` sentinel or an internal "export" path in the existing service to fetch all matching records without pagination caps. Alternatively, iterate with a large page size and stream rows as they come.
- Use `csv.NewWriter(w)` to stream directly to `http.ResponseWriter`.

### 3.3 Scanning Settings — Migration & Model

#### Migration: `009_create_scanning_settings.up.sql`

```sql
CREATE TABLE IF NOT EXISTS scanning_settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    enabled INTEGER NOT NULL DEFAULT 0,
    base_url TEXT NOT NULL DEFAULT 'http://localhost:11434/v1',
    model TEXT NOT NULL DEFAULT 'qwen3-vl:4b',
    api_key TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT OR IGNORE INTO scanning_settings (id) VALUES (1);
```

Single-row table (`id = 1` enforced by CHECK constraint). No indexes needed beyond the primary key.

#### Migration: `010_update_default_scanning_model.up.sql`

One-time data migration for installs created while the default model was still `moondream:1.8b`:

```sql
UPDATE scanning_settings
SET model = 'qwen3-vl:4b', updated_at = CURRENT_TIMESTAMP
WHERE id = 1 AND model = 'moondream:1.8b';
```

#### New model: `internal/models/scanning.go`

```go
type ScanningSettings struct {
    ID        int64     `json:"-" gorm:"primaryKey"`
    Enabled   bool      `json:"enabled" gorm:"not null;default:0"`
    BaseURL   string    `json:"base_url" gorm:"column:base_url;not null;default:'http://localhost:11434/v1'"`
    Model     string    `json:"model" gorm:"not null;default:'qwen3-vl:4b'"`
    APIKey    string    `json:"-" gorm:"column:api_key"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ScanningSettings) TableName() string { return "scanning_settings" }

type ScanResult struct {
    Vendor     string            `json:"vendor"`
    Date       string            `json:"date"`
    Currency   string            `json:"currency"`
    TotalAmount int64            `json:"total_amount"`
    LineItems  []LineItem        `json:"line_items"`
    Confidence map[string]string `json:"confidence"`
}

type LineItem struct {
    Description string `json:"description"`
    Amount      int64  `json:"amount"`
}
```

### 3.4 Scanning Service

#### New file: `internal/services/scanning_service.go`

```go
type ScanningService struct {
    db    *gorm.DB
    store storage.ObjectStore
    sem   chan struct{} // buffered channel, cap 2
}
```

**Methods**:

| Method | Signature | Notes |
|---|---|---|
| `GetSettings` | `(ctx) (*ScanningSettings, error)` | Reads row `id=1` |
| `UpdateSettings` | `(ctx, *ScanningSettings) error` | Upserts row `id=1`; validates `BaseURL` is `http`/`https` |
| `TestConnection` | `(ctx, baseURL, model, apiKey string) (bool, string)` | `GET {baseURL}/models` with 5s timeout; checks model is listed |
| `ScanImage` | `(ctx, file, filename, contentType string) (*ScanResult, string, error)` | Full scan pipeline (see below) |
| `DeleteTempScan` | `(ctx, storageKey string) error` | Validates `scan-tmp/` prefix, calls `store.Delete()` |

**`ScanImage` pipeline**:

1. Validate `contentType` ∈ `{image/jpeg, image/png, image/webp}`. Reject others → `ValidationError`.
2. Upload to ObjectStore with key `scan-tmp/{uuid}_{filename}`.
3. Read file bytes, base64-encode.
4. Build OpenAI-compatible request:
   - `model`: from settings
   - `max_tokens`: 2000
   - `messages`: one user message with text prompt + image_url content parts
5. `POST {baseURL}/chat/completions` with `context.WithTimeout(ctx, 60*time.Second)`.
6. Parse assistant content → `json.Unmarshal` into `ScanResult`. On failure → `ExtractionError`.
7. Validate: `total_amount >= 0`, `date` parses as `2006-01-02`. On failure → `ExtractionError`.
8. Return `(*ScanResult, storageKey, nil)`.

**Error on timeout**: On `context.DeadlineExceeded`, delete the temp image from ObjectStore before returning the error. Handler maps to 504.

**Concurrency**: Non-blocking `select` on `sem` (cap 2). If full → `ConcurrencyError` → handler maps to 429.

**Logging**: Use `slog`. Log `file_size_bytes` and `content_type` only. Never log base64 or raw bytes.

### 3.5 Scanning Handler

#### New file: `internal/handlers/scanning_handler.go`

| Route | Method | Handler | Notes |
|---|---|---|---|
| `GET /api/settings/scanning` | GET | `handleGetSettings` | Returns settings with `api_key_set: bool` instead of raw key |
| `PUT /api/settings/scanning` | PUT | `handleUpdateSettings` | Upserts settings; omitting `api_key` leaves it unchanged |
| `POST /api/settings/scanning/test` | POST | `handleTestConnection` | Returns `{ ok, message }` — always HTTP 200 |
| `GET /api/scanning/health` | GET | `handleHealth` | `enabled` check + `TestConnection`; always HTTP 200 |
| `POST /api/scanning/invoice` | POST | `handleScanInvoice` | Multipart upload, 10 MB limit |
| `DELETE /api/scanning/temp` | DELETE | `handleDeleteTemp` | Body `{ storage_key }`, validates prefix |

All routes registered on `protectedMux` (JWT auth via existing `AuthMiddleware`).

**Error response shape for scanning endpoints** (E8-S5):

```json
{ "error": "string_error_code", "detail": "optional human-readable detail" }
```

Error code mapping:

| Condition | HTTP | `error` |
|---|---|---|
| Scanning disabled | 503 | `scanning_disabled` |
| Bad file type | 400 | `invalid_file_type` |
| Bad storage key prefix | 400 | `invalid_storage_key` |
| Extraction parse failure | 422 | `extraction_failed` |
| Vision API timeout | 504 | `scan_timeout` |
| Concurrency limit | 429 | `too_many_scans` |

### 3.6 Attachment Handler Extension (E8-S4)

Extend `POST /api/attachments` to accept an optional `source_storage_key` form field:

- If `source_storage_key` is present **and no file uploaded**: validate prefix `scan-tmp/`, download from ObjectStore, re-upload to permanent key `attachments/{entity_type}/{entity_id}/{uuid}_{original_filename}`, delete the temp key, create DB record.
- If both `source_storage_key` and a file are provided: return 400.
- Existing file-upload path is unchanged.

### 3.7 Registration in `cmd/server/main.go`

```go
scanningService := services.NewScanningService(db, store)
scanningHandler := handlers.NewScanningHandler(scanningService)
scanningHandler.RegisterRoutes(protectedMux)

exportHandler := handlers.NewExportHandler(expenseService, incomeService, categoryService)
exportHandler.RegisterRoutes(protectedMux)
```

### 3.8 Route name discrepancy note

The user-flow doc (`invoice-scan-user-flows.md`) uses `/api/settings/scan/test-connection` and `/api/settings/scan/` while `docs/tickets.md` specifies `/api/settings/scanning`, `/api/settings/scanning/test`, etc. Per task instructions, **tickets.md routes are authoritative**. The implementation will use:

- `GET /api/settings/scanning`
- `PUT /api/settings/scanning`
- `POST /api/settings/scanning/test`
- `GET /api/scanning/health`
- `POST /api/scanning/invoice`
- `DELETE /api/scanning/temp`

---

## 4. Frontend Plan

### 4.1 New Dependencies

```bash
cd frontend && npm install recharts
```

Recharts bundles ~150 KB gzipped (one-time addition). D3 is a transitive dependency. No tree-shaking config needed — Vite handles it automatically via ESM imports.

### 4.2 Dashboard Enhancements (Phase A)

#### New components

| Component | Path | Description |
|---|---|---|
| `PeriodSelector` | `src/components/dashboard/PeriodSelector.tsx` | Preset buttons ("This Month", "Last Month", "This Year") + custom date range with two `<input type="date">` and Apply/Clear. Computes `date_from`/`date_to` and calls parent callback. |
| `MonthlyTrendChart` | `src/components/dashboard/MonthlyTrendChart.tsx` | Recharts `<BarChart>` with two `<Bar>` series (income green, expenses red), `<XAxis>` month labels, `<YAxis>` amounts (no division), `<Tooltip>`, wrapped in `<ResponsiveContainer height={300}>`. |
| `CategoryBreakdownChart` | `src/components/dashboard/CategoryBreakdownChart.tsx` | Recharts `<PieChart>` with `innerRadius` (donut), per-segment `<Cell>` colors from a 10-color palette (or category color if set), `<Legend>` with name + amount + percentage, `<Tooltip>`. |

#### Updated files

| File | Changes |
|---|---|
| `src/pages/DashboardPage.tsx` | Add `PeriodSelector` above cards; manage `dateFrom`/`dateTo` state (synced to URL query params); fetch `monthly-trend` and `expense-by-category` in parallel with `summary`; render charts in a two-column grid below cards. |
| `src/api/dashboard.ts` | Add `getMonthlyTrend(dateFrom, dateTo)` and `getExpenseByCategory(dateFrom, dateTo)`. |
| `src/types/dashboard.ts` | Add `MonthlyTrendItem` and `CategoryBreakdownItem` types. |

#### State management

- `dateFrom` and `dateTo` are derived from URL search params (`useSearchParams`).
- `PeriodSelector` updates the URL params; `DashboardPage` reacts via `useEffect` dependency on those params.
- All three API calls (`summary`, `monthly-trend`, `expense-by-category`) fire in `Promise.all` when dates change.

### 4.3 CSV Export (Phase B)

#### New component

| Component | Path | Description |
|---|---|---|
| `ExportButton` | `src/components/ExportButton.tsx` | Accepts `type` (`"expense"` or `"income"`) and current filter params. On click, constructs the full URL with auth token in a query param (or uses a hidden `<a>` with download attribute). Triggers browser download. |

#### Updated files

| File | Changes |
|---|---|
| `src/pages/ExpensesPage.tsx` | Add `<ExportButton type="expense" ... />` in the page header, passing current `dateFrom`, `dateTo`, `categoryId`. |
| `src/pages/IncomePage.tsx` | Add `<ExportButton type="income" ... />` in the page header, passing current `dateFrom`, `dateTo`, `categoryId`. |

#### Download mechanism

The export endpoint requires JWT auth. Two approaches:

1. **Preferred**: Use `fetch()` with `Authorization` header, receive the response as a Blob, create an object URL, and trigger download via a temporary `<a>` element with `download` attribute. This avoids exposing the JWT in URL params.
2. **Alternative**: Append `?token=...` to the URL and use `window.location.href`. Less secure (token in browser history/logs).

Implement option 1 in `ExportButton`.

### 4.4 Scanning Settings (Phase C)

#### New API module: `src/api/scanning.ts`

```typescript
export function getScanningSettings(): Promise<ScanningSettingsResponse>
export function updateScanningSettings(data: ScanningSettingsUpdate): Promise<ScanningSettingsResponse>
export function testScanningConnection(): Promise<{ ok: boolean; message: string }>
export function getScanningHealth(): Promise<{ ok: boolean; message: string }>
export function scanInvoice(file: File): Promise<ScanResponse>
export function deleteTempScan(storageKey: string): Promise<void>
```

#### New types: `src/types/scanning.ts`

```typescript
interface ScanningSettingsResponse {
  enabled: boolean
  base_url: string
  model: string
  api_key_set: boolean
}

interface ScanningSettingsUpdate {
  enabled: boolean
  base_url: string
  model: string
  api_key?: string
}

interface ScanResponse {
  scan_result: ScanResult
  temp_storage_key: string
}

interface ScanResult {
  vendor: string
  date: string
  currency: string
  total_amount: number
  line_items: LineItem[]
  confidence: Record<string, 'low' | 'medium' | 'high'>
}

interface LineItem {
  description: string
  amount: number
}
```

#### Updated: `src/pages/SettingsPage.tsx`

Add a "Document Scanning" section with:

- **Enable toggle**: checkbox bound to `enabled`.
- **Base URL**: text input, placeholder `http://localhost:11434/v1`.
- **Model**: text input, placeholder `qwen3-vl:4b`.
- **API Key**: password input, optional. If `api_key_set === true`, show "API key is saved — enter a new value to replace".
- **Test Connection** button → calls `POST /api/settings/scanning/test` → shows inline result badge.
- **Save** button → calls `PUT /api/settings/scanning`.

When `enabled` is false, all fields are visually disabled (greyed out) except the toggle itself.

### 4.5 Scanning Health Gate (Phase C)

#### New hook: `src/hooks/useScanningHealth.ts`

```typescript
function useScanningHealth(): {
  isHealthy: boolean
  message: string
  isLoading: boolean
}
```

- Calls `GET /api/scanning/health` on mount.
- Re-fetches on `window` `focus` event (enables: configure in Settings tab → switch back → button auto-updates).

#### Updated: `src/pages/InvoicesPage.tsx`

- Add "Scan Invoice" button next to "+ Add Invoice".
- If `isLoading`: disabled, no tooltip.
- If `!isHealthy`: disabled, grey, `title={message}`.
- If `isHealthy`: enabled, opens `ScanModal`.

### 4.6 Scan Modal & Review Form (Phase E)

#### New component: `src/components/scanning/ScanModal.tsx`

**Step 1 — File picker**:
- `<input type="file" accept="image/jpeg,image/png,image/webp">`.
- Client-side validation: file type + size ≤ 10 MB.
- On valid selection: auto-calls `scanInvoice(file)`.
- Loading state: full-modal overlay with "Scanning image…" spinner and elapsed-seconds counter.

**Step 2 — Review form** (after successful scan):
- Pre-populated editable fields: Vendor, Date, Total Amount, Currency, Notes.
- Line items table below main fields (read-only display from `line_items` array).
- **Confidence indicators**: fields where `confidence[field] === "low"` show a ⚠ icon with tooltip "Low confidence — please verify".
- "Create Invoice" button: disabled until vendor, date, and amount are non-empty.
- "Cancel" button: calls `DELETE /api/scanning/temp`, closes modal.

**Modal dismiss** (Escape, backdrop click): same cleanup as Cancel.

**On "Create Invoice"**:
1. `POST /api/invoices` with reviewed form values.
2. `POST /api/attachments` with `{ entity_type: "invoice", entity_id, source_storage_key }`.
3. On success: close modal, refresh list, toast "Invoice created from scan".
4. On failure: show inline error, keep modal open, don't delete temp image (user can retry).

**Error mapping** (E8-S5):

| HTTP status | `error` code | User message |
|---|---|---|
| 503 | `scanning_disabled` | "Document scanning is not configured. Please go to Settings to enable it." |
| 422 | `extraction_failed` | "Could not extract data from this image. Please check the image is a clear photo of a receipt or invoice, and try again." |
| 504 | `scan_timeout` | "Scanning timed out. Make sure Ollama is running and the model is loaded (`ollama run qwen3-vl:4b`)." |
| 429 | `too_many_scans` | "Another scan is in progress. Please wait a moment and try again." |
| Network error | — | "Could not reach the server. Check your connection." |

---

## 5. Cross-Cutting Concerns

### 5.1 Authentication

All new routes are registered on `protectedMux`, which is wrapped by the existing `AuthMiddleware`. No new auth logic needed. Verify during implementation that all six scanning routes and two new dashboard routes and the export route pass through middleware.

### 5.2 CORS

The CSV export uses `fetch()` + Blob download (not a direct navigation), so it goes through the existing CORS middleware. The `Content-Disposition` header must be exposed to the browser:

- Add `Content-Disposition` to `Access-Control-Expose-Headers` in `CORSMiddleware` if not already present.

### 5.3 Integer Money — CSV & Charts

- **CSV `amount` column**: write the raw `int64` value. No division, no decimal point. E.g., `150000` stays `150000`.
- **Chart Y-axis**: amounts displayed as-is from the API (already major currency units). `<YAxis>` uses a `tickFormatter` that calls the existing `formatAmount()` utility.
- The `formatAmount` utility already handles display formatting. Charts and CSV must never independently divide by 100.

### 5.4 UTF-8 BOM for CSV

The export handler writes `\xEF\xBB\xBF` as the first three bytes of the response body before any CSV content. This ensures Microsoft Excel on Windows auto-detects UTF-8 encoding for Vietnamese and other non-ASCII characters.

### 5.5 Scanning API Key Security

- `GET /api/settings/scanning` returns `api_key_set: bool`, never the raw key.
- The raw `api_key` is stored in the DB (plain text for now — acceptable for a single-user self-hosted app; can be encrypted in M4 if needed).
- If `api_key` is empty in both settings and the `OPENAI_API_KEY` env var, no `Authorization` header is sent to the vision API.

---

## 6. Testing Strategy

### 6.1 Backend Unit Tests

| Test file | Coverage |
|---|---|
| `internal/services/dashboard_service_test.go` | Extend with tests for `GetMonthlyTrend` (data across months, empty months) and `GetExpenseByCategory` (multiple categories, no data). |
| `internal/services/scanning_service_test.go` | **New.** Mock HTTP client for Ollama API. Test cases: happy path (valid JSON response), timeout (context deadline exceeded), malformed JSON, unsupported file type, `DeleteTempScan` with valid/invalid prefix. |

### 6.2 Handler Tests

The existing codebase does not have handler-level tests. Follow the same pattern: if handler tests are introduced in M3, add minimal integration tests for:

- `GET /api/export/transactions` — verify CSV headers and BOM.
- `POST /api/scanning/invoice` — verify 503 when disabled, 400 for bad file type.

### 6.3 Manual QA Checklist (aligned with flow doc edge cases)

| # | Scenario | Expected | Ticket |
|---|---|---|---|
| 1 | Ollama not running → Test Connection | Error: "Could not reach {base_url}" | E8-S1 |
| 2 | Ollama running, model not pulled → Test Connection | Error: "Model qwen3-vl:4b not found" | E8-S1 |
| 3 | Scan with clear receipt JPEG | Review form pre-filled with vendor, date, amount | E8-S2, E8-S4 |
| 4 | Scan with blurry/non-receipt image | Review form opens with blank fields + banner | E8-S2 |
| 5 | Scan timeout (>60s) | Frontend shows Ollama troubleshooting message | E8-S5 |
| 6 | Cancel after scan (before save) | Temp image deleted, no invoice created | E8-S4 |
| 7 | Escape key / backdrop click during review | Same cleanup as Cancel | E8-S4 |
| 8 | Two concurrent scans | Both succeed | E8-S5 |
| 9 | Three concurrent scans | Third gets 429 | E8-S5 |
| 10 | Settings persistence across reload | Values restored from DB | E8-S1 |
| 11 | API key env-var fallback | Scan succeeds when field empty but env var set | E8-S1 |
| 12 | Dashboard period switch to "Last Month" | All cards and charts recalculate | E6-S2 |
| 13 | Dashboard custom date range | Charts show only data in range | E6-S2 |
| 14 | Monthly trend chart with empty month | Shows 0 bar for that month | E6-S3 |
| 15 | Category breakdown with no expenses | Shows empty state message | E6-S4 |
| 16 | CSV export 50 expenses | Downloads file with 51 rows (header + data) | E6-S5 |
| 17 | CSV amount column for `150000` | Contains `150000` not `1500.00` | E6-S5 |
| 18 | CSV with Vietnamese characters in Excel | Displays correctly (BOM-based) | E6-S5 |
| 19 | Unauthenticated request to any new endpoint | Returns 401 | All |
| 20 | Scan confirms → invoice created + image attached | Attachment visible in invoice detail view | E8-S4 |

---

## 7. Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|
| **Ollama unavailable** (not installed, not running) | Scan feature unusable | Medium | Health gate disables scan button with clear guidance. Settings page Test Connection gives specific error. App functions fully without scanning. |
| **Model not pulled** (`qwen3-vl:4b` missing) | Scan returns 404 from Ollama | Medium | `TestConnection` specifically checks model list and surfaces actionable message: "Run `ollama pull qwen3-vl:4b`". |
| **Partial/malformed JSON from vision model** | Extraction fails or has missing fields | High | Backend accepts partial results (missing fields → null). Frontend renders blank fields with "please fill manually" banner. `ExtractionError` on unparseable JSON → 422 with clear message. |
| **Orphan temp files in `scan-tmp/`** | Storage leak | Medium | Cancel flow always calls `DELETE /api/scanning/temp`. On timeout/error, backend deletes temp file before returning error. For edge cases (browser crash), a future M4 cron job can purge `scan-tmp/` files older than 24h. |
| **Concurrent scans overwhelming Ollama** | Ollama OOM or extreme latency | Low | Buffered channel semaphore (cap 2) rejects excess requests with 429. Frontend disables scan button during in-flight scan. |
| **Recharts bundle size** | Frontend bundle grows | Low | One-time ~150 KB gzipped. Acceptable for a self-hosted app. Can lazy-load chart components via `React.lazy()` if needed. |
| **CSV injection** | Malicious formulas in CSV when opened in Excel | Low | Prefix cell values starting with `=`, `+`, `-`, `@` with a single quote or tab. Implement in the export handler. |
| **Vision API key in DB unencrypted** | Key exposure if DB file is stolen | Low | Single-user self-hosted app. Acceptable risk. Document that the key is stored in plaintext. Can add encryption in M4. |

---

## 8. Definition of Done

### Per-ticket checklist

**E6-S2 (Period Switcher)**:
- [ ] "This Month", "Last Month", "This Year" presets work and recalculate all dashboard data.
- [ ] Custom date range picker applies correctly.
- [ ] Period selection persisted in URL query params.
- [ ] Dashboard loads within 2 seconds for any period (NF-02).

**E6-S3 (Monthly Trend Chart)**:
- [ ] Bar chart renders with income (green) and expenses (red) bars grouped by month.
- [ ] Months with no data show as 0.
- [ ] Tooltip shows exact amounts on hover.
- [ ] Chart uses `<ResponsiveContainer>` and adapts to viewport width.

**E6-S4 (Category Breakdown Chart)**:
- [ ] Donut chart shows expense proportions by category.
- [ ] Legend shows category name, amount, and percentage.
- [ ] Empty state shown when no expenses exist for period.
- [ ] Category colors from palette (or custom color if set).

**E6-S5 (CSV Export)**:
- [ ] Export button on ExpensesPage and IncomePage (not Dashboard).
- [ ] Downloaded CSV has UTF-8 BOM.
- [ ] Columns: `date, type, category, description, amount`.
- [ ] Amount is raw integer — no division, no decimal.
- [ ] Vietnamese characters display correctly in Excel.
- [ ] Unauthenticated requests return 401.

**E8-S1 (Scanning Settings)**:
- [ ] Fresh DB returns default settings (disabled, localhost:11434/v1, qwen3-vl:4b).
- [ ] Settings persist across server restart.
- [ ] API key never returned in GET response (only `api_key_set: bool`).
- [ ] Test Connection shows result within 5 seconds.
- [ ] All settings endpoints return 401 without auth.

**E8-S2 (Scan Endpoint)**:
- [ ] Valid JPEG → 200 with `scan_result` and `temp_storage_key` starting with `scan-tmp/`.
- [ ] Scanning disabled → 503.
- [ ] Unsupported file type → 400.
- [ ] Timeout → 504 with `scan_timeout` error.
- [ ] Unparseable JSON → 422 with `extraction_failed` error.
- [ ] `DELETE /api/scanning/temp` with invalid prefix → 400.
- [ ] Unauthenticated → 401.

**E8-S3 (Health Gate)**:
- [ ] Scan button disabled when scanning disabled or unhealthy.
- [ ] Tooltip explains why button is disabled.
- [ ] Button auto-updates on window focus (Settings → enable → switch back).
- [ ] Button enabled when scanning healthy.

**E8-S4 (Scan Review Form)**:
- [ ] Review form pre-populated from scan result.
- [ ] Line items table shown (SC-07).
- [ ] Confidence indicators on low-confidence fields (SC-09).
- [ ] Edited values used in created invoice (not raw extracted values).
- [ ] Scanned image attached to invoice after save (SC-10).
- [ ] Cancel at any point → no invoice created, temp image deleted (SC-06).
- [ ] Escape/backdrop click → same cleanup.

**E8-S5 (Robustness)**:
- [ ] Missing `total_amount` in model response → 422 (not 500).
- [ ] Two concurrent scans succeed; third returns 429.
- [ ] Timeout shows specific Ollama troubleshooting message.
- [ ] User can retry after any error without closing modal.
- [ ] No base64/image bytes in server logs.
- [ ] All scanning endpoints require auth.
