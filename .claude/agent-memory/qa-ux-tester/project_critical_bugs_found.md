---
name: MoneyApp Critical Issues from Ticket Review
description: Known critical and major bugs/gaps discovered during the 2026-04-26 ticket breakdown review
type: project
---

Issues confirmed in the ticket breakdown review (2026-04-26). Check if these are resolved before writing tests for affected areas.

**Critical — routing conflict:** `GET /api/invoices/summary` (E4-S8) conflicts with `GET /api/invoices/:id` (E4-S6). The `:id` handler will capture "summary" as a path param, returning a 404 or 400. Resolution: rename to `/api/invoices/stats` or move to `/api/dashboard/invoice-summary`.

**Critical — overdue logic duplication:** E4-S4 (M1) and E4-S7 (M2) both implement the same `UPDATE invoices SET status='overdue'` logic. Developer will either duplicate or skip one. The M1 ticket already fully satisfies the requirement.

**Critical — backup endpoint deferred:** NF-13 backup endpoint (`GET /api/backup`) is in M4 despite being a Should-priority reliability requirement. No data recovery mechanism exists for M1-M3.

**Major — E7-S5 missing dependencies:** Cascade delete ticket only lists E2-S6 as dependency; must also include E3-S4 and E4-S6 (income and invoice delete methods).

**Major — No unit test ticket:** NF-20 (service-layer unit tests) has no corresponding ticket. CI runs `go test ./...` on an empty test suite.

**Major — No performance test gate:** NF-01 (500ms for 10k records) and NF-04 (indexed queries) have no acceptance criteria in any ticket.

**How to apply:** When writing Playwright or integration tests, verify the routing conflict is resolved before testing the invoice summary endpoint. When writing service tests, confirm the unit test ticket exists so tests have a home.
