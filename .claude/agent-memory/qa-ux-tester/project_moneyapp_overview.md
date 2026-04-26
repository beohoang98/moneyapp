---
name: MoneyApp Project Overview
description: Tech stack, architecture, scope, and milestone structure for MoneyApp personal finance app
type: project
---

MoneyApp is a single-user personal finance web app: Go backend, React/TypeScript frontend (Vite), SQLite database, MinIO object storage.

**Why:** Solo developer or small team; self-hosted; no multi-tenancy in MVP.

**Tech decisions relevant to QA:**
- Amounts stored as integers (minor currency units, e.g., cents/pips) — no float amounts anywhere
- Polymorphic attachments table: `entity_type` + `entity_id` pattern (not entity-scoped routes)
- JWT authentication, stateless logout (client-side token clear only)
- WAL mode for SQLite; foreign keys enabled
- Single configured currency (VND or USD from env var `CURRENCY`)

**Milestone structure:**
- M0: Infrastructure skeleton (health, migrations, CI, Docker, API client)
- M1: Auth + CRUD for expenses/income/invoices + dashboard summary cards (no files, no charts)
- M2: File attachments (MinIO), custom categories, invoice auto-overdue
- M3: Charts (Recharts), CSV export, dashboard period switcher
- M4: Polish — password change, backup endpoint, tags, PDF export, notifications

**How to apply:** When reviewing tickets or writing tests, always check which milestone a feature lives in — M1 has no attachment support, so delete flows in M1 should not cascade to MinIO.
