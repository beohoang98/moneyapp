# Milestone 2 — File Attachments & Enhanced Lists: Status Report

**Last updated**: 2026-04-27  
**Overall status**: ✅ Substantially complete — 13 of 16 tickets fully done; 3 tickets partial (per-row attachment indicator missing across expense, income, and invoice list views)

---

## Executive Summary

All sixteen Milestone 2 tickets have been implemented and the core feature set is functional. The file attachment infrastructure (E7-S1 through E7-S5) is complete end-to-end: the `attachments` table (migration 008) is in place, uploads are validated for MIME type and size, images render as inline thumbnails, files download through the authenticated proxy endpoint, and cascade delete is wired into all three entity services. Custom categories (E5-S2 through E5-S4) are fully operational, including rename/delete protection for default categories and transaction reassignment to "Uncategorized" on category removal. Enhanced list features—expense sorting (M2-01), income category filter (M2-02), invoice date-range filter (M2-03), and income running total (M2-04)—are all implemented in both backend and frontend. The overdue-check admin endpoint (E4-S7) is reachable and authenticated.

Three entity-attachment integration tickets (E2-S7, E3-S5, E4-S5) are marked Partial because the per-row attachment indicator (paperclip icon / count) called for in their acceptance criteria is not implemented: the list views present a generic "Files" button on every row regardless of whether any attachments exist. This is a UX completeness gap, not a data integrity issue. QA automated one Playwright test (1/1 passed) covering login, expenses, attachment upload (JPEG fixture confirmed), invoices, and categories. File download and thumbnail flows were exercised only manually (browser smoke, not automated), and the MinIO (S3) storage path was not exercised in this round. **Follow-up (post–status doc):** dev CORS now allows both `http://localhost:5173` and `http://127.0.0.1:5173` via `CORS_ALLOWED_ORIGINS` (see `backend/internal/config/config.go`); attachment modals dismiss on **Escape** on expense, income, and invoice pages. The `CategoryForm` component was inlined into `CategoriesPage.tsx` rather than extracted to a separate file, a minor deviation from the ticket spec.

---

## Tech Lead — M2 Implementation

Evidence gathered by inspecting file presence, handler routes, service method signatures, and frontend component imports. No automated test run was performed as part of this status review; QA findings (documented separately) provide runtime evidence.

| Ticket | Title | Status | Evidence |
|--------|-------|--------|----------|
| **E7-S1** | File Upload Infrastructure | ✅ Done | `backend/migrations/008_create_attachments.up.sql` — schema with CHECK constraints on `entity_type` and `mime_type`, UNIQUE on `storage_key`, index `idx_attachments_entity`. `backend/internal/models/attachment.go`. `backend/internal/services/attachment_service.go` — `Upload`, `ListByEntity`, `Delete`, `DeleteByEntity`. `backend/internal/handlers/attachment_handler.go` — `POST /api/attachments`, `GET /api/attachments`, `DELETE /api/attachments/{id}`. `frontend/src/api/attachments.ts`. `frontend/src/components/attachments/FileUpload.tsx` and `AttachmentList.tsx`. |
| **E7-S2** | File Size Validation | ✅ Done | `attachment_handler.go` — `http.MaxBytesReader` applied before parsing multipart. `attachment_service.go` — server-side `size_bytes` check before upload. `FileUpload.tsx` — client-side `file.size` guard with error message before HTTP request. |
| **E7-S3** | File Download | ✅ Done | `attachment_handler.go` line 25 — `GET /api/attachments/{id}/download`; handler streams file from storage with `Content-Disposition: attachment`. `AttachmentList.tsx` — Download button triggers the endpoint. |
| **E7-S4** | Image Thumbnail Preview | ✅ Done | `attachment_handler.go` line 26 — `GET /api/attachments/{id}/preview` with `Content-Disposition: inline`. `AttachmentList.tsx` — renders `<img>` for JPEG/PNG at thumbnail size; PDF icon for `application/pdf`; click expands to larger preview. |
| **E7-S5** | Cascade Delete Attachments | ✅ Done | `expense_service.go` line 115, `income_service.go` line 110, `invoice_service.go` line 129 — all three `Delete()` methods call `attachmentService.DeleteByEntity(ctx, entityType, id)` before removing the parent record. Storage failures are logged but do not block DB deletion. |
| **E2-S7** | Attach Receipts to Expenses | ⚠️ Partial | `ExpensesPage.tsx` imports and renders `FileUpload` + `AttachmentList` in the expense detail modal. Upload and list work (Playwright-confirmed, JPEG fixture uploaded). **Gap**: expense list rows show a generic "Files" button on every row (line 194 of `ExpensesPage.tsx`) — no per-row paperclip indicator or count distinguishing records with vs. without attachments (AC2 not met). |
| **E3-S5** | Attach Documents to Income | ⚠️ Partial | `IncomePage.tsx` imports and renders `FileUpload` + `AttachmentList` in the income edit modal. Upload path is functional. **Gap**: income list rows have no per-row attachment indicator (same pattern as E2-S7 — AC2 not met). |
| **E4-S5** | Attach Invoice PDFs | ⚠️ Partial | `InvoicesPage.tsx` imports and renders `FileUpload` + `AttachmentList` in the invoice detail modal. `AttachmentList` handles PDF vs. image rendering. **Gap**: invoice list rows have no per-row attachment indicator (AC2 not met). |
| **E4-S7** | Overdue Check Admin Endpoint | ✅ Done | `invoice_handler.go` line 21 — `POST /api/invoices/check-overdue` registered. `handleCheckOverdue` (line 252) calls `invoiceService.UpdateOverdueStatuses(ctx)` and returns `{"updated_count": N}`. Endpoint is authenticated (JWT middleware applied at mux level). |
| **E5-S2** | Create Custom Categories | ✅ Done | `category_handler.go` — `POST /api/categories` with name/type validation. `category_service.go` — `Create()` enforces uniqueness and `is_default = false`. `CategoriesPage.tsx` — two sections (Expense / Income), distinct CTAs **“Add expense category”** / **“Add income category”** and inline form. Note: `CategoryForm` is inlined in `CategoriesPage.tsx` rather than extracted (see Minor deviations). |
| **E5-S3** | Rename and Delete Custom Categories | ✅ Done | `category_handler.go` — `PUT /api/categories/{id}` and `DELETE /api/categories/{id}`. `category_service.go` — `Update()` and `Delete()` reject requests on default categories (403). `CategoriesPage.tsx` — default categories display lock icon; only custom categories show Edit/Delete buttons and confirmation dialog. |
| **E5-S4** | Reassign Transactions on Category Delete | ✅ Done | `category_service.go` lines 112–127 — `Delete()` opens a transaction, locates the "Uncategorized" category for the same type, runs `UPDATE expenses`/`UPDATE incomes` reassignment, then deletes the category record. `is_default = 1` guard prevents "Uncategorized" from being deleted. |
| **M2-01** | Expense List Sorting | ✅ Done | `expense_service.go` — `SortBy`/`SortOrder` fields in list params; whitelisted column switch at line 202; default `date DESC`. `expense_handler.go` reads `sort_by`/`sort_order` query params. `ExpensesPage.tsx` — clickable column headers toggle sort, arrow indicator, URL-persisted state. |
| **M2-02** | Income Filter by Category | ✅ Done | `IncomePage.tsx` line 6 — imports and uses `CategoryFilter` component for income. `category_id` URL param wired to `getIncomes()` call. Backend `income_service.go` `List()` already accepted `category_id` from M1. |
| **M2-03** | Invoice Date Range Filter | ✅ Done | `invoice_handler.go` lines 200–225 — `date_field` validated (`issue_date`\|`due_date`), `date_from`/`date_to` accepted and validated as ISO dates, ordered correctly. `InvoicesPage.tsx` uses `DateRangeFilter` with a "Due Date / Issue Date" combobox toggle; state URL-persisted. |
| **M2-04** | Income Running Total | ✅ Done | `IncomePage.tsx` — `totalAmount` state (line 19) populated from `result.total_amount` (line 50); displayed in `.summary-bar` (line 131–134) via `formatAmount()`. Backend `total_amount` field already present in income list response from M1 (E3-S1). |

---

## Minor Deviations / Notes

1. **E2-S7 / E3-S5 / E4-S5 — Per-row attachment indicator**: Tickets specify "an attachment indicator (e.g., paperclip icon with count)" in list views as AC2. Current implementation shows a "Files" button on every row unconditionally. The button opens the attachment modal correctly, but provides no visual scan signal. **QA UX — M2** in `qa-result.md` rates this Medium severity. Resolution requires fetching attachment count per entity in the list query or loading counts lazily on render.

2. **E5-S2 — `CategoryForm.tsx` not extracted**: Ticket spec calls for a dedicated `src/components/categories/CategoryForm.tsx` component. The form is instead rendered inline within `CategoriesPage.tsx`. Functionally identical; no user-facing impact. Can be extracted in M3 polish if desired.

3. **E4-S5 — Inline PDF viewer**: The ticket spec suggests an `<iframe>` or `<embed>` for in-page PDF preview. Current behavior opens the PDF in a new browser tab via the `/preview` endpoint with `Content-Disposition: inline`. QA noted this causes tab sprawl (rated Low). Consistent with a pragmatic single-user app trade-off; document as intended if no change is planned.

4. **CORS — dev hostnames (resolved)**: Backend `CORSMiddleware` now uses an allowlist from `CORS_ALLOWED_ORIGINS` (default includes both `http://localhost:5173` and `http://127.0.0.1:5173`). Historical Finding **C-01** in `qa-result.md` described the prior single-origin behavior.

---

## QA — Test Planning

**Document**: [`docs/test-plans/milestone-2.md`](../test-plans/milestone-2.md)  
**Status**: Draft (dated 2026-04-26) — full TS-25–TS-40 execution not completed in this QA pass.

| Ticket | Test Suite | Cases | AC Bullets Covered |
|--------|------------|-------|--------------------|
| E7-S1 | TS-25 | 12 | AC1 (local), AC2 (s3), AC3 (type reject), AC4 (cascade note), AC5 (orphan) |
| E7-S2 | TS-26 | 5 | AC1 (client), AC2 (server 413) |
| E7-S3 | TS-27 | 5 | AC1 (download), AC2 (404) |
| E7-S4 | TS-28 | 5 | AC1 (JPEG thumb), AC2 (PDF icon), AC3 (click expand) |
| E7-S5 | TS-29 | 6 | AC1 (cascade), AC2 (storage down) |
| E2-S7 | TS-30 | 5 | AC1 (upload), AC2 (indicator) |
| E3-S5 | TS-31 | 4 | AC1 (upload), AC2 (indicator) |
| E4-S5 | TS-32 | 4 | AC1 (upload+preview), AC2 (click open) |
| E4-S7 | TS-33 | 4 | AC1 (updated_count), AC2 (zero count) |
| E5-S2 | TS-34 | 6 | AC1 (create), AC2 (409 duplicate), AC3 (is_default=false) |
| E5-S3 | TS-35 | 7 | AC1 (rename), AC2 (403 default), AC3 (confirm dialog) |
| E5-S4 | TS-36 | 5 | AC1 (reassign), AC2 (view uncategorized), AC3 (no data loss) |
| M2-01 | TS-37 | 5 | AC1 (sort desc), AC2 (toggle asc) |
| M2-02 | TS-38 | 4 | AC1 (filter), AC2 (clear) |
| M2-03 | TS-39 | 5 | AC1 (issue_date), AC2 (due_date) |
| M2-04 | TS-40 | 3 | AC1 (total displayed) |
| **Total** | | **85** | All AC bullets across all 16 tickets |

The plan covers: two storage environment matrices (`m2-local` with `STORAGE_TYPE=local`; `m2-minio` with `STORAGE_TYPE=s3`), a traceability matrix, cross-cutting auth checks for all new M2 endpoints, integer-amount regression checks, category "Uncategorized" invariant checks, responsive layout checks (NF-16), and an automation backlog of Playwright and Go integration test candidates.

**Documented gaps in the test plan** (noted by QA, not blocking sign-off):

| Gap | Ticket | Detail |
|-----|--------|--------|
| MIME spoofing (magic-byte sniff) | E7-S1 | Server checks declared `Content-Type`, not file magic bytes. Known limitation. |
| `entity_id` existence validation | E7-S1 | Behavior for non-existent `entity_id` in upload not specified; test case to be added once confirmed. |
| Upload progress for large files | E7-S1, E2-S7 | Requires manual testing with throttled network; Playwright automation cannot easily simulate. |
| `sort_by` SQL injection whitelist | M2-01 | TS-37-05 covers this; confirm error response does not leak SQL column name. |
| `date_field` default (omitted param) | M2-03 | TS-39-03 leaves expected default open — confirm with developer. |
| M1 carryover defects (D-01, D-03, D-04) | M1 | Must be resolved before M2 sign-off; cascade delete depends on clean entity delete behavior. |

---

## QA — Results

**Document**: [`docs/milestone-02/qa-result.md`](./qa-result.md)  
**Date**: 2026-04-27  
**Method**: Cursor IDE Browser MCP (manual smoke) + Playwright automated re-test (`e2e/m2-retest.spec.ts`)  
**Environment**: `STORAGE_TYPE=local`, backend `:8080`, frontend Vite dev `:5173`, seed user `admin`/`changeme`

**Overall posture**: Core flows pass. Critical and Should-priority features are functional.

| Area | Result |
|------|--------|
| Login → dashboard | ✅ Pass |
| Expenses list + filters shell | ✅ Pass |
| Expense attachment upload (JPEG `testing/files/test-invoice.jpeg`) | ✅ Pass (Playwright-automated) |
| Invoices list + date filter UI | ✅ Pass |
| Categories page structure | ✅ Pass |
| Income list + running total | ✅ Pass (manual smoke) |
| File download / thumbnail / lightbox | Not fully automated — manual smoke only |
| `POST /api/invoices/check-overdue` | Not run — API-only; curl test required per test plan |
| MinIO (S3) storage path | Not exercised in this QA round |

**Playwright automated run**: `e2e/m2-retest.spec.ts` — **1 test, 1 passed** (~1.9 s). Three minor test-only fixes were applied (`package.json` ESM flag, `exact: true` for strict heading matcher, `.first()` on duplicate filename locator after re-upload). Subsequent runs may seed an expense via API before opening **Files** (hermetic clean DB).

**Known open findings**:

| ID | Severity | Summary |
|----|----------|---------|
| (UX) | Medium | Per-row attachment indicator missing on expense, income, and invoice list rows — "Files" button appears on every row regardless of attachment count. Matches E2-S7 / E3-S5 / E4-S5 AC2 gap noted above. |

Historical manual findings **C-01** (CORS) and **C-02** (Escape on attachment modal) are **addressed in code**; see `docs/milestone-02/qa-result.md` for the original smoke notes.

**Screenshot policy**: Playwright writes under **`e2e/test-results/screenshots/`** (ignored via `e2e/.gitignore`). Root **`.gitignore`** also ignores **`docs/milestone-02/screenshots/`** for any legacy or manual captures there.

---

## Risks / Follow-ups for M3

M3 (Document Scanning) depends directly on M2 being complete: the scanning flow reuses the M2 file attachment upload and storage infrastructure (`ObjectStore`, `AttachmentService`, migration 008). Before starting M3, the three outstanding partial tickets (E2-S7, E3-S5, E4-S5 per-row indicators) should either be closed or formally deferred, and the M1 carryover defects (D-01, D-03, D-04) should be resolved if they still apply. Additionally, M3 requires a vision API key decision (OQ7 in `requirements.md`) and env documentation for `VISION_API_KEY` / `VISION_API_PROVIDER` before scanning sprint work can begin. The MinIO (S3) attachment path should also be exercised in a full QA pass before M3, since scanning uploads will exercise the same `STORAGE_TYPE=s3` code path in production-like deployments.
