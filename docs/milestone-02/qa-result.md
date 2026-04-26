# Milestone 02 — Browser QA result

**Date:** 2026-04-27  
**Status:** Draft (smoke / exploratory)  
**Method:** Cursor IDE Browser MCP (`browser_navigate`, `browser_snapshot`, `browser_fill`, `browser_click`, `browser_wait_for`, `browser_search`, `browser_console_messages`, `browser_network_requests`)  
**Plan reference:** `docs/test-plans/milestone-2.md` (full TS coverage not executed in this pass)

## Test fixtures

| Fixture | Use |
|---------|-----|
| **`testing/files/test-invoice.jpeg`** | Canonical **JPEG** for invoice image upload, expense receipt, and thumbnail/preview checks (real Vietnamese retail receipt; ~2.1 MB, under 10 MB limit). Use in the browser file picker, Playwright `setInputFiles`, or `curl -F file=@testing/files/test-invoice.jpeg` from repo root. |

## Environment

| Item | Value |
|------|--------|
| Frontend | Vite dev server, `http://localhost:5173` (see CORS note below) |
| Backend | `CGO_ENABLED=1 go run ./cmd/server`, default `http://localhost:8080` |
| API base | `http://localhost:8080/api` (frontend default) |
| Storage | `STORAGE_TYPE=local` (from server log) |
| Auth | Seed user `admin` / `changeme` |

## Summary

| Area | Result |
|------|--------|
| Login → dashboard | **Pass** when UI origin is `http://localhost:5173` |
| Login from `http://127.0.0.1:5173` | **Blocked** — browser CORS preflight failed (see Finding C-01) |
| Expenses list + filters shell | **Pass** — list, category combobox, pagination controls present |
| Expense “Files” → attachment panel | **Pass** — modal title “Expense Attachments” and expense summary line rendered |
| Categories page | **Pass** — headings “Categories”, “Expense Categories”, “Income Categories”, “+ Add Category” |
| Invoices list + date filter UI | **Pass** — status chips; combobox “Due Date” / “Issue Date”; From/To date fields |
| Income list + running total | **Pass** — “Total” summary and table rows (Salary) visible; total matches visible rows in spot check |
| File upload / download / thumbnail | **Not run** — no automated file chooser in this session; re-run using **`testing/files/test-invoice.jpeg`** per [Test fixtures](#test-fixtures) |
| `POST /api/invoices/check-overdue` | **Not run** — API-only; use curl per test plan |

## Findings

### C-01 — CORS origin mismatch (Major / dev ergonomics)

**Observed:** With the UI at `http://127.0.0.1:5173`, `POST http://localhost:8080/api/auth/login` failed in the browser: preflight response `Access-Control-Allow-Origin` was `http://localhost:5173`, which does not match the page origin `http://127.0.0.1:5173`.

**Evidence:** Console message: *“The 'Access-Control-Allow-Origin' header has a value 'http://localhost:5173' that is not equal to the supplied origin.”*

**Workaround used for testing:** Open the app at `http://localhost:5173` so origin matches backend CORS allowlist.

**Suggestion:** Allow both origins in dev, or document “use localhost hostname only” in `CLAUDE.md` / test plans.

### C-02 — Attachment modal dismiss (Minor / UX)

**Observed:** After opening “Expense Attachments”, `Escape` did not remove the overlay in one attempt; direct navigation to `/categories` cleared the route and overlay.

**Suggestion:** Confirm expected keyboard / backdrop-close behavior and add to test plan if intentional.

---

## QA UX — M2

**Source:** QA synthesis from browser smoke, Playwright re-test, ticket AC cross-check (`docs/test-plans/milestone-2.md`), and attachment / list / category flows.  
**Date:** 2026-04-27  
**Status:** Draft — recommendations for product polish (not automated pass/fail unless linked to findings above).

### Strengths

- **File upload** — “Drop file here or click to upload”, accepted types, and 10 MB hint are easy to parse; inline validation errors are visible without hunting in the console.
- **Income total** — Filter context + “Total” reads clearly and matches the expense-list pattern (M2-04 intent).
- **Invoices** — Status chips plus **Due date / Issue date** dimension for the date range is understandable once the control is noticed.
- **Attachments (post thumbnail fix)** — Seeing the actual image in the list matches user expectations after upload.

### Friction / recommendations

| Topic | Severity | Notes |
|-------|----------|--------|
| **Dev hostname (C-01)** | High (dev UX) | `127.0.0.1` vs `localhost` for the UI → login fails with CORS in the console; feels like a broken app. Prefer **both origins allowed in dev** or a **single clear doc + optional in-app hint** when login fails for network/CORS. |
| **Attachment modal (C-02)** | Medium | **Escape** / **click-outside** to dismiss was inconsistent in smoke testing. Align with other modals and document; reduces trap feeling. |
| **“Files” on every row** | Medium | No **per-row signal** which records have attachments (tickets suggested indicator/count). Table feels busy without scan benefit; consider paperclip + count or de-emphasizing when count is 0. |
| **Categories — duplicate CTA labels** | Low | Two identical **“+ Add Category”** buttons (expense vs income sections) add cognitive load; differentiate (“Add expense category” / “Add income category”) or anchor each button visually to its section. |
| **PDF opens new tab** | Low | Repeated “open preview” can **sprawl tabs**; consider in-modal preview or explicit Open vs Download. |
| **Delete attachment errors** | Medium | Failed delete should not leave an **ambiguous modal state**; pair with user-visible error (e.g. toast) per code-review theme. |

### Accessibility & mobile (brief)

- Keep table action targets (**Files** / **Edit** / **Delete**) at a comfortable **touch size** on narrow viewports (see NF-16 in test plan).
- When attachment **lightbox** behavior is finalized, add **Escape to close** and **focus management** for keyboard and screen-reader users.

### UX summary

M2 is **coherent and usable** for a single-user finance app: uploads and list enhancements fit the product. The largest **felt-quality** wins are **dev CORS ergonomics**, **predictable modal dismissal**, and **attachment discoverability** on list rows.

---

## Traceability (smoke ↔ M2 themes)

| M2 theme | This session |
|----------|----------------|
| E7 / attachments entry from lists | Exercised “Files” → modal |
| E5 / categories UI | Page load and structure only |
| M2-03 invoice date field toggle | Combobox present |
| M2-04 income total | “Total” visible with amount |
| E7-S2 / E7-S3 / E7-S4 (size, download, preview) | Not exercised |

---

## Playwright re-test — 2026-04-27

**Method:** Automated Playwright e2e (`npm test` from `e2e/`)  
**Spec:** `e2e/m2-retest.spec.ts`  
**Browsers:** Chromium (Desktop Chrome)  
**Command:** `cd e2e && npm install && npx playwright install chromium && npm test`  
**Result:** ✅ **1 passed** (1 test, ~1.9 s total)  
**Servers:** Reused existing — backend on `http://localhost:8080` (`STORAGE_TYPE=local`), frontend Vite dev server on `http://localhost:5173`

### Fixes applied (test-only, no product code changed)

| File | Change | Reason |
|------|--------|--------|
| `e2e/package.json` | Added `"type": "module"` | `e2e/playwright.config.ts` and spec use `import.meta.url` (ESM); without this flag Node treated the files as CJS, throwing `exports is not defined` |
| `e2e/m2-retest.spec.ts` line 46 | Added `exact: true` to `getByRole('heading', { name: 'Categories' })` | Strict mode violation — three headings matched "Categories" (`<h1>Categories</h1>`, `<h2>Expense Categories</h2>`, `<h2>Income Categories</h2>`); `exact: true` targets only the page `<h1>` |
| `e2e/m2-retest.spec.ts` (post-run) | `getByText('test-invoice.jpeg').first()` after upload | Re-uploads or duplicate rows can surface two identical filename nodes; strict `toBeVisible` on the locator alone fails |

### Screenshots produced

All 7 screenshots written to `docs/milestone-02/screenshots/`:

| File | Description |
|------|-------------|
| [`docs/milestone-02/screenshots/01-login.png`](screenshots/01-login.png) | Login page before credentials entered |
| [`docs/milestone-02/screenshots/02-dashboard.png`](screenshots/02-dashboard.png) | Dashboard after successful login as `admin` |
| [`docs/milestone-02/screenshots/03-expenses.png`](screenshots/03-expenses.png) | Expenses list page |
| [`docs/milestone-02/screenshots/04-attachments-modal-before-upload.png`](screenshots/04-attachments-modal-before-upload.png) | Attachment modal open on first expense (before upload) |
| [`docs/milestone-02/screenshots/05-attachments-after-upload.png`](screenshots/05-attachments-after-upload.png) | Attachment modal after `test-invoice.jpeg` uploaded — filename and "Attachments (N)" badge visible |
| [`docs/milestone-02/screenshots/06-invoices.png`](screenshots/06-invoices.png) | Invoices list page |
| [`docs/milestone-02/screenshots/07-categories.png`](screenshots/07-categories.png) | Categories page showing "Expense Categories" and "Income Categories" sections |

### Coverage vs. previous manual pass

| Area | Manual (2026-04-27) | Playwright re-test |
|------|-------------------|-------------------|
| Login → dashboard | Pass | ✅ Pass (automated) |
| Expenses list | Pass | ✅ Pass |
| Expense attachment upload (JPEG) | Not run | ✅ Pass — `test-invoice.jpeg` uploaded via `setInputFiles`; filename and badge confirmed |
| Invoices list | Pass | ✅ Pass |
| Categories page | Pass | ✅ Pass |
| Income list | Pass | Not covered (out of scope for this spec) |
| CORS with `127.0.0.1` origin | Blocked (C-01) | N/A — test uses `localhost:5173` matching backend allowlist |

---

---

## QA re-test — 2026-04-27 (re-run)

**Command:** `cd e2e && npm install && npx playwright install chromium && npm test`  
**Playwright outcome:** ✅ **1 passed** (1 test, ~3.0 s total) — `m2-retest.spec.ts` › "login, core pages, expense JPEG upload"  
**Servers:** Reused existing — backend on `http://localhost:8080`, frontend Vite dev server on `http://localhost:5173`  
**Browsers:** Chromium (Desktop Chrome)  
**Fixes already in codebase:** CORS origin normalised and Escape / backdrop dismiss for attachment modal were addressed in a prior session; no product code changes were needed for this re-run.  
**Screenshots:** `e2e/test-results/screenshots/` (7 files: login, dashboard, expenses, attachment-before-upload, attachment-after-upload, invoices, categories)

---

## Follow-up

1. Re-run (or extend) against `docs/test-plans/milestone-2.md` TS-25–TS-40 with scripted uploads and MinIO profile if required.  
2. Resolve or document C-01 before external QA.  
3. Optional: have QA automation (e.g. Playwright) assert login + one attachment round-trip with a fixture file.
