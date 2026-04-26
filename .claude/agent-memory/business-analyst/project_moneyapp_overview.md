---
name: MoneyApp Project Overview
description: Core project context — goals, tech stack, scope boundaries, and key architectural decisions for MoneyApp
type: project
---

MoneyApp is a self-hosted personal finance web app (single user) for tracking expenses, incomes, and invoices/bills with file attachments.

**Tech Stack**: React + TypeScript (frontend), Go REST API (backend), SQLite (database), MinIO (file storage via Docker Compose).

**MVP scope**: Auth, expense CRUD, income CRUD, invoice CRUD, default categories, dashboard summary cards.

**Explicitly deferred**: Multi-user, bank sync, mobile native, multi-currency, automated recurring transactions, budget goals, tax exports, on-device OCR (Tesseract), batch scanning.

**Document Scanning feature added (v1.1)**: Epic 8 added to requirements. Server-side vision LLM API (Claude Vision preferred) handles receipt/invoice OCR. Core flow: user uploads image → backend calls vision API → structured JSON returned → pre-filled review form → user confirms → record + attachment saved. Depends on Epic 7 (file attachments). Planned for Milestone 3. Key open questions: OQ7 (which vision API), OQ8 (persist scan_results table or stateless), OQ9 (image lifecycle if user discards scan).

**Why:** Solo/small-team project; scoped to avoid over-engineering while delivering core personal finance value.

**How to apply:** Proposals should respect the deferred list; flag any suggestion that pulls in multi-user or bank-sync complexity as out of scope.

Key architectural decisions recorded in requirements.md:
- Monetary amounts stored as integers (minor currency units) to avoid float errors — NF-10
- MinIO object keys never publicly accessible; use pre-signed URLs or proxy — NF-07
- Passwords hashed with bcrypt cost >= 12 — NF-08
- DB migrations via versioned files (e.g., golang-migrate) — NF-19
- SQLite WAL mode recommended for performance — D2 note in risks
- MinIO key only written to DB after successful upload to avoid orphans — D7

**Docs location**: `/Users/beohoang98/0_projects/learn-go/moneyapp/docs/requirements.md`
