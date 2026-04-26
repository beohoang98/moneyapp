# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MoneyApp is a personal finance web app — expenses, incomes, invoices/bills with file attachments and document scanning (OCR). Single-user, self-hosted.

## Development Commands

### Backend (Go)
SQLite uses CGO (`mattn/go-sqlite3`). If `go run` fails opening the DB, use `CGO_ENABLED=1` (default on most macOS/Linux installs when a C toolchain is present).
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
docker compose up minio -d      # Start MinIO (needed only when STORAGE_TYPE=s3)
make dev-backend                 # Shortcut: run backend
make dev-frontend                # Shortcut: run frontend
```

MinIO console: `http://localhost:9001` (user: `minioadmin`, pass: `minioadmin`)

> **Local dev without MinIO**: set `STORAGE_TYPE=local` and `LOCAL_STORAGE_PATH=./uploads` in `.env`. The server creates the directory on startup. No Docker required for basic development.

## Architecture

### Backend (`backend/`)
- **Entry point**: `cmd/server/main.go` — HTTP server using Go 1.22+ `net/http` route patterns (e.g., `GET /health`)
- **`internal/database/`** — SQLite via `mattn/go-sqlite3`. Connection opens with WAL mode and foreign keys enabled. Migrations tracked in a `migrations` table.
- **`internal/models/`** — Domain structs (Expense, Income, Invoice). All monetary amounts are `int64` in minor currency units (cents), never floats.
- **`internal/handlers/`** — HTTP handlers (route registration)
- **`internal/services/`** — Business logic layer
- **`internal/storage/`** — Configurable storage backend. `ObjectStore` in `storage.go` defines Upload/Download/Delete/HealthCheck; `LocalStorage` (`local.go`) and `MinIOStorage` (`minio.go`, S3-compatible) implement it. `cmd/server/main.go` selects the implementation from `STORAGE_TYPE` (`local` \| `s3`). See `.env.example` for all variables.

### Frontend (`frontend/`)
- Vite + React 19 + TypeScript
- Source organized into `src/{components,pages,api,types}/`

### Key Design Decisions
- **Amounts as integers**: All money stored as `int64` minor units — never use `float64` for currency
- **SQLite embedded**: No external DB server. WAL mode + foreign keys enabled via connection string params
- **Configurable storage**: Receipts, invoices, attachments stored via `ObjectStore` — local filesystem (`STORAGE_TYPE=local`) or S3-compatible MinIO (`STORAGE_TYPE=s3`). Always referenced by an opaque `storage_key` in the DB.
- **Document scanning**: Uses vision LLM API (e.g., Claude Vision) server-side to extract data from receipt/invoice images — no on-device OCR

## Custom Agents

Four specialized agents are configured in `.claude/agents/`:
- **business-analyst** — Requirements analysis, feature planning, task breakdown
- **fullstack-tech-lead** — Technical implementation and architecture across Go backend and React frontend
- **code-reviewer** — Independent pre-merge review: same stack context as tech-lead, adversarial stance (risk, AC alignment, security, maintainability); default is review-only, not implementation
- **qa-ux-tester** — Testing, UX review, Playwright automation, documentation review

Invoke with `@"agent-name (agent)"` in conversation.

## Documentation

- `docs/requirements.md` — Full PRD with functional/non-functional requirements, epics, milestones
- `docs/tickets.md` — Implementable story tickets broken down by milestone (M0–M4)
- `docs/tickets-review.md` — QA review of ticket breakdown with issues to address
