# Test Plan — Milestone 0: Project Foundation

**Milestone goal**: Runnable skeleton with infrastructure in place. No user-facing features yet.  
**Exit criteria (from tickets.md)**: `go run ./cmd/server` and `npm run dev` start without errors; MinIO is reachable; health endpoint returns 200; migrations run on startup; frontend renders a shell with routing.  
**NF references covered**: NF-09, NF-16, NF-18, NF-19, NF-21  
**Date**: 2026-04-26  
**Status**: Draft

---

## Table of Contents

1. [Environment Setup](#environment-setup)
2. [Traceability Matrix](#traceability-matrix)
3. [M0-01 — Database Migration Runner](#m0-01--database-migration-runner)
4. [M0-02 — Environment Configuration](#m0-02--environment-configuration)
5. [M0-03 — API Error Handling & Middleware](#m0-03--api-error-handling--middleware)
6. [M0-04 — Frontend Routing & App Shell](#m0-04--frontend-routing--app-shell)
7. [M0-05 — Frontend API Client](#m0-05--frontend-api-client)
8. [M0-06 — Health Check Endpoint](#m0-06--health-check-endpoint)
9. [M0-07 — Docker Compose Enhancement](#m0-07--docker-compose-enhancement)
10. [M0-08 — CI Pipeline](#m0-08--ci-pipeline)
11. [Cross-Cutting Checks](#cross-cutting-checks)
12. [Automation Backlog](#automation-backlog)

---

## Environment Setup

### Required environments

| Environment | Description | Notes |
|---|---|---|
| **local-fresh** | Fresh checkout, no `.env`, no DB file | Verifies defaults and first-run behavior |
| **local-configured** | `.env` copied from `.env.example`, MinIO running | Standard dev environment |
| **local-minio-down** | MinIO stopped (`docker compose stop minio`) | Tests degraded/partial health |
| **local-db-corrupt** | DB file replaced with invalid content | Tests migration error handling |
| **local-partial-migrations** | DB with only `001_` applied, `002_` pending | Tests incremental migration |
| **docker-compose** | Full stack via `docker compose up` | Tests M0-07 |

### Prerequisites

```bash
# Start MinIO only (for local-configured)
docker compose up minio -d

# Verify MinIO is up
curl http://localhost:9000/minio/health/live

# Start backend
cd backend && go run ./cmd/server

# Start frontend
cd frontend && npm run dev
```

### Environment variables under test (M0-02)

```
PORT=8080
DB_PATH=./moneyapp.db
JWT_SECRET=change-me
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=moneyapp
MINIO_USE_SSL=false
BCRYPT_COST=12
TOKEN_EXPIRY_HOURS=24
CURRENCY=VND
```

---

## Traceability Matrix

| Ticket | Title | Test Suite | Case Count | AC Bullets Covered |
|---|---|---|---|---|
| M0-01 | Migration Runner | TS-01 | 6 | AC1, AC2, AC3 |
| M0-02 | Env Config | TS-02 | 5 | AC1, AC2, AC3 |
| M0-03 | Error Handling & Middleware | TS-03 | 8 | AC1, AC2, AC3 |
| M0-04 | Frontend Routing & Shell | TS-04 | 8 | AC1, AC2, AC3, AC4 |
| M0-05 | Frontend API Client | TS-05 | 6 | AC1, AC2, AC3 |
| M0-06 | Health Endpoint | TS-06 | 6 | AC1, AC2, AC3 |
| M0-07 | Docker Compose | TS-07 | 4 | AC1, AC2 |
| M0-08 | CI Pipeline | TS-08 | 5 | AC1, AC2, AC3 |

---

## M0-01 — Database Migration Runner

**Ticket**: M0-01 | **NF**: NF-19 | **Priority**: Must  
**File under test**: `internal/database/migrate.go`, `backend/migrations/001_create_users.up.sql`

### Preconditions

- `DB_PATH` points to a writable directory.
- `backend/migrations/` directory exists and contains at least `001_create_users.up.sql`.
- No database file exists at `DB_PATH` (for fresh-DB cases) OR a pre-seeded DB (for incremental cases).

### Test Suite TS-01

| ID | Scenario | Environment | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-01-01 | Fresh DB — all migrations applied | local-fresh | 1. Delete DB file if present. 2. Start server (`go run ./cmd/server`). 3. Query `SELECT * FROM migrations ORDER BY id` via sqlite3 CLI. | All migration filenames present in `migrations` table; `users` table exists with columns `id, username, password_hash, created_at`. | AC1 |
| TS-01-02 | Idempotency — no re-apply on restart | local-configured | 1. Start server (migrations already applied). 2. Stop and restart server. 3. Query `migrations` table. | Row count unchanged; no duplicate entries; no errors in server logs. | AC2 |
| TS-01-03 | Incremental — only new migrations applied | local-partial-migrations | 1. Manually seed DB with `migrations` table containing only `001_create_users.up.sql`. 2. Add a second file `002_add_column.up.sql` (test fixture). 3. Start server. 4. Query `migrations` table. | Only `002_add_column.up.sql` is applied and recorded; row for `001_` is unchanged. | AC2 |
| TS-01-04 | **Negative**: Migration SQL syntax error → rollback | local-fresh | 1. Replace `001_create_users.up.sql` with invalid SQL (`CREATE TABLE INVALID (`). 2. Start server. 3. Observe exit and logs. | Server exits with non-zero code; log contains descriptive error; DB does not contain a partial `users` table (transaction rolled back). | AC3 |
| TS-01-05 | **Negative**: DB file not writable | local-fresh | 1. Set `DB_PATH` to a read-only path. 2. Start server. | Server exits immediately with a descriptive error message referencing the DB path. | AC3 |
| TS-01-06 | Ordering — migrations applied in numeric prefix order | local-fresh | 1. Place three migration files: `003_`, `001_`, `002_` (out of order on disk). 2. Start server. 3. Check `migrations` table ordering and absence of errors. | Migrations applied in `001 → 002 → 003` order regardless of filesystem order. | AC1 |

---

## M0-02 — Environment Configuration

**Ticket**: M0-02 | **NF**: NF-21 | **Priority**: Must  
**Files under test**: `.env.example` (root), `frontend/.env.example`, `internal/config/config.go`

### Preconditions

- Project is a fresh checkout.
- No `.env` file exists at project root.

### Test Suite TS-02

| ID | Scenario | Environment | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-02-01 | Default values — server starts from .env.example | local-fresh | 1. `cp .env.example .env`. 2. Start server. 3. Confirm listening on port 8080. | Server starts with no errors; logs show port 8080 and SQLite path matching `DB_PATH` default. | AC1 |
| TS-02-02 | Custom `DB_PATH` — DB created at custom location | local-configured | 1. Set `DB_PATH=/tmp/custom_moneyapp.db` in `.env`. 2. Start server. 3. Check for file at `/tmp/custom_moneyapp.db`. | SQLite file is created at the custom path; default path file not created. | AC3 |
| TS-02-03 | Custom `PORT` — server listens on custom port | local-configured | 1. Set `PORT=9090` in `.env`. 2. Start server. 3. `curl http://localhost:9090/api/health`. | 200 response from port 9090; port 8080 not listening. | AC1 |
| TS-02-04 | **Negative**: `JWT_SECRET` unset — warning or error | local-fresh | 1. Start server without setting `JWT_SECRET` (remove from `.env`). 2. Observe startup logs. | Server either exits with an error referencing `JWT_SECRET` OR logs a visible warning about using an insecure default; does not silently accept an empty secret. | AC2 |
| TS-02-05 | `.env.example` completeness | local-fresh | 1. Open `.env.example`. 2. Verify each variable from the spec is present: `PORT`, `DB_PATH`, `JWT_SECRET`, `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET`, `MINIO_USE_SSL`, `BCRYPT_COST`, `TOKEN_EXPIRY_HOURS`, `CURRENCY`. | All 11 variables present with their example values; file has no extra undocumented variables. | AC1 |

---

## M0-03 — API Error Handling & Middleware

**Ticket**: M0-03 | **NF**: NF-09 | **Priority**: Must  
**Files under test**: `internal/handlers/response.go`, `internal/handlers/middleware.go`

### Preconditions

- Server running with `local-configured` environment.
- A browser or curl is available to send requests.

### Test Suite TS-03

| ID | Scenario | Environment | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-03-01 | Error response shape — 404 | local-configured | `curl -s http://localhost:8080/api/nonexistent` | HTTP 404; `Content-Type: application/json`; body is `{"error":"..."}` (no raw Go error text, no stack trace). | AC1 |
| TS-03-02 | Error response — no internal detail leakage | local-configured | Trigger any expected error condition (e.g., invalid route). Inspect response body. | Body contains no file paths, Go package names, SQL statements, or stack traces. | AC1, NF-09 |
| TS-03-03 | Panic recovery — 500 returned, server stays alive | local-configured | 1. Temporarily inject a `panic("test")` in a handler or use a test endpoint if one exists. 2. Send request. 3. Send a second normal request immediately after. | First request returns HTTP 500 with generic message (e.g., `{"error":"internal server error"}`); second request succeeds; server did not crash. | AC2 |
| TS-03-04 | CORS — preflight from frontend origin | local-configured | `curl -s -X OPTIONS http://localhost:8080/api/health -H "Origin: http://localhost:5173" -H "Access-Control-Request-Method: GET" -v` | Response includes `Access-Control-Allow-Origin: http://localhost:5173` (or `*` per config); HTTP 204 or 200. | AC3 |
| TS-03-05 | CORS — request from non-frontend origin blocked | local-configured | `curl -s http://localhost:8080/api/health -H "Origin: http://evil.example.com" -v` | Response does NOT include `Access-Control-Allow-Origin: http://evil.example.com` (header absent or set to a different value). | AC3 |
| TS-03-06 | Logging middleware — request logged | local-configured | Send `GET /api/health`; check server stdout/log output. | Log line contains method (`GET`), path (`/api/health`), HTTP status (`200`), and duration (e.g., `3ms`). | AC2 |
| TS-03-07 | `respondJSON` — correct Content-Type header | local-configured | `curl -si http://localhost:8080/api/health` | Response headers include `Content-Type: application/json`. | AC1 |
| TS-03-08 | Request body size limit — oversized payload rejected | local-configured | `curl -s -X POST http://localhost:8080/api/health -d "$(python3 -c "print('x'*2000000)")"` | HTTP 413 or 400; does not hang; server not affected. | AC1 |

---

## M0-04 — Frontend Routing & App Shell

**Ticket**: M0-04 | **NF**: NF-16 | **Priority**: Must  
**Files under test**: `src/App.tsx`, `src/layouts/AppLayout.tsx`, `src/layouts/AuthLayout.tsx`, `src/pages/*`, `src/components/Sidebar.tsx`, `src/components/Toast.tsx`

### Preconditions

- `npm run dev` running at `http://localhost:5173`.
- Modern browser (Chrome/Firefox latest).

### Test Suite TS-04

| ID | Scenario | Viewport | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-04-01 | Route `/` → Dashboard in AppLayout | Desktop (1280×800) | Navigate to `http://localhost:5173/`. | DashboardPage placeholder renders inside AppLayout; sidebar visible with all 6 nav links (Dashboard, Expenses, Income, Invoices, Categories, Settings). | AC1 |
| TS-04-02 | Route `/expenses` → ExpensesPage in AppLayout | Desktop | Navigate to `/expenses`. | ExpensesPage placeholder renders inside AppLayout; URL stays `/expenses`. | AC1 |
| TS-04-03 | Route `/login` → LoginPage in AuthLayout | Desktop | Navigate to `/login`. | LoginPage renders; NO sidebar present; centered card layout visible. | AC2 |
| TS-04-04 | Unknown route → NotFoundPage | Desktop | Navigate to `/foo/bar`. | NotFoundPage renders with 404 message; no blank white screen. | AC3 |
| TS-04-05 | All nav links navigate correctly | Desktop | Click each sidebar link in turn. | URL changes to `/`, `/expenses`, `/income`, `/invoices`, `/categories`, `/settings`; page component updates accordingly; no full page reload. | AC1 |
| TS-04-06 | Responsive — sidebar collapses at < 768px | Mobile (375×812) | Load app at narrow viewport; inspect navigation. | Sidebar hidden or replaced by a top/hamburger nav; all page routes still accessible via mobile nav. | AC4, NF-16 |
| TS-04-07 | Responsive — 768px boundary | Tablet (768×1024) | Load app exactly at 768px wide. | Sidebar (or equivalent) visible; layout not broken; no horizontal scrollbar. | AC4, NF-16 |
| TS-04-08 | Toast component renders and auto-dismisses | Desktop | If a toast trigger is available in the shell, fire a test toast (error or info). | Toast visible; auto-dismisses after ~3 seconds; can stack multiple toasts without overlap issues. | AC1 (component readiness) |

#### Layout visual checklist (M0-04)

- [ ] Sidebar nav links have visible active state on current route
- [ ] Main content area has appropriate padding/margin (no content flush to edge)
- [ ] Font rendering: consistent typeface across all placeholder pages
- [ ] No unstyled HTML or missing CSS classes visible on any route
- [ ] AppLayout `<Outlet />` fills remaining space after sidebar without overflow

---

## M0-05 — Frontend API Client

**Ticket**: M0-05 | **Priority**: Must  
**Files under test**: `src/api/client.ts`, `src/types/api.ts`, `src/api/auth.ts`, `src/hooks/useAuth.ts`

### Preconditions

- Frontend dev server running.
- Backend running with `local-configured` environment.
- Browser DevTools available (Network tab + Application/localStorage).

### Test Suite TS-05

| ID | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|
| TS-05-01 | Valid token injected into request header | 1. Set `localStorage.setItem('token', 'test-jwt-value')`. 2. Trigger any API call (e.g., navigate to a page that calls an endpoint). 3. Inspect Network tab. | Request headers include `Authorization: Bearer test-jwt-value`. | AC1 |
| TS-05-02 | No token — no Authorization header sent | 1. `localStorage.removeItem('token')`. 2. Trigger API call. 3. Inspect Network tab. | No `Authorization` header in the request. | AC1 |
| TS-05-03 | 401 response — token cleared, redirect to `/login` | 1. Mock or force a 401 response from backend. 2. Trigger API call from any non-login page. | `localStorage` token removed; browser redirected to `/login`. | AC2 |
| TS-05-04 | `VITE_API_BASE_URL` used as base | 1. Set `VITE_API_BASE_URL=http://localhost:8080/api` in `frontend/.env`. 2. Restart dev server. 3. Trigger any API call. 4. Inspect Network tab. | All requests go to `http://localhost:8080/api/...`; no hardcoded URLs. | AC3 |
| TS-05-05 | Non-2xx response — typed error thrown | 1. Trigger a request to an endpoint that returns 404 or 500. 2. Observe error handling in the UI or console. | `apiClient` throws or rejects; calling code can distinguish error; no unhandled promise rejection crashes the app. | AC2 |
| TS-05-06 | TypeScript types — `ApiResponse`, `ApiListResponse`, `ApiError` exported | Code inspection of `src/types/api.ts`. | All three types present and match documented shapes: `ApiResponse<T>`, `ApiListResponse<T>`, `ApiError`. | AC3 |

---

## M0-06 — Health Check Endpoint

**Ticket**: M0-06 | **NF**: NF-18 | **Priority**: Must  
**File under test**: `internal/handlers/health.go` | **Endpoint**: `GET /api/health`

### Preconditions

- Backend running.
- MinIO may be up or down depending on scenario.
- `curl` available.

### Test Suite TS-06

| ID | Scenario | Environment | Command | Expected HTTP Status | Expected Body |
|---|---|---|---|---|---|
| TS-06-01 | All healthy | local-configured (MinIO up) | `curl -s http://localhost:8080/api/health` | `200` | `{"status":"ok","database":"ok","storage":"ok"}` |
| TS-06-02 | MinIO down — degraded | local-minio-down | `curl -s http://localhost:8080/api/health` | `200` | `{"status":"degraded","database":"ok","storage":"error"}` |
| TS-06-03 | **Negative**: DB unreachable — 503 | local-db-corrupt | Replace DB file with empty text file, restart server with a DB pinger that fails, then `curl /api/health`. | `503` | Body contains `"database":"error"` |
| TS-06-04 | Response shape — all fields present | local-configured | `curl -s http://localhost:8080/api/health \| jq 'keys'` | `200` | JSON keys: `["database","status","storage"]` exactly |
| TS-06-05 | Content-Type header correct | local-configured | `curl -si http://localhost:8080/api/health` | `200` | Header `Content-Type: application/json` present |
| TS-06-06 | Endpoint accessible without auth | local-configured | `curl -s http://localhost:8080/api/health` (no Authorization header) | `200` | Health data returned; no 401 |

---

## M0-07 — Docker Compose Enhancement

**Ticket**: M0-07 | **Priority**: Should  
**Files under test**: `docker-compose.yml`, `backend/Dockerfile`, `frontend/Dockerfile`

### Preconditions

- Docker Desktop or Docker Engine running.
- Ports 8080, 5173, 9000, 9001 free.

### Test Suite TS-07

| ID | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|
| TS-07-01 | Full stack starts cleanly | 1. `docker compose up --build -d`. 2. Wait for all containers healthy. 3. `curl http://localhost:8080/api/health`. | All three services (`backend`, `frontend`, `minio`) in "running" state; health endpoint returns `{"status":"ok",...}`. | AC1 |
| TS-07-02 | Backend waits for MinIO health check | 1. Observe `docker compose up` logs. 2. Check that `backend` container does not start (or does not log ready) until MinIO health check passes. | `docker compose logs backend` shows MinIO dependency satisfied before backend first handles requests; no "connection refused to MinIO" errors at startup. | AC2 |
| TS-07-03 | MinIO health check definition present | Inspect `docker-compose.yml`. | MinIO service has a `healthcheck` block using `curl --fail http://localhost:9000/minio/health/live` (or equivalent). | AC2 |
| TS-07-04 | **Negative**: Backend Dockerfile produces runnable binary | 1. `docker build -t moneyapp-backend ./backend`. 2. `docker run --rm -e DB_PATH=:memory: moneyapp-backend`. | Container starts (or exits cleanly due to missing MinIO) without `exec format error` or missing binary errors. | AC1 |

---

## M0-08 — CI Pipeline

**Ticket**: M0-08 | **Priority**: Should  
**File under test**: `.github/workflows/ci.yml`

### Preconditions

- Repository pushed to GitHub (or equivalent CI host).
- Pipeline configured to trigger on push to any branch.

### Test Suite TS-08

| ID | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|
| TS-08-01 | Clean push — all jobs pass | Push a branch with no code errors. | All CI jobs green: `backend-lint`, `backend-build`, `backend-test`, `frontend-lint`, `frontend-build`. | AC1 |
| TS-08-02 | Go lint failure detected | 1. Push a branch with a `go vet` violation (e.g., `fmt.Sprintf` without args). 2. Observe `backend-lint` job. | `backend-lint` fails with a message identifying the offending file/line; other jobs may still run. | AC2 |
| TS-08-03 | TypeScript error detected | 1. Push a branch with a TS type mismatch. 2. Observe `frontend-build` job. | `frontend-build` fails referencing the type error; error message visible in CI logs. | AC3 |
| TS-08-04 | Go tests run | In CI logs for `backend-test`, confirm `go test ./...` was executed. | Test output visible in logs; no job skipped. | AC1 |
| TS-08-05 | CI triggers on PR | Open a draft PR from a feature branch. | CI pipeline starts automatically; all jobs listed in PR check status. | AC1 |

---

## Cross-Cutting Checks

These checks apply across multiple M0 tickets and should be verified once the full M0 stack is assembled.

### Security spot-checks

| Check | Steps | Expected |
|---|---|---|
| No secrets in responses | Call `/api/health` and any error endpoint. Inspect full response body. | No JWT secret, MinIO credentials, DB path with sensitive data, or internal stack traces visible. |
| `.env` not committed | `git ls-files .env` | Empty output (`.env` not tracked). |
| `.env.example` committed | `git ls-files .env.example` | File present in repo. |

### Startup smoke test (all-in-one)

Run after all M0 tickets are implemented:

```bash
# 1. Fresh clone
git clone <repo> /tmp/moneyapp-smoke && cd /tmp/moneyapp-smoke

# 2. Copy env
cp .env.example .env
cp frontend/.env.example frontend/.env

# 3. Start MinIO
docker compose up minio -d
sleep 5

# 4. Start backend (should migrate and listen)
cd backend && go run ./cmd/server &
sleep 3

# 5. Verify health
curl -s http://localhost:8080/api/health | jq .

# 6. Start frontend
cd ../frontend && npm run dev &
sleep 5

# 7. Verify frontend reachable
curl -s http://localhost:5173 | grep -c "<div"

# 8. Cleanup
kill %1 %2
```

Expected: Step 5 returns `{"status":"ok",...}`; step 7 returns `> 0`.

### Panic/recovery smoke

```bash
# If a /debug/panic test endpoint is added during dev, use it:
curl -s http://localhost:8080/debug/panic
# Immediately follow with:
curl -s http://localhost:8080/api/health
```

Expected: First request returns HTTP 500 with `{"error":"internal server error"}`; second returns 200; server did not exit.

### CORS smoke

```bash
curl -sv -X OPTIONS http://localhost:8080/api/health \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: GET" \
  2>&1 | grep -i "access-control"
```

Expected: `Access-Control-Allow-Origin` header present matching the frontend origin.

---

## Automation Backlog

The following test cases are candidates for Playwright or Go test automation in a future pass. They are **not required** for M0 sign-off but are recorded here to reduce future discovery cost.

### Playwright — frontend routes (TS-04)

```
Route: describe('App Shell Routing')
  - hooks: page.goto('/'), page.goto('/expenses'), page.goto('/login'), page.goto('/foo')
  - selectors to target:
      sidebar nav: [data-testid="sidebar"] or nav[aria-label="Main navigation"]
      page content: [data-testid="page-dashboard"], [data-testid="page-expenses"], etc.
      auth layout: [data-testid="auth-layout"]
      not-found: [data-testid="page-not-found"] or role=heading with "404"
  - viewport tests: use page.setViewportSize({ width: 375, height: 812 }) for mobile
```

### Playwright — API client 401 redirect (TS-05-03)

```
Route: describe('API Client 401 Handling')
  - Use page.route('**/api/**', route => route.fulfill({ status: 401, body: '{"error":"unauthorized"}' }))
  - Assert page.url() ends with '/login' after intercepted 401
  - Assert localStorage.getItem('token') === null
```

### Go integration tests — migration runner (TS-01)

```
Package: internal/database
  - TestRunMigrations_FreshDB: use t.TempDir() for DB path; assert tables exist
  - TestRunMigrations_Idempotent: run twice; assert migration count unchanged
  - TestRunMigrations_PartialApply: pre-seed migrations table; assert only new ones applied
  - TestRunMigrations_RollbackOnError: inject bad SQL; assert no partial schema
```

### Go integration tests — health handler (TS-06)

```
Package: internal/handlers
  - TestHealthHandler_AllOk: mock DB ping OK + MinIO OK → 200 {"status":"ok",...}
  - TestHealthHandler_MinioDown: mock MinIO fail → 200 {"status":"degraded",...}
  - TestHealthHandler_DBDown: mock DB ping fail → 503
```

### Go unit tests — config loader (TS-02)

```
Package: internal/config
  - TestLoad_Defaults: unset all env vars; assert defaults match .env.example values
  - TestLoad_CustomDBPath: set DB_PATH; assert Config.DBPath matches
  - TestLoad_MissingJWTSecret: unset JWT_SECRET; assert warning log or error
```

---

## Coverage Gaps & Notes

The following gaps were noticed while authoring this plan against tickets M0-01 through M0-08:

| Gap | Ticket | Detail |
|---|---|---|
| **JWT_SECRET policy unspecified** | M0-02 | AC2 says "logs a warning (or errors out, depending on security policy)" — the policy is not decided. Test TS-02-04 covers both outcomes; ticket should be clarified before M0 sign-off. |
| **No test endpoint for panic recovery** | M0-03 | AC2 (panic → 500) is hard to test in black-box without a dedicated `/debug/panic` endpoint or an injectable handler. Recommend adding a dev-only panic endpoint or a Go httptest unit test. |
| **Frontend `Toast` testability** | M0-04 | The `Toast` component is specified but no trigger mechanism is defined in M0. TS-04-08 is conditional ("if trigger is available"). A `data-testid` attribute and a dev-only trigger prop should be added to enable reliable testing. |
| **`useAuth` hook state persistence** | M0-05 | The hook stores the token in `localStorage` but the spec does not address page refresh behavior (does the hook re-hydrate from localStorage on mount?). This should be tested: TS-05-01 assumes a pre-set localStorage token but the hook loading it on mount is implied, not explicitly stated. |
| **M0-07 frontend Dockerfile optional** | M0-07 | The ticket marks the `frontend` Compose service as "optional". TS-07 only tests the backend + MinIO path. If the frontend service is added, extend TS-07 with a browser reachability check. |
| **M0-08 CI host not specified** | M0-08 | Ticket says "GitHub Actions (or equivalent)". TS-08 assumes GitHub Actions. If a different CI host (GitLab CI, etc.) is used, selector/log-reading steps differ. |
| **DB-down simulation difficulty** | M0-06 | Simulating a fully unreachable SQLite DB (TS-06-03) is non-trivial — SQLite is embedded. The most realistic approach is a unit/httptest test with a mocked `db.Ping()` that returns an error. Black-box approach: replace DB file with a directory of same name (causes Open to fail). |
