---
name: M1 Verification Run — 2026-04-26 findings
description: Results and defects from M1 implementation verification at commit d21fe23
type: project
---

M1 smoke verification ran on 2026-04-26, commit d21fe23. Lint + build + 55 service tests all green. Service coverage 78.1%.

**Why:** M1 is the first fully functional milestone (auth, CRUD, dashboard). This is the first post-implementation QA pass.

**How to apply:** Reference these defects when reviewing or merging M1 fixes.

## Defects found

- **D-01 (Major)** — `POST /api/expenses` missing `date` returns 400 instead of defaulting to today. `expense_service.go Create()`.
- **D-02 (Minor)** — Invalid date format in `date_from`/`date_to` query params silently returns 200 empty instead of 400. Affects expenses and incomes handlers.
- **D-03 (Major)** — `PUT /api/invoices/:id` requires `status` field or returns 400. `invoice_service.go Update()` validates status as required. Frontend edit form must always send current status.
- **D-04 (Major)** — `per_page` not capped at 100. `per_page=9999` returns all rows. Applies to expenses, incomes, invoices handlers.
- **D-05 (Minor)** — `GET /api/dashboard/summary` always returns `unpaid_amount: 0` and `overdue_amount: 0` (hardcoded). `InvoiceStats` model has no per-status amount fields; `dashboard_service.go GetSummary()` hardcodes them.

## Test result summary

Pass: 43 curl/API test cases | Fail: 5 defects | Blocked: ~40 (browser/UI/perf)

## Gotchas

- Running binary was stale (built 2.5h before test) — always kill and restart `go run ./cmd/server` before verification.
- `.env` file is NOT auto-loaded by the server; env vars must be exported in shell or `.env` must be sourced manually. JWT_SECRET defaults to insecure dev value if not set.
