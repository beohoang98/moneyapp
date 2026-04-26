# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MoneyApp is a personal finance web app — expenses, incomes, invoices/bills with file attachments and document scanning (OCR). Single-user, self-hosted.

## Development Commands

### Backend (Go)
```bash
cd backend && go run ./cmd/server          # Run server (default port 8080, override with PORT env)
cd backend && go test ./...                 # Run all tests
cd backend && go test ./internal/services/  # Run tests for a specific package
cd backend && go build ./cmd/server         # Build binary
```

### Frontend (React/TypeScript)
```bash
cd frontend && npm run dev      # Dev server
cd frontend && npm run build    # Type-check + production build
cd frontend && npm run lint     # ESLint
```

### Infrastructure
```bash
docker compose up minio -d      # Start MinIO (S3-compatible storage)
make dev-backend                 # Shortcut: run backend
make dev-frontend                # Shortcut: run frontend
```

MinIO console: `http://localhost:9001` (user: `minioadmin`, pass: `minioadmin`)

## Architecture

### Backend (`backend/`)
- **Entry point**: `cmd/server/main.go` — HTTP server using Go 1.22+ `net/http` route patterns (e.g., `GET /health`)
- **`internal/database/`** — SQLite via `mattn/go-sqlite3`. Connection opens with WAL mode and foreign keys enabled. Migrations tracked in a `migrations` table.
- **`internal/models/`** — Domain structs (Expense, Income, Invoice). All monetary amounts are `int64` in minor currency units (cents), never floats.
- **`internal/handlers/`** — HTTP handlers (route registration)
- **`internal/services/`** — Business logic layer
- **`internal/storage/`** — MinIO client wrapper (`minio-go/v7`). Auto-creates bucket on init. Provides Upload/Download/Delete operations.

### Frontend (`frontend/`)
- Vite + React 19 + TypeScript
- Source organized into `src/{components,pages,api,types}/`

### Key Design Decisions
- **Amounts as integers**: All money stored as `int64` minor units — never use `float64` for currency
- **SQLite embedded**: No external DB server. WAL mode + foreign keys enabled via connection string params
- **MinIO for files**: Receipts, invoices, attachments stored in MinIO, referenced by storage key in DB
- **Document scanning**: Uses vision LLM API (e.g., Claude Vision) server-side to extract data from receipt/invoice images — no on-device OCR

## Custom Agents

Three specialized agents are configured in `.claude/agents/`:
- **business-analyst** — Requirements analysis, feature planning, task breakdown
- **fullstack-tech-lead** — Technical implementation across Go backend and React frontend
- **qa-ux-tester** — Testing, UX review, Playwright automation, documentation review

Invoke with `@"agent-name (agent)"` in conversation.

## Documentation

- `docs/requirements.md` — Full PRD with functional/non-functional requirements, epics, milestones
- `docs/tickets.md` — Implementable story tickets broken down by milestone (M0–M4)
- `docs/tickets-review.md` — QA review of ticket breakdown with issues to address
