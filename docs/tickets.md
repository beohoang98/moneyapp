# MoneyApp -- Story Tickets

**Generated from**: `docs/requirements.md` v1.2
**Date**: 2026-04-26
**Status**: Draft

---

## Table of Contents

- [Milestone 0 -- Project Foundation](#milestone-0----project-foundation)
- [Milestone 1 -- MVP Core](#milestone-1----mvp-core)
- [Milestone 2 -- File Attachments & Enhanced Lists](#milestone-2----file-attachments--enhanced-lists)
- [Milestone 3 -- Reports & Export](#milestone-3----reports--export)
- [Milestone 4 -- Polish & Nice-to-Haves](#milestone-4----polish--nice-to-haves)

---

## Milestone 0 -- Project Foundation

**Goal**: Runnable skeleton with infrastructure in place. No user-facing features yet.

**Exit criteria**: `go run ./cmd/server` and `npm run dev` start without errors; storage backend is reachable (MinIO bucket when `STORAGE_TYPE=s3`, writable path when `STORAGE_TYPE=local`); health endpoint returns 200; migrations run on startup; frontend renders a shell with routing.

---

### M0-01: Database Migration Runner

| Field | Value |
|---|---|
| **Ticket ID** | M0-01 |
| **Title** | Implement versioned database migration runner |
| **Epic** | Infrastructure |
| **Milestone** | M0 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | None |

**Description**: Replace the current stub `migrate()` function in `internal/database/db.go` with a proper versioned migration runner. Each migration should be a numbered `.sql` file in a `migrations/` directory. The runner must track which migrations have been applied via the existing `migrations` table and apply only unapplied ones in order. This satisfies NF-19.

**Backend tasks**:
- Create `backend/migrations/` directory for `.sql` files.
- Implement `RunMigrations(db *sql.DB, migrationsDir string) error` in `internal/database/migrate.go`. Read all `*.up.sql` files from the embedded `migrations/` directory (use `embed.FS`), sort by filename prefix (e.g., `001_`, `002_`), skip already-applied ones, execute new ones in a transaction, and record each in the `migrations` table.
- Add the initial migration file `001_create_users.up.sql` creating the `users` table: `id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP`.
- Update `database.Open()` to call `RunMigrations`.
- Enable WAL mode and foreign keys (already done in `Open`).

**Frontend tasks**: None.

**Acceptance criteria**:
- Given a fresh database, when the server starts, then all migration files are applied in order and recorded in the `migrations` table.
- Given an existing database with some migrations applied, when the server starts, then only new migrations are applied.
- Given a migration that fails, when the runner encounters the error, then no partial changes are committed (transaction rollback) and the server exits with a descriptive error log.

---

### M0-02: Environment Configuration and .env.example

| Field | Value |
|---|---|
| **Ticket ID** | M0-02 |
| **Title** | Create .env.example and config loader |
| **Epic** | Infrastructure |
| **Milestone** | M0 |
| **Priority** | Must |
| **Size** | XS |
| **Dependencies** | None |

**Description**: Create a `.env.example` file documenting all required environment variables (NF-21). Implement a config struct in Go that loads values from environment variables with sensible defaults.

**Backend tasks**:
- Create `.env.example` at project root with:
  - `PORT=8080`
  - `DB_PATH=./moneyapp.db`
  - `JWT_SECRET=change-me`
  - `BCRYPT_COST=12`
  - `TOKEN_EXPIRY_HOURS=24`
  - `CURRENCY=VND`
  - `STORAGE_TYPE=local` *(storage backend selector: `local` or `s3`)*
  - `LOCAL_STORAGE_PATH=./data/storage` *(used when `STORAGE_TYPE=local`; created on startup if missing)*
  - `MINIO_ENDPOINT=localhost:9000` *(used when `STORAGE_TYPE=s3`)*
  - `MINIO_ACCESS_KEY=minioadmin` *(used when `STORAGE_TYPE=s3`)*
  - `MINIO_SECRET_KEY=minioadmin` *(used when `STORAGE_TYPE=s3`)*
  - `MINIO_BUCKET=moneyapp` *(used when `STORAGE_TYPE=s3`)*
  - `MINIO_USE_SSL=false` *(used when `STORAGE_TYPE=s3`)*
- Create `internal/config/config.go` with a `Config` struct (fields: `Port`, `DBPath`, `JWTSecret`, `BcryptCost`, `TokenExpiryHours`, `Currency`, `StorageType`, `LocalStoragePath`, `MinioEndpoint`, `MinioAccessKey`, `MinioSecretKey`, `MinioBucket`, `MinioUseSSL`) and `Load() (*Config, error)` function that reads from `os.Getenv` with defaults.
- Update `cmd/server/main.go` to use `config.Load()` instead of inline `os.Getenv`.

**Frontend tasks**:
- Create `frontend/.env.example` with `VITE_API_BASE_URL=http://localhost:8080/api`.

**Acceptance criteria**:
- Given a clean checkout, when a developer copies `.env.example` to `.env`, then the server starts successfully with default values (`STORAGE_TYPE=local`).
- Given `STORAGE_TYPE=local` and `LOCAL_STORAGE_PATH=./data/storage`, when the server starts, then the `./data/storage` directory is created if it does not already exist.
- Given `STORAGE_TYPE=s3`, when the server starts, then it reads all `MINIO_*` variables and connects to the configured endpoint; if the endpoint or bucket is unreachable, then the server exits at startup with a non-zero status and a descriptive log (fail-fast — no degraded `/api/health` while running).
- Given `JWT_SECRET` is not set, when the server starts, then it logs a warning (or errors out, depending on security policy).
- Given `DB_PATH` is set to a custom path, when the server starts, then the SQLite file is created at that path.

---

### M0-03: API Error Handling and Response Patterns

| Field | Value |
|---|---|
| **Ticket ID** | M0-03 |
| **Title** | Standardized API error responses and JSON helpers |
| **Epic** | Infrastructure |
| **Milestone** | M0 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | None |

**Description**: Establish consistent JSON response patterns across the API. All endpoints must return structured JSON for both success and error cases. Error responses must not leak internal details (NF-09). This includes JSON encoding/decoding helpers and a standard error response format.

**Backend tasks**:
- Create `internal/handlers/response.go` with helpers:
  - `respondJSON(w http.ResponseWriter, status int, data any)` -- writes JSON with correct Content-Type and status.
  - `respondError(w http.ResponseWriter, status int, message string)` -- writes `{"error": "message"}` JSON.
  - `decodeJSON(r *http.Request, dst any) error` -- decodes request body with size limit (1MB default).
- Define a standard list response wrapper: `{"data": [...], "total": N, "page": N, "per_page": N}`.
- Define a standard single-item response: `{"data": {...}}`.
- Create `internal/handlers/middleware.go` with:
  - `LoggingMiddleware` -- logs method, path, status, and duration.
  - `RecoveryMiddleware` -- catches panics and returns 500 with a generic message.
  - `CORSMiddleware` -- allows the frontend origin (configurable).
- Update `cmd/server/main.go` to wrap the mux with middleware.
- Update the `/health` endpoint to use `respondJSON`.

**Frontend tasks**: None (frontend API client is M0-05).

**Acceptance criteria**:
- Given any API error, when the response is returned, then it has Content-Type `application/json` and body `{"error": "..."}` with an appropriate HTTP status code.
- Given a panic in a handler, when the request is processed, then the client receives a 500 response and the server logs the panic details but does not crash.
- Given a request from the frontend origin, when CORS headers are checked, then the response includes `Access-Control-Allow-Origin` matching the configured frontend URL.

---

### M0-04: Frontend Routing and App Shell

| Field | Value |
|---|---|
| **Ticket ID** | M0-04 |
| **Title** | Set up React Router and application shell layout |
| **Epic** | Infrastructure |
| **Milestone** | M0 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | None |

**Description**: Install React Router and set up the application shell with a sidebar/navbar layout, route definitions, and placeholder pages. This establishes the navigation structure all features will build on.

**Frontend tasks**:
- Install `react-router-dom` (`npm install react-router-dom`).
- Create `src/layouts/AppLayout.tsx` -- sidebar with navigation links (Dashboard, Expenses, Income, Invoices, Categories, Settings) and a main content area using `<Outlet />`.
- Create `src/layouts/AuthLayout.tsx` -- centered single-card layout for login page.
- Create placeholder pages in `src/pages/`:
  - `DashboardPage.tsx`
  - `ExpensesPage.tsx`
  - `IncomePage.tsx`
  - `InvoicesPage.tsx`
  - `CategoriesPage.tsx`
  - `SettingsPage.tsx`
  - `LoginPage.tsx`
  - `NotFoundPage.tsx`
- Update `src/App.tsx` to configure routes:
  - `/login` -- AuthLayout > LoginPage
  - `/` -- AppLayout > DashboardPage (index redirect)
  - `/expenses` -- AppLayout > ExpensesPage
  - `/income` -- AppLayout > IncomePage
  - `/invoices` -- AppLayout > InvoicesPage
  - `/categories` -- AppLayout > CategoriesPage
  - `/settings` -- AppLayout > SettingsPage
  - `*` -- NotFoundPage
- Create `src/components/Sidebar.tsx` with nav links.
- Create `src/components/Toast.tsx` -- a reusable toast/notification component for success, error, and info feedback messages. Auto-dismiss after a configurable duration (default 3 seconds). Support stacking multiple toasts. This component is referenced by E2-S6 (delete success) and other CRUD operations throughout the app.
- Ensure the layout is responsive for screens >= 768px (NF-16): sidebar collapses to top nav on smaller screens.

**Acceptance criteria**:
- Given the frontend is running, when the user navigates to `/expenses`, then the ExpensesPage placeholder renders inside the AppLayout.
- Given the user navigates to `/login`, then the LoginPage renders inside the AuthLayout (no sidebar).
- Given a non-existent route like `/foo`, then the NotFoundPage renders.
- Given a screen width < 768px, when the layout renders, then the sidebar collapses to a mobile-friendly navigation.

---

### M0-05: Frontend API Client

| Field | Value |
|---|---|
| **Ticket ID** | M0-05 |
| **Title** | Set up typed API client with interceptors |
| **Epic** | Infrastructure |
| **Milestone** | M0 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | M0-03 |

**Description**: Create a centralized API client using `fetch` (no extra library needed for this project size) with JWT token injection, error handling, and TypeScript types matching the backend response format.

**Frontend tasks**:
- Create `src/api/client.ts`:
  - Base URL from `import.meta.env.VITE_API_BASE_URL`.
  - `apiClient` object with methods: `get<T>(path)`, `post<T>(path, body)`, `put<T>(path, body)`, `delete(path)`, `upload<T>(path, formData)`.
  - Automatically attach `Authorization: Bearer <token>` header from localStorage.
  - On 401 response, clear token and redirect to `/login`.
  - Parse JSON responses and throw typed errors on non-2xx.
- Create `src/types/api.ts`:
  - `ApiResponse<T> = { data: T }`.
  - `ApiListResponse<T> = { data: T[]; total: number; page: number; per_page: number }`.
  - `ApiError = { error: string }`.
- Create `src/api/auth.ts` with `login(username, password)` and `logout()` functions (stubs for now, wired in M1).
- Create `src/hooks/useAuth.ts` with auth state management: `isAuthenticated`, `token`, `login()`, `logout()`. Store token in localStorage.

**Acceptance criteria**:
- Given a valid token in localStorage, when any API call is made, then the `Authorization` header is included.
- Given a 401 response from any API call, when the error is handled, then the user is redirected to `/login` and the stored token is cleared.
- Given the API base URL is configured via `VITE_API_BASE_URL`, when the client makes a request, then it uses that base URL.

---

### M0-06: Health Check Endpoint Enhancement

| Field | Value |
|---|---|
| **Ticket ID** | M0-06 |
| **Title** | Enhance health endpoint with DB and storage backend status |
| **Epic** | Infrastructure |
| **Milestone** | M0 |
| **Priority** | Must |
| **Size** | XS |
| **Dependencies** | M0-02, M0-03 |

**Description**: Upgrade the existing `GET /api/health` endpoint to report the status of both the database and the configured storage backend (NF-18). Return structured JSON with individual component statuses. The storage check must be appropriate to the active `STORAGE_TYPE` — it must not report an error for a MinIO endpoint that is intentionally absent when running in local mode.

**Backend tasks**:
- Move the health handler to `internal/handlers/health.go`.
- Response format: `{"status": "ok"|"degraded", "database": "ok"|"error", "storage": "ok"|"error", "storage_type": "local"|"s3"}`.
- Ping the database (`db.Ping()`).
- Check storage health based on `STORAGE_TYPE`:
  - `local` — verify `LOCAL_STORAGE_PATH` exists and is writable (e.g., create and delete a probe file).
  - `s3` — attempt a MinIO `BucketExists` check against the configured endpoint.
- If database is down, return HTTP 503. If only storage is unhealthy, return 200 with `"storage": "error"` and `"status": "degraded"`.

**Frontend tasks**: None.

**Acceptance criteria**:
- Given `STORAGE_TYPE=s3` and both DB and MinIO are healthy, when `GET /api/health` is called, then the response is `200 {"status":"ok","database":"ok","storage":"ok","storage_type":"s3"}`.
- Given `STORAGE_TYPE=s3` and the server is already running, when MinIO becomes unreachable and `GET /api/health` is called, then the response is `200` with `"status":"degraded"`, `"database":"ok"`, `"storage":"error"`, and `"storage_type":"s3"`. (If MinIO is unreachable at **process startup**, the server exits with fail-fast instead of serving this response.)
- Given `STORAGE_TYPE=local` and the local path is writable, when `GET /api/health` is called, then the response is `200 {"status":"ok","database":"ok","storage":"ok","storage_type":"local"}` — even when no MinIO service is running.
- Given `STORAGE_TYPE=local` and the local path is not writable, when `GET /api/health` is called, then `"storage": "error"` is returned.
- Given the database is unreachable, when `GET /api/health` is called, then the response is `503`.

---

### M0-07: Docker Compose Enhancement

| Field | Value |
|---|---|
| **Ticket ID** | M0-07 |
| **Title** | Add backend service and health checks to Docker Compose |
| **Epic** | Infrastructure |
| **Milestone** | M0 |
| **Priority** | Should |
| **Size** | XS |
| **Dependencies** | M0-02 |

**Description**: Enhance `docker-compose.yml` with a backend service (using a multi-stage Dockerfile), frontend dev service, and health checks for MinIO. This enables `docker compose up` for a full local development stack using MinIO as the S3-compatible storage backend.

**Backend tasks**:
- Create `backend/Dockerfile` -- multi-stage build: build stage using `golang:1.26` image, runtime stage using `gcr.io/distroless/base` or `alpine:latest`. Copy binary, expose port 8080.
- Add `backend` service to `docker-compose.yml` depending on `minio`, with env vars from `.env`. Explicitly set `STORAGE_TYPE=s3` in the backend service environment (or document that the `.env` file used by Compose must set it) so the stack uses MinIO rather than local filesystem storage.
- Add `healthcheck` to the MinIO service (`curl --fail http://localhost:9000/minio/health/live`).

**Frontend tasks**:
- Create `frontend/Dockerfile` for production build (multi-stage: node build, nginx serve).
- Optionally add a `frontend` service to `docker-compose.yml` for dev mode.

**Acceptance criteria**:
- Given the Docker Compose file, when `docker compose up` is run, then the backend, frontend, and MinIO all start and the backend can reach MinIO.
- Given MinIO is starting, when the backend depends on it, then the backend waits for MinIO's health check to pass before starting.
- Given `STORAGE_TYPE=s3` is set in the Compose stack, when a file is uploaded, then it is stored in the MinIO bucket (not the local filesystem).
- **Note**: For bare-metal `go run` without Docker, set `STORAGE_TYPE=local` in `.env` to use local filesystem storage without requiring MinIO.

---

### M0-08: CI Pipeline Setup

| Field | Value |
|---|---|
| **Ticket ID** | M0-08 |
| **Title** | Set up CI pipeline with lint and build checks |
| **Epic** | Infrastructure |
| **Milestone** | M0 |
| **Priority** | Should |
| **Size** | S |
| **Dependencies** | M0-01, M0-04 |

**Description**: Create a GitHub Actions (or equivalent) CI pipeline that runs on every push and PR. It should lint and build both the backend and frontend to catch errors early.

**Backend tasks**:
- Create `.github/workflows/ci.yml`.
- Jobs: `backend-lint` (run `go vet ./...`), `backend-build` (run `go build ./cmd/server`), `backend-test` (run `go test ./...`).

**Frontend tasks**:
- CI jobs: `frontend-lint` (run `npm run lint`), `frontend-build` (run `npm run build`).

**Acceptance criteria**:
- Given a push to any branch, when the CI pipeline runs, then it executes lint and build for both backend and frontend.
- Given a linting error in Go code, when the pipeline runs, then it fails with a clear error message.
- Given a TypeScript type error, when the pipeline runs, then it fails with a clear error message.

---

## Milestone 1 -- MVP Core

**Goal**: Fully functional core with auth, CRUD for all entities, basic filtering, and dashboard summary cards. No file attachments or charts yet.

**Exit criteria**: User can log in, record expenses/income/invoices, filter lists, and see a dashboard summary. All data persists across restarts.

---

### E1-S1: User Login

| Field | Value |
|---|---|
| **Ticket ID** | E1-S1 |
| **Title** | Implement user login with JWT authentication |
| **Epic** | Epic 1 -- Authentication |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | M0-01, M0-02, M0-03, M0-05 |

**Description**: Implement the login flow end-to-end. The backend issues a JWT on valid credentials, and the frontend provides a login form that stores the token and redirects to the dashboard. A seed user must be created on first run (since there is no registration -- see AU-08).

**Backend tasks**:
- Add `golang.org/x/crypto/bcrypt` and a JWT library (e.g., `github.com/golang-jwt/jwt/v5`) to `go.mod`.
- Create migration `002_seed_default_user.up.sql`: insert a default user `admin` with a bcrypt-hashed default password (e.g., `changeme`, cost 12). Use a raw bcrypt hash literal in the SQL.
- Create `internal/services/auth_service.go`:
  - `AuthService` struct with `db *sql.DB` and `jwtSecret []byte` and `tokenExpiry time.Duration`.
  - `Login(ctx, username, password string) (token string, err error)` -- query user by username, compare bcrypt hash, generate JWT with `sub: userID`, `exp`, `iat` claims.
  - `ValidateToken(tokenString string) (userID int64, err error)` -- parse and validate JWT.
- Create `internal/handlers/auth_handler.go`:
  - `POST /api/auth/login` -- accepts `{"username": "...", "password": "..."}`, returns `{"data": {"token": "...", "expires_at": "..."}}` or 401.
- Create `internal/handlers/middleware.go` (extend from M0-03):
  - `AuthMiddleware(authService) func(http.Handler) http.Handler` -- extracts `Authorization: Bearer <token>`, validates, injects user ID into request context. Returns 401 on invalid/missing token.
- Register auth routes and protect all `/api/*` routes except `/api/auth/login` and `/api/health`.

**Frontend tasks**:
- Implement `src/pages/LoginPage.tsx`:
  - Form with `username` and `password` fields, submit button.
  - Client-side validation: both fields required (NF-14).
  - On submit, call `POST /api/auth/login`.
  - On success, store token in localStorage via `useAuth` hook, redirect to `/`.
  - On 401, display "Invalid username or password" error.
- Update `src/hooks/useAuth.ts`:
  - `login(username, password)` -- calls API, stores token.
  - `isAuthenticated` -- checks if token exists and is not expired (decode JWT client-side for exp check).
- Create `src/components/ProtectedRoute.tsx` -- wraps routes that require auth; redirects to `/login` if not authenticated.
- Update `src/App.tsx` to wrap AppLayout routes with `ProtectedRoute`.

**Acceptance criteria**:
- Given valid credentials (`admin` / `changeme`), when the user submits the login form, then the system returns a JWT token and the frontend redirects to the dashboard.
- Given invalid credentials, when the user submits the login form, then the system returns HTTP 401 and the form displays "Invalid username or password".
- Given a blank username or password, when the user clicks submit, then client-side validation prevents the request and shows field-level errors.
- Given no token or an expired token, when the user tries to access any protected page, then they are redirected to `/login`.

---

### E1-S2: Token Expiry

| Field | Value |
|---|---|
| **Ticket ID** | E1-S2 |
| **Title** | JWT token expiry after configurable idle period |
| **Epic** | Epic 1 -- Authentication |
| **Milestone** | M1 |
| **Priority** | Should |
| **Size** | S |
| **Dependencies** | E1-S1 |

**Description**: Tokens must expire after a configurable period (default 24 hours, from `TOKEN_EXPIRY_HOURS` env var). The frontend must detect expired tokens and redirect to login gracefully.

**Backend tasks**:
- Ensure the JWT `exp` claim is set to `now + TOKEN_EXPIRY_HOURS` in `AuthService.Login()`.
- In `AuthMiddleware`, check `exp` claim during validation. Return 401 with `{"error": "token expired"}` if expired.

**Frontend tasks**:
- In `src/api/client.ts`, on 401 responses, distinguish between "token expired" and "invalid token" if needed, clear localStorage, and redirect to `/login`.
- Optionally decode the JWT on the frontend to proactively check expiry before making requests (avoids unnecessary 401 round trips).

**Acceptance criteria**:
- Given a token issued 25 hours ago (with default 24h expiry), when the user makes an API request, then the server returns 401 and the frontend redirects to `/login`.
- Given `TOKEN_EXPIRY_HOURS=1` in the environment, when a token is issued, then it expires after 1 hour.

---

### E1-S3: User Logout

| Field | Value |
|---|---|
| **Ticket ID** | E1-S3 |
| **Title** | Manual logout from the application |
| **Epic** | Epic 1 -- Authentication |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | XS |
| **Dependencies** | E1-S1 |

**Description**: Allow the user to manually log out, clearing their session. Since JWTs are stateless, logout is primarily a client-side operation (clear token). Optionally, the backend can accept a logout call for audit logging.

**Backend tasks**:
- Create `POST /api/auth/logout` -- returns 200. (Stateless JWT means no server-side invalidation needed for MVP. This endpoint exists for future token blacklisting if needed.)

**Frontend tasks**:
- Add a "Logout" button to the sidebar/navbar in `src/components/Sidebar.tsx`.
- On click, call `logout()` from `useAuth` hook: clear token from localStorage, redirect to `/login`.
- Show a brief confirmation or simply redirect.

**Acceptance criteria**:
- Given a logged-in user, when they click the Logout button, then the token is removed from localStorage and they are redirected to `/login`.
- Given a logged-out user, when they try to navigate to a protected route, then they are redirected to `/login`.

---

### E5-S1: Default Categories

| Field | Value |
|---|---|
| **Ticket ID** | E5-S1 |
| **Title** | Seed default expense and income categories |
| **Epic** | Epic 5 -- Categories |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | XS |
| **Dependencies** | M0-01 |

**Description**: Create the categories table and seed it with default categories for both expenses and income. Categories are required before expenses or income can be created (foreign key dependency).

**Backend tasks**:
- Create migration `003_create_categories.up.sql`:
  ```sql
  CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('expense', 'income')),
    is_default BOOLEAN NOT NULL DEFAULT 0,
    color TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  CREATE UNIQUE INDEX idx_categories_name_type ON categories(name, type);
  ```
- Create migration `004_seed_default_categories.up.sql`:
  - Expense categories (is_default=1): Food, Transport, Housing, Health, Entertainment, Shopping, Utilities, Other, Uncategorized.
  - Income categories (is_default=1): Salary, Freelance, Investment, Gift, Other, Uncategorized.
- Create `internal/models/category.go` with `Category` struct: `ID int64`, `Name string`, `Type string` (expense|income), `IsDefault bool`, `Color *string`, `CreatedAt time.Time`.
- Create `internal/services/category_service.go`:
  - `ListCategories(ctx, categoryType string) ([]Category, error)` -- query categories filtered by type.
- Create `internal/handlers/category_handler.go`:
  - `GET /api/categories?type=expense|income` -- returns list of categories. Protected by auth middleware.
- Register the route in `cmd/server/main.go`.

**Frontend tasks**:
- Create `src/types/category.ts`: `Category { id: number; name: string; type: 'expense' | 'income'; is_default: boolean; color?: string }`.
- Create `src/api/categories.ts`: `getCategories(type: 'expense' | 'income'): Promise<Category[]>`.

**Acceptance criteria**:
- Given a fresh database, when migrations run, then the `categories` table contains 9 expense categories and 6 income categories, all with `is_default = true`.
- Given an authenticated request to `GET /api/categories?type=expense`, then the response contains all default expense categories.
- Given an unauthenticated request, then the response is 401.

---

### E2-S1: Create Expense

| Field | Value |
|---|---|
| **Ticket ID** | E2-S1 |
| **Title** | Create a new expense record |
| **Epic** | Epic 2 -- Expense Tracking |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E1-S1, E5-S1 |

**Description**: Allow the user to create an expense with amount, date, category, and optional description. Amounts are stored as integers in minor currency units (NF-10). The frontend must convert display amounts to minor units before sending to the API.

**Backend tasks**:
- Create migration `005_create_expenses.up.sql`:
  ```sql
  CREATE TABLE expenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount INTEGER NOT NULL CHECK(amount > 0),
    date DATE NOT NULL,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    description TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  CREATE INDEX idx_expenses_date ON expenses(date);
  CREATE INDEX idx_expenses_category_id ON expenses(category_id);
  ```
- Update `internal/models/models.go`: update `Expense` struct to use `CategoryID int64` (FK) instead of `Category string`. Remove `Currency` field (single-currency app).
- Create `internal/services/expense_service.go`:
  - `ExpenseService` struct with `db *sql.DB`.
  - `Create(ctx, expense *Expense) error` -- validate amount > 0, validate category_id exists and is type "expense", default date to today if empty, insert row, return with generated ID.
- Create `internal/handlers/expense_handler.go`:
  - `POST /api/expenses` -- accepts `{"amount": 50000, "date": "2026-04-26", "category_id": 1, "description": "Lunch"}`. Returns 201 with created expense. Returns 400 on validation errors with field-level messages.
- Register the route.

**Frontend tasks**:
- Create `src/types/expense.ts`: `Expense { id: number; amount: number; date: string; category_id: number; category_name?: string; description: string; created_at: string; updated_at: string }`.
- Create `src/api/expenses.ts`: `createExpense(data: CreateExpensePayload): Promise<Expense>`.
- Create `src/components/expenses/ExpenseForm.tsx`:
  - Fields: amount (number input, displayed in major units with currency symbol), date (date picker, defaults to today), category (dropdown loaded from `GET /api/categories?type=expense`), description (textarea, optional).
  - Client-side validation (NF-14): amount required and > 0, date required, category required.
  - On submit, convert amount from major to minor units (multiply by 100 or appropriate factor), call `createExpense`.
  - Display field-level error messages.
  - On success, close form and refresh the list (or navigate to the list).
- Create `src/pages/ExpensesPage.tsx` with an "Add Expense" button that opens the form (modal or inline).

**Acceptance criteria**:
- Given a valid amount (50000), date (2026-04-26), and category (Food), when the user saves the expense, then it appears in the expense list and is persisted to the `expenses` table.
- Given a negative or zero amount, when the user tries to save, then validation rejects it with "Amount must be greater than zero".
- Given a missing date, when the user tries to save, then the date defaults to today.
- Given an invalid category_id, when the API receives the request, then it returns 400 with a descriptive error.

---

### E2-S2: List Expenses (Paginated)

| Field | Value |
|---|---|
| **Ticket ID** | E2-S2 |
| **Title** | View paginated list of all expenses |
| **Epic** | Epic 2 -- Expense Tracking |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E2-S1 |

**Description**: Display all expenses in a paginated table/list with columns for date, category, description, and amount. Supports server-side pagination.

**Backend tasks**:
- Add to `ExpenseService`:
  - `List(ctx, page int, perPage int) (expenses []Expense, total int64, err error)` -- paginated query with `LIMIT/OFFSET`. Join with `categories` table to include `category_name`.
  - Default `perPage = 20`, max `perPage = 100`.
- Add to expense handler:
  - `GET /api/expenses?page=1&per_page=20` -- returns `{"data": [...], "total": N, "page": 1, "per_page": 20}`. Each expense includes `category_name` from the join.

**Frontend tasks**:
- Create `src/api/expenses.ts`: `getExpenses(params: { page: number; per_page: number }): Promise<ApiListResponse<Expense>>`.
- Update `src/pages/ExpensesPage.tsx`:
  - Render a table with columns: Date, Category, Description, Amount, Actions (edit/delete buttons).
  - Display amounts in major currency units with locale formatting (NF-17).
  - Pagination controls (Previous/Next, page number display) at the bottom.
  - Loading state while fetching.
  - Empty state when no expenses exist ("No expenses recorded yet").

**Acceptance criteria**:
- Given 25 expenses exist, when the user views the expense list with default pagination (20 per page), then 20 expenses are shown on page 1 with a "Next" button.
- Given page 2 is requested, then the remaining 5 expenses are shown.
- Given no expenses exist, then an empty state message is displayed.
- Given the list is loading, then a loading indicator is shown.
- Given 10,000 expense records in the database, when the list endpoint is called, then the response time is under 500ms on a local network (NF-01, NF-04).

---

### E2-S3: Filter Expenses by Date Range

| Field | Value |
|---|---|
| **Ticket ID** | E2-S3 |
| **Title** | Filter expenses by date range |
| **Epic** | Epic 2 -- Expense Tracking |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E2-S2 |

**Description**: Add date range filtering to the expense list. The user selects a start date and end date, and only expenses within that range are shown.

**Backend tasks**:
- Extend `ExpenseService.List()` to accept optional `dateFrom` and `dateTo` parameters.
- Update `GET /api/expenses` to accept query params `date_from=2026-01-01&date_to=2026-01-31`. Validate date format (YYYY-MM-DD). If `date_from` is after `date_to`, return 400.
- Ensure the `idx_expenses_date` index is used by the query.

**Frontend tasks**:
- Create `src/components/filters/DateRangeFilter.tsx`: two date inputs (From/To) with an Apply button and a Clear button.
- Integrate the filter into `ExpensesPage.tsx` above the table.
- On filter apply, update the API call params and reset to page 1.
- Persist filter state in URL query params (e.g., `/expenses?date_from=2026-01-01&date_to=2026-01-31`) so the view is shareable/bookmarkable.

**Acceptance criteria**:
- Given expenses exist in January and February, when the user filters by date range 2026-01-01 to 2026-01-31, then only January expenses are shown.
- Given the user clears the filter, then all expenses are shown again.
- Given the user applies a filter and refreshes the page, then the filter persists via URL query params.

---

### E2-S4: Filter Expenses by Category

| Field | Value |
|---|---|
| **Ticket ID** | E2-S4 |
| **Title** | Filter expenses by category |
| **Epic** | Epic 2 -- Expense Tracking |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E2-S2 |

**Description**: Add category filtering to the expense list. The user selects one or more categories from a dropdown, and only matching expenses are shown. Note: promoted to Must to align with the M1 key story list in requirements.md which includes E2-S4.

**Backend tasks**:
- Extend `ExpenseService.List()` to accept optional `categoryID` parameter (or multiple: `category_ids=1,2,3`).
- Update `GET /api/expenses` to accept `category_id=1` or `category_ids=1,2,3`.

**Frontend tasks**:
- Create `src/components/filters/CategoryFilter.tsx`: a dropdown (or multi-select) populated from `GET /api/categories?type=expense`.
- Integrate alongside the date range filter in `ExpensesPage.tsx`.
- Update URL query params to include `category_id`.

**Acceptance criteria**:
- Given expenses exist in Food and Transport categories, when the user filters by Food, then only Food expenses are shown.
- Given the user selects multiple categories, then expenses from all selected categories are shown.
- Given the user clears the category filter, then all expenses are shown.

---

### E2-S5: Edit Expense

| Field | Value |
|---|---|
| **Ticket ID** | E2-S5 |
| **Title** | Edit an existing expense |
| **Epic** | Epic 2 -- Expense Tracking |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | XS |
| **Dependencies** | E2-S1 |

**Description**: Allow the user to edit any field of an existing expense. The edit form is pre-populated with the current values.

**Backend tasks**:
- Add to `ExpenseService`:
  - `GetByID(ctx, id int64) (*Expense, error)` -- returns single expense or `ErrNotFound`.
  - `Update(ctx, id int64, expense *Expense) error` -- validate same as Create, update row, set `updated_at = CURRENT_TIMESTAMP`.
- Add to expense handler:
  - `GET /api/expenses/:id` -- returns single expense. 404 if not found.
  - `PUT /api/expenses/:id` -- accepts same body as POST, updates the record. Returns 200 with updated expense. 404 if not found. 400 on validation errors.

**Frontend tasks**:
- Reuse `ExpenseForm.tsx` in edit mode: accept an optional `expense` prop to pre-populate fields.
- Add an "Edit" button/icon to each row in the expense table.
- On click, open the form with pre-populated data.
- On save, call `PUT /api/expenses/:id` and refresh the list.

**Acceptance criteria**:
- Given an existing expense, when the user clicks Edit, then the form opens with all current values pre-filled.
- Given the user changes the amount and saves, then the updated amount is reflected in the list and database.
- Given the expense ID does not exist, when the API is called, then it returns 404.

---

### E2-S6: Delete Expense

| Field | Value |
|---|---|
| **Ticket ID** | E2-S6 |
| **Title** | Delete an expense record |
| **Epic** | Epic 2 -- Expense Tracking |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | XS |
| **Dependencies** | E2-S1 |

**Description**: Allow the user to delete an expense with a confirmation dialog. Deletion removes the database record. Note: since the `attachments` table does not exist in M1, no cascade delete of attachments is needed. In M2, E7-S5 will retroactively add cascade behavior to this delete method.

**Backend tasks**:
- Add to `ExpenseService`:
  - `Delete(ctx, id int64) error` -- delete the row. Return `ErrNotFound` if it does not exist.
- Add to expense handler:
  - `DELETE /api/expenses/:id` -- returns 204 on success. 404 if not found.

**Frontend tasks**:
- Add a "Delete" button/icon to each row in the expense table.
- Create `src/components/ConfirmDialog.tsx` -- reusable confirmation modal with a message, Confirm button, and Cancel button (NF-15).
- On delete click, show ConfirmDialog: "Are you sure you want to delete this expense?"
- On confirm, call `DELETE /api/expenses/:id`, remove from list, show success toast.
- On cancel, close dialog with no changes.

**Acceptance criteria**:
- Given an existing expense, when the user clicks Delete and confirms, then the expense is removed from the list and the database.
- Given the user clicks Delete and cancels, then no data is modified.
- Given the expense ID does not exist, when the API is called, then it returns 404.

---

### E2-S8: Running Total for Filtered Expenses

| Field | Value |
|---|---|
| **Ticket ID** | E2-S8 |
| **Title** | Display running total for filtered expense view |
| **Epic** | Epic 2 -- Expense Tracking |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | XS |
| **Dependencies** | E2-S3 |

**Description**: Show the total amount for the currently applied filters at the top of the expense list. This gives the user an at-a-glance spending summary for the filtered period.

**Backend tasks**:
- The `total` field in the list response already represents count. Add a `total_amount` field to the list response: `{"data": [...], "total": N, "per_page": 20, "page": 1, "total_amount": 1500000}`.
- Compute `SUM(amount)` alongside the `COUNT(*)` using the same WHERE clause as the list query.

**Frontend tasks**:
- Display the total amount in a summary bar above the expense table: "Total: 1,500,000 VND" (formatted per NF-17).
- Update whenever filters change.

**Acceptance criteria**:
- Given 3 expenses totaling 500,000 in the current filter, when the list loads, then "Total: 500,000 VND" (or configured currency) is displayed above the table.
- Given the filter changes, then the total updates to reflect only the filtered expenses.

---

### E3-S1: Create Income

| Field | Value |
|---|---|
| **Ticket ID** | E3-S1 |
| **Title** | Create a new income record |
| **Epic** | Epic 3 -- Income Management |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E1-S1, E5-S1 |

**Description**: Allow the user to create an income record with amount, date, source/category, and optional description. Mirrors the expense creation flow but uses income categories.

**Backend tasks**:
- Create migration `006_create_incomes.up.sql`:
  ```sql
  CREATE TABLE incomes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount INTEGER NOT NULL CHECK(amount > 0),
    date DATE NOT NULL,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    description TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  CREATE INDEX idx_incomes_date ON incomes(date);
  CREATE INDEX idx_incomes_category_id ON incomes(category_id);
  ```
- Update `internal/models/models.go`: update `Income` struct to use `CategoryID int64` instead of `Source string`. Remove `Currency` field.
- Create `internal/services/income_service.go`:
  - `IncomeService` with `Create`, `List`, `GetByID`, `Update`, `Delete` methods (same pattern as expense).
  - Validate category_id is of type "income".
- Create `internal/handlers/income_handler.go`:
  - `POST /api/incomes` -- accepts `{"amount": 10000000, "date": "2026-04-26", "category_id": 10, "description": "April salary"}`. Returns 201.
  - `GET /api/incomes?page=1&per_page=20&date_from=...&date_to=...&category_id=...` -- paginated, filterable list with `total_amount`.
  - `GET /api/incomes/:id` -- single income.
  - `PUT /api/incomes/:id` -- update income.
  - `DELETE /api/incomes/:id` -- delete income. Returns 204.

**Frontend tasks**:
- Create `src/types/income.ts`: `Income { id: number; amount: number; date: string; category_id: number; category_name?: string; description: string; created_at: string; updated_at: string }`.
- Create `src/api/incomes.ts`: `createIncome`, `getIncomes`, `getIncome`, `updateIncome`, `deleteIncome`.
- Create `src/components/income/IncomeForm.tsx`: same pattern as ExpenseForm but loads income categories.
- Implement `src/pages/IncomePage.tsx`: same structure as ExpensesPage (table, pagination, filters, running total, add/edit/delete).

**Acceptance criteria**:
- Given a valid amount and date, when the user saves the income record, then it appears in the income list and is persisted.
- Given an amount of zero, when the user tries to save, then the system rejects it with an error message.
- Given a category of type "expense" is used, when the API validates, then it returns 400.

---

### E3-S2: List Income (Paginated)

| Field | Value |
|---|---|
| **Ticket ID** | E3-S2 |
| **Title** | View paginated list of all income records |
| **Epic** | Epic 3 -- Income Management |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E3-S1 |

**Description**: Display all income records in a paginated table with columns for date, source/category, description, and amount. Same UX pattern as expense list.

**Backend tasks**: Already covered in E3-S1 (`GET /api/incomes` endpoint).

**Frontend tasks**: Already covered in E3-S1 (IncomePage with table, pagination).

**Acceptance criteria**:
- Given 25 income records exist, when the user views the income list with default pagination (20 per page), then 20 records are shown with pagination controls.
- Given no income records exist, then an empty state message is displayed.
- Given 10,000 income records in the database, when the list endpoint is called, then the response time is under 500ms on a local network (NF-01, NF-04).

---

### E3-S3: Filter Income by Date Range

| Field | Value |
|---|---|
| **Ticket ID** | E3-S3 |
| **Title** | Filter income by date range |
| **Epic** | Epic 3 -- Income Management |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E3-S2 |

**Description**: Add date range filtering to the income list, reusing the same DateRangeFilter component from expenses.

**Backend tasks**: Already covered in E3-S1 (`date_from`/`date_to` query params on `GET /api/incomes`).

**Frontend tasks**:
- Reuse `DateRangeFilter.tsx` in `IncomePage.tsx`.
- Persist filter state in URL query params.

**Acceptance criteria**:
- Given income records in January and February, when the user filters Jan 1-31, then only January records are shown.
- Given the filter is cleared, then all records are shown.

---

### E3-S4: Edit and Delete Income

| Field | Value |
|---|---|
| **Ticket ID** | E3-S4 |
| **Title** | Edit or delete an income record |
| **Epic** | Epic 3 -- Income Management |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | XS |
| **Dependencies** | E3-S1 |

**Description**: Allow the user to edit any field of an income record or delete it with confirmation. Reuses patterns from expense edit/delete. Note: since the `attachments` table does not exist in M1, no cascade delete of attachments is needed. In M2, E7-S5 will retroactively add cascade behavior to this delete method.

**Backend tasks**: Already covered in E3-S1 (`PUT /api/incomes/:id`, `DELETE /api/incomes/:id`).

**Frontend tasks**:
- Reuse `IncomeForm.tsx` in edit mode.
- Add Edit and Delete action buttons to income table rows.
- Reuse `ConfirmDialog.tsx` for delete confirmation.

**Acceptance criteria**:
- Given an existing income record, when the user edits and saves, then the changes are reflected in the list.
- Given the user deletes an income record and confirms, then it is removed from the list and database.
- Given the user cancels deletion, then no data is modified.

---

### E4-S1: Create Invoice/Bill

| Field | Value |
|---|---|
| **Ticket ID** | E4-S1 |
| **Title** | Create an invoice/bill record |
| **Epic** | Epic 4 -- Invoice & Bill Management |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | M |
| **Dependencies** | E1-S1, M0-01 |

**Description**: Allow the user to create an invoice record with vendor name, amount, issue date, due date, status (unpaid/paid/overdue), and optional description. Invoices do not use the categories system -- they are organized by vendor and status.

**Backend tasks**:
- Create migration `007_create_invoices.up.sql`:
  ```sql
  CREATE TABLE invoices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    vendor_name TEXT NOT NULL,
    amount INTEGER NOT NULL CHECK(amount > 0),
    issue_date DATE NOT NULL,
    due_date DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'unpaid' CHECK(status IN ('unpaid', 'paid', 'overdue')),
    description TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  CREATE INDEX idx_invoices_status ON invoices(status);
  CREATE INDEX idx_invoices_due_date ON invoices(due_date);
  CREATE INDEX idx_invoices_issue_date ON invoices(issue_date);
  ```
- Update `internal/models/models.go`: update `Invoice` struct: remove `Currency`, `AttachmentID`; add `IssueDate`, rename fields to match schema. `VendorName string`, `Amount int64`, `IssueDate time.Time`, `DueDate time.Time`, `Status string`, `Description string`.
- Create `internal/services/invoice_service.go`:
  - `InvoiceService` with `Create`, `List`, `GetByID`, `Update`, `Delete` methods.
  - `Create` validates: vendor_name not empty, amount > 0, due_date >= issue_date.
- Create `internal/handlers/invoice_handler.go`:
  - `POST /api/invoices` -- accepts `{"vendor_name": "...", "amount": 5000000, "issue_date": "2026-04-01", "due_date": "2026-04-30", "status": "unpaid", "description": "..."}`. Returns 201.
  - Validate `due_date >= issue_date`, return 400 if not.

**Frontend tasks**:
- Create `src/types/invoice.ts`: `Invoice { id: number; vendor_name: string; amount: number; issue_date: string; due_date: string; status: 'unpaid' | 'paid' | 'overdue'; description: string; created_at: string; updated_at: string }`.
- Create `src/api/invoices.ts`: `createInvoice`, `getInvoices`, `getInvoice`, `updateInvoice`, `deleteInvoice`.
- Create `src/components/invoices/InvoiceForm.tsx`:
  - Fields: vendor_name (text), amount (number), issue_date (date), due_date (date), status (select: unpaid/paid/overdue), description (textarea).
  - Validation: vendor_name required, amount > 0, issue_date and due_date required, due_date >= issue_date.
- Implement `src/pages/InvoicesPage.tsx` with "Add Invoice" button.

**Acceptance criteria**:
- Given all required fields, when the user saves the invoice, then it appears in the invoice list under the correct status.
- Given a missing vendor name or amount, then validation rejects the form with field-level errors.
- Given due_date earlier than issue_date, then the system rejects it with "Due date must be on or after issue date".

---

### E4-S2: List Invoices with Status Filter

| Field | Value |
|---|---|
| **Ticket ID** | E4-S2 |
| **Title** | View invoices in a paginated list, filterable by status |
| **Epic** | Epic 4 -- Invoice & Bill Management |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E4-S1 |

**Description**: Display all invoices in a paginated list with tabs or filter for status (All, Unpaid, Paid, Overdue). Each row shows vendor, amount, issue date, due date, and status.

**Backend tasks**:
- Add to `InvoiceService`:
  - `List(ctx, page, perPage int, status string, dateFrom, dateTo string) ([]Invoice, total int64, totalAmount int64, err error)` -- filter by status, paginate.
- Add to invoice handler:
  - `GET /api/invoices?page=1&per_page=20&status=unpaid` -- filterable by status. Returns `total_amount` for the filtered set.

**Frontend tasks**:
- Update `InvoicesPage.tsx`:
  - Status tabs or toggle buttons: All | Unpaid | Paid | Overdue.
  - Table columns: Vendor, Amount, Issue Date, Due Date, Status (with color badge), Actions.
  - Pagination controls.
  - Status badges: Unpaid (yellow), Paid (green), Overdue (red).

**Acceptance criteria**:
- Given invoices with mixed statuses, when the user clicks the "Unpaid" tab, then only unpaid invoices are shown.
- Given the "All" tab is selected, then all invoices are shown regardless of status.
- Given 10,000 invoice records in the database, when the list endpoint is called, then the response time is under 500ms on a local network (NF-01, NF-04).
- Persist the status filter in URL query params (e.g., `/invoices?status=unpaid`) so the view is shareable/bookmarkable.

---

### E4-S3: Mark Invoice as Paid

| Field | Value |
|---|---|
| **Ticket ID** | E4-S3 |
| **Title** | Quick-action: mark an invoice as paid |
| **Epic** | Epic 4 -- Invoice & Bill Management |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | XS |
| **Dependencies** | E4-S1 |

**Description**: Provide a quick action to mark an unpaid/overdue invoice as "paid" without opening the full edit form. This is a common action that should be frictionless.

**Backend tasks**:
- Add to invoice handler:
  - `PATCH /api/invoices/:id/status` -- accepts `{"status": "paid"}`. Validates that the status transition is valid (only unpaid/overdue -> paid). Returns 200 with updated invoice.

**Frontend tasks**:
- Add a "Mark as Paid" button to each unpaid/overdue invoice row.
- On click, show a confirmation dialog (NF-15): "Mark this invoice from [Vendor] as paid?"
- On confirm, call `PATCH /api/invoices/:id/status`, update the row in-place without full page reload.
- On success, the status badge changes to green "Paid".

**Acceptance criteria**:
- Given an unpaid invoice, when the user clicks "Mark as Paid" and confirms, then the status updates to "paid" immediately in the list.
- Given a paid invoice, then the "Mark as Paid" button is not shown.
- Given a paid invoice, when `PATCH /api/invoices/:id/status` is called with `{"status": "paid"}`, then the server returns 422 Unprocessable Entity with `{"error": "Invoice is already paid"}`.
- Given any invoice, when `PATCH /api/invoices/:id/status` is called with `{"status": "unpaid"}`, then the server returns 400 Bad Request with `{"error": "Only 'paid' is a valid target status via this endpoint"}`.

---

### E4-S4: Overdue Invoice Highlighting

| Field | Value |
|---|---|
| **Ticket ID** | E4-S4 |
| **Title** | Highlight overdue invoices in the list |
| **Epic** | Epic 4 -- Invoice & Bill Management |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E4-S1 |

**Description**: Invoices that are unpaid and past their due date should be visually highlighted as overdue. The status should be computed/updated when the list is fetched.

**Backend tasks**:
- Create a function `InvoiceService.UpdateOverdueStatuses(ctx) (int64, error)` -- execute `UPDATE invoices SET status = 'overdue', updated_at = CURRENT_TIMESTAMP WHERE status = 'unpaid' AND due_date < date('now')`. Return the count of updated rows. This is the single source of truth for overdue transitions (IV-08).
- Call `UpdateOverdueStatuses()`:
  1. On server startup.
  2. At the start of `InvoiceService.List()`, before returning results.
- Log the count of updated invoices: "Updated N invoices to overdue status".
- Decision: use the persisted UPDATE approach so the status is always correct regardless of how data is accessed.

**Frontend tasks**:
- Style overdue invoice rows: red status badge, optional row background tint.
- Show a warning icon next to overdue invoices.

**Acceptance criteria**:
- Given an invoice with status "unpaid" and due_date in the past, when the page loads, then the invoice is displayed as "overdue" with red styling.
- Given an invoice with due_date in the future, then its status remains "unpaid" with yellow styling.
- Given the transition to overdue happens server-side, then no manual user action is required.
- Given the server starts with 3 unpaid invoices past due, when the startup completes, then all 3 are updated to "overdue" and the server logs "Updated 3 invoices to overdue status".

---

### E4-S6: Edit and Delete Invoice

| Field | Value |
|---|---|
| **Ticket ID** | E4-S6 |
| **Title** | Edit or delete an invoice record |
| **Epic** | Epic 4 -- Invoice & Bill Management |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | XS |
| **Dependencies** | E4-S1 |

**Description**: Allow the user to edit all fields of an invoice or delete it with confirmation.

**Backend tasks**:
- Add to invoice handler:
  - `GET /api/invoices/:id` -- returns single invoice.
  - `PUT /api/invoices/:id` -- updates all fields with same validation as Create.
  - `DELETE /api/invoices/:id` -- deletes invoice. Returns 204.

**Frontend tasks**:
- Reuse `InvoiceForm.tsx` in edit mode.
- Add Edit and Delete action buttons to each invoice row.
- Reuse `ConfirmDialog.tsx` for delete confirmation.

**Acceptance criteria**:
- Given an existing invoice, when the user edits the vendor name and saves, then the change is reflected in the list.
- Given the user deletes an invoice and confirms, then it is removed.
- Given the user cancels deletion, then no data is modified.

---

### E4-S8: Outstanding Invoice Total

| Field | Value |
|---|---|
| **Ticket ID** | E4-S8 |
| **Title** | Display total outstanding amount for invoices |
| **Epic** | Epic 4 -- Invoice & Bill Management |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | XS |
| **Dependencies** | E4-S2 |

**Description**: Show the total outstanding amount (sum of unpaid + overdue invoices) at the top of the invoice list page. This gives an at-a-glance view of liabilities.

**Backend tasks**:
- Add a dedicated endpoint or include in the list response:
  - `GET /api/invoices/stats` -- returns `{"total_outstanding": 15000000, "unpaid_count": 3, "overdue_count": 2}`. Note: this endpoint uses `/stats` instead of `/summary` to avoid a routing conflict with `GET /api/invoices/:id` (Go's `net/http` mux would match "summary" as an `:id` parameter).
- Alternatively, include `total_amount` in the existing list response when filtered by status.

**Frontend tasks**:
- Display a summary bar at the top of `InvoicesPage.tsx`: "Outstanding: 15,000,000 VND (3 unpaid, 2 overdue)".
- Update when filters change or invoices are modified.

**Acceptance criteria**:
- Given 3 unpaid invoices totaling 10,000,000 and 2 overdue totaling 5,000,000, when the page loads, then "Outstanding: 15,000,000" is shown.
- Given an invoice is marked as paid, then the outstanding total decreases accordingly.

---

### E6-S1: Dashboard Summary Cards

| Field | Value |
|---|---|
| **Ticket ID** | E6-S1 |
| **Title** | Dashboard with income, expense, and balance summary cards |
| **Epic** | Epic 6 -- Dashboard & Reports |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | M |
| **Dependencies** | E2-S1, E3-S1 |

**Description**: The main dashboard page displays three summary cards showing total income, total expenses, and net balance (income minus expenses) for the current month. This is the first thing the user sees after login.

**Backend tasks**:
- Create `internal/services/dashboard_service.go`:
  - `DashboardService` with `db *sql.DB`.
  - `GetSummary(ctx, dateFrom, dateTo string) (*DashboardSummary, error)`:
    - Query `SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE date BETWEEN ? AND ?`.
    - Query `SELECT COALESCE(SUM(amount), 0) FROM incomes WHERE date BETWEEN ? AND ?`.
    - Compute net balance = total_income - total_expenses.
    - Return `DashboardSummary { TotalIncome int64, TotalExpenses int64, NetBalance int64, DateFrom string, DateTo string }`.
- Create `internal/handlers/dashboard_handler.go`:
  - `GET /api/dashboard/summary?date_from=2026-04-01&date_to=2026-04-30` -- returns `{"data": {"total_income": ..., "total_expenses": ..., "net_balance": ..., "date_from": "...", "date_to": "..."}}`.
  - Default to current month if no dates provided.

**Frontend tasks**:
- Create `src/api/dashboard.ts`: `getDashboardSummary(dateFrom, dateTo): Promise<DashboardSummary>`.
- Create `src/types/dashboard.ts`: `DashboardSummary { total_income: number; total_expenses: number; net_balance: number }`.
- Create `src/components/dashboard/SummaryCard.tsx` -- a card component with title, amount (formatted), and optional color/icon. Positive balance = green, negative = red.
- Update `src/pages/DashboardPage.tsx`:
  - Three `SummaryCard` components in a row: Total Income (green), Total Expenses (red), Net Balance (green/red based on sign).
  - Default period: current calendar month.
  - Loading skeleton while data is being fetched.

**Acceptance criteria**:
- Given the current month is April 2026, when the dashboard loads, then it shows three summary cards with income, expenses, and balance computed from records dated in April 2026.
- Given the user has no records for the current month, then all cards show 0.
- Given the data is loading, then skeleton/placeholder cards are shown.

---

### M1-01: Service Layer Unit Test Coverage

| Field | Value |
|---|---|
| **Ticket ID** | M1-01 |
| **Title** | Unit test coverage for service-layer business logic |
| **Epic** | Infrastructure |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | M |
| **Dependencies** | E2-S1, E3-S1, E4-S1, E5-S1 |

**Description**: Implement unit tests for critical service-layer business logic in `internal/services/`. This addresses NF-20 (unit test coverage requirement). Without these tests, the CI pipeline runs an empty test suite that always passes.

**Backend tasks**:
- Write unit tests for the following service methods (minimum):
  - `ExpenseService.Create` -- rejects zero/negative amounts; rejects invalid category_id; rejects expense-type mismatch (income category used for expense).
  - `IncomeService.Create` -- rejects zero/negative amounts; rejects category type mismatch.
  - `InvoiceService.Create` -- rejects due_date before issue_date; rejects empty vendor name.
  - `InvoiceService.UpdateOverdueStatuses` -- only updates unpaid invoices past due date; does not update paid invoices; does not update invoices with future due dates.
  - `CategoryService.Delete` -- reassigns transactions to Uncategorized (tested via E5-S4); rejects deletion of default categories.
- Target minimum 70% test coverage for `internal/services/` package.
- Use SQLite in-memory database for tests (`:memory:` connection string).
- Tests must be runnable via `go test ./internal/services/...`.

**Frontend tasks**: None.

**Acceptance criteria**:
- Given `go test ./internal/services/...` is run, then all listed test cases pass.
- Given `go test -cover ./internal/services/...`, then coverage is at least 70%.
- Given a regression is introduced (e.g., removing the amount > 0 check), then at least one test fails.

---

### E6-S6: Invoice Summary on Dashboard

| Field | Value |
|---|---|
| **Ticket ID** | E6-S6 |
| **Title** | Display unpaid/overdue invoice count and total on dashboard |
| **Epic** | Epic 6 -- Dashboard & Reports |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E6-S1, E4-S1 |

**Description**: Add an invoice summary section to the dashboard showing the count and total amount of unpaid and overdue invoices. This ensures outstanding bills are always visible.

**Backend tasks**:
- Extend `DashboardService.GetSummary()` or add a separate method:
  - `GetInvoiceSummary(ctx) (*InvoiceSummary, error)`:
    - Query `SELECT status, COUNT(*), COALESCE(SUM(amount), 0) FROM invoices WHERE status IN ('unpaid', 'overdue') GROUP BY status`.
    - Return `InvoiceSummary { UnpaidCount int, UnpaidAmount int64, OverdueCount int, OverdueAmount int64, TotalOutstanding int64 }`.
- Include in `GET /api/dashboard/summary` response or create `GET /api/dashboard/invoices`.

**Frontend tasks**:
- Create `src/components/dashboard/InvoiceSummaryCard.tsx` -- shows unpaid count/amount and overdue count/amount with a link to the Invoices page.
- Add to `DashboardPage.tsx` below the income/expense summary cards.

**Acceptance criteria**:
- Given 3 unpaid invoices totaling 10,000,000 and 1 overdue invoice totaling 5,000,000, when the dashboard loads, then the invoice summary shows "4 outstanding invoices - 15,000,000 VND (3 unpaid, 1 overdue)".
- Given no outstanding invoices, then the summary shows "No outstanding invoices".

---

## Milestone 2 -- File Attachments & Enhanced Lists

**Goal**: Add file upload/download via configurable storage backend (local or S3), inline previews, custom categories, and enhanced sorting. Auto-overdue for invoices.

**Exit criteria**: User can attach files to any record, preview images inline, download files, manage custom categories, and invoices auto-transition to overdue.

---

### E7-S1: File Upload Infrastructure

| Field | Value |
|---|---|
| **Ticket ID** | E7-S1 |
| **Title** | Implement file upload/download via configurable storage backend with database tracking |
| **Epic** | Epic 7 -- File Attachments |
| **Milestone** | M2 |
| **Priority** | Must |
| **Size** | L |
| **Dependencies** | M0-01, M0-02 |

**Description**: Build the core file attachment system. Files are uploaded to the configured storage backend (local filesystem or S3-compatible) and tracked in an `attachments` table. This is a shared infrastructure ticket that all entity-specific attachment features build on. The implementation must use a **storage interface** so the backend is not hard-coded to MinIO.

**Backend tasks**:
- Reuse the **`ObjectStore`** interface in `internal/storage/storage.go` and implementations **`LocalStorage`** (`local.go`) and **`MinIOStorage`** (`minio.go`, S3-compatible) from M0 — extend only if attachment flows need additional methods.
- **Do not** introduce a parallel storage abstraction or duplicate factory logic unless consolidating: today `cmd/server/main.go` branches on `cfg.StorageType` and constructs the store; E7-S1 may extract `internal/storage.NewObjectStore(cfg *config.Config) (ObjectStore, error)` only if attachment init or tests need a single entrypoint.
- `AttachmentService` must accept `storage.ObjectStore` (interface), not a concrete MinIO client.
- Create migration `008_create_attachments.up.sql`:
  ```sql
  CREATE TABLE attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type TEXT NOT NULL CHECK(entity_type IN ('expense', 'income', 'invoice')),
    entity_id INTEGER NOT NULL,
    filename TEXT NOT NULL,
    mime_type TEXT NOT NULL CHECK(mime_type IN ('application/pdf', 'image/jpeg', 'image/png')),
    size_bytes INTEGER NOT NULL,
    storage_key TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  CREATE INDEX idx_attachments_entity ON attachments(entity_type, entity_id);
  ```
- Create `internal/models/attachment.go`: `Attachment { ID, EntityType, EntityID, Filename, MimeType, SizeBytes, StorageKey, CreatedAt }`.
- Create `internal/services/attachment_service.go`:
  - `AttachmentService` with `db *sql.DB` and `store storage.ObjectStore` (interface, not concrete type).
  - `Upload(ctx, entityType string, entityID int64, file multipart.File, header *multipart.FileHeader) (*Attachment, error)`:
    - Validate mime type (PDF, JPEG, PNG only -- FA-01).
    - Validate file size <= 10MB (FA-02 server-side check).
    - Generate a unique storage key: `{entityType}/{entityID}/{uuid}_{filename}`.
    - Upload via `storage.Upload(...)`.
    - Insert attachment record in DB within a transaction (NF-11).
    - If storage upload fails, do not insert DB record (NF-12).
    - If DB insert fails after storage upload, delete from storage (cleanup).
  - `ListByEntity(ctx, entityType string, entityID int64) ([]Attachment, error)`.
  - `Delete(ctx, attachmentID int64) error` -- delete from both DB and storage.
  - `DeleteByEntity(ctx, entityType string, entityID int64) error` -- delete all attachments for a record (used when parent record is deleted -- FA-05).
- Create `internal/handlers/attachment_handler.go`:
  - `POST /api/attachments` -- multipart form: `entity_type`, `entity_id`, `file`. Returns 201 with attachment metadata.
  - `GET /api/attachments?entity_type=expense&entity_id=1` -- list attachments for an entity.
  - `DELETE /api/attachments/:id` -- delete single attachment. Returns 204.

**Frontend tasks**:
- Create `src/types/attachment.ts`: `Attachment { id: number; entity_type: string; entity_id: number; filename: string; mime_type: string; size_bytes: number; storage_key: string; created_at: string }`.
- Create `src/api/attachments.ts`: `uploadAttachment(entityType, entityId, file)`, `getAttachments(entityType, entityId)`, `deleteAttachment(id)`.
- Create `src/components/attachments/FileUpload.tsx`:
  - Drag-and-drop area or file input button.
  - Client-side validation before upload: file type (PDF, JPEG, PNG) and size (<= 10MB -- FA-02 client-side).
  - Upload progress indicator.
  - Error display for rejected files.
- Create `src/components/attachments/AttachmentList.tsx`:
  - Shows list of attached files with filename, size, and actions (download, delete).
  - Delete button with confirmation dialog.

**Acceptance criteria**:
- Given `STORAGE_TYPE=local` and a valid JPEG file under 10MB, when the user uploads it for an expense, then the file is written to `LOCAL_STORAGE_PATH` and an attachment record is created in the database.
- Given `STORAGE_TYPE=s3` and a valid JPEG file under 10MB, when the user uploads it for an expense, then the file is stored in the MinIO bucket and an attachment record is created in the database.
- Given a file over 10MB, when the user attempts to upload, then the client rejects it before sending the request with "File must be under 10 MB".
- Given a `.exe` file, when the user attempts to upload, then it is rejected with "Only PDF, JPEG, and PNG files are allowed".
- Given a record with two attached files, when the parent record is deleted, then both files are removed from the storage backend (verified by absent storage keys).

---

### E7-S2: File Size Validation

| Field | Value |
|---|---|
| **Ticket ID** | E7-S2 |
| **Title** | Enforce 10MB maximum file size |
| **Epic** | Epic 7 -- File Attachments |
| **Milestone** | M2 |
| **Priority** | Must |
| **Size** | XS |
| **Dependencies** | E7-S1 |

**Description**: Ensure files over 10MB are rejected both client-side (before upload) and server-side (as a safety net). This is partially covered in E7-S1 but called out separately per the requirements.

**Backend tasks**:
- In the upload handler, use `http.MaxBytesReader` to limit request body to 10MB + overhead.
- In `AttachmentService.Upload`, check `header.Size <= 10*1024*1024` before uploading.
- Return 413 (Payload Too Large) if exceeded.

**Frontend tasks**:
- In `FileUpload.tsx`, check `file.size <= 10 * 1024 * 1024` before making the upload request.
- Display error: "File exceeds the 10 MB limit" immediately without sending the request.

**Acceptance criteria**:
- Given a file of 11MB, when the user selects it, then the frontend shows an error and does not initiate an upload.
- Given a crafted request bypassing the frontend that sends a 15MB file, then the server returns 413.

---

### E7-S3: File Download

| Field | Value |
|---|---|
| **Ticket ID** | E7-S3 |
| **Title** | Download attached files |
| **Epic** | Epic 7 -- File Attachments |
| **Milestone** | M2 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E7-S1 |

**Description**: Allow the user to download any attached file. The backend proxies the download from the storage backend (NF-07: raw storage objects must not be publicly accessible). Works for both `local` and `s3` backends — the handler calls `storage.Download(key)` and streams the result to the client.

**Backend tasks**:
- Add to attachment handler:
  - `GET /api/attachments/:id/download` -- look up attachment by ID, call `storage.Download(storageKey)`, and stream the file to the client with correct `Content-Type` and `Content-Disposition: attachment; filename="..."` headers.
- Alternatively, generate a pre-signed URL with short expiry (e.g., 15 minutes) and redirect the client.
- Decision: use proxy approach for simplicity and security.

**Frontend tasks**:
- In `AttachmentList.tsx`, add a "Download" button/link for each file.
- On click, trigger a download via `window.open` or creating an `<a>` tag with `download` attribute pointing to `/api/attachments/:id/download`.

**Acceptance criteria**:
- Given an attached PDF file, when the user clicks Download, then the browser downloads the file with the correct filename and content type.
- Given an attachment ID that does not exist, when the download is requested, then the server returns 404.

---

### E7-S4: Image Thumbnail Preview

| Field | Value |
|---|---|
| **Ticket ID** | E7-S4 |
| **Title** | Display inline thumbnail for image attachments |
| **Epic** | Epic 7 -- File Attachments |
| **Milestone** | M2 |
| **Priority** | Should |
| **Size** | S |
| **Dependencies** | E7-S1, E7-S3 |

**Description**: Show a small thumbnail preview for JPEG and PNG attachments so the user can visually confirm they uploaded the correct file. PDFs show a file icon instead.

**Backend tasks**:
- Add to attachment handler:
  - `GET /api/attachments/:id/preview` -- for images (JPEG/PNG), stream the image with appropriate Content-Type for inline display. For PDFs, optionally return a generic PDF icon or the PDF itself for the browser's built-in viewer.
- This can be the same as the download endpoint but with `Content-Disposition: inline` instead of `attachment`.

**Frontend tasks**:
- Update `AttachmentList.tsx`:
  - For `image/jpeg` and `image/png` attachments, render an `<img>` tag with `src="/api/attachments/:id/preview"` at thumbnail size (e.g., 80x80px with `object-fit: cover`).
  - For `application/pdf` attachments, show a PDF icon with the filename.
  - On clicking an image thumbnail, open a larger preview (modal or lightbox).

**Acceptance criteria**:
- Given a JPEG attachment, when the attachment list loads, then a thumbnail preview of the image is shown.
- Given a PDF attachment, when the attachment list loads, then a PDF icon is shown with the filename.
- Given the user clicks an image thumbnail, then a larger preview is displayed.

---

### E7-S5: Cascade Delete Attachments

| Field | Value |
|---|---|
| **Ticket ID** | E7-S5 |
| **Title** | Delete attachments when parent record is deleted |
| **Epic** | Epic 7 -- File Attachments |
| **Milestone** | M2 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E7-S1, E2-S6, E3-S4, E4-S6 |

**Description**: When an expense, income, or invoice is deleted, all associated attachments must be deleted from both the database and the storage backend (FA-05). This must be done within a transaction and works for both `local` and `s3` backends.

**Backend tasks**:
- Update `ExpenseService.Delete()`, `IncomeService.Delete()`, `InvoiceService.Delete()`:
  - Before deleting the record, call `AttachmentService.DeleteByEntity(ctx, entityType, entityID)`.
  - Wrap the entire operation in a DB transaction (NF-11).
  - Delete from storage first, then delete DB records. If storage delete fails, log the error but still delete the DB record (orphaned files are less harmful than orphaned DB records -- handled by D3 audit endpoint).
- Implement `AttachmentService.DeleteByEntity()`:
  - Query all attachment storage keys for the entity.
  - Delete each via `storage.Delete(key)`.
  - Delete all attachment DB records.

**Frontend tasks**: No additional frontend work needed -- existing delete flows already trigger the backend delete.

**Acceptance criteria**:
- Given an expense with 2 attached files, when the user deletes the expense, then both files are removed from the storage backend and the attachment records are deleted from the database.
- Given the storage backend is temporarily unreachable during delete, then the DB records are still deleted and the server logs a warning about orphaned files.

---

### E2-S7: Attach Receipts to Expenses

| Field | Value |
|---|---|
| **Ticket ID** | E2-S7 |
| **Title** | Attach receipt images or PDFs to expenses |
| **Epic** | Epic 2 -- Expense Tracking |
| **Milestone** | M2 |
| **Priority** | Should |
| **Size** | M |
| **Dependencies** | E2-S1, E7-S1 |

**Description**: Integrate the file attachment system with the expense detail view. The user can upload and view receipts when editing or viewing an expense.

**Backend tasks**: Already covered by E7-S1 (generic attachment endpoints with `entity_type=expense`).

**Frontend tasks**:
- Update `ExpenseForm.tsx` (edit mode) or create an expense detail view:
  - Below the form fields, show the `AttachmentList` component for `entity_type="expense"` and the expense's ID.
  - Include the `FileUpload` component to add new attachments.
- In the expense list table, show an attachment icon/count for expenses that have attachments.

**Acceptance criteria**:
- Given an expense, when the user uploads a receipt JPEG, then it appears in the expense's attachment list.
- Given an expense with attachments, when viewing the expense list, then an attachment indicator (e.g., paperclip icon with count) is shown.

---

### E3-S5: Attach Documents to Income

| Field | Value |
|---|---|
| **Ticket ID** | E3-S5 |
| **Title** | Attach pay slips or documents to income records |
| **Epic** | Epic 3 -- Income Management |
| **Milestone** | M2 |
| **Priority** | Should |
| **Size** | M |
| **Dependencies** | E3-S1, E7-S1 |

**Description**: Integrate the file attachment system with income records. Same pattern as expense attachments.

**Backend tasks**: Already covered by E7-S1 (generic attachment endpoints with `entity_type=income`).

**Frontend tasks**:
- Update `IncomeForm.tsx` (edit mode) or create an income detail view:
  - Include `AttachmentList` and `FileUpload` components for `entity_type="income"`.
- In the income list, show an attachment indicator for records with attachments.

**Acceptance criteria**:
- Given an income record, when the user uploads a pay slip PDF, then it appears in the record's attachment list.
- Given an income record with attachments, when viewing the income list, then an attachment indicator is shown.

---

### E4-S5: Attach Invoice PDFs

| Field | Value |
|---|---|
| **Ticket ID** | E4-S5 |
| **Title** | Attach actual invoice PDF to invoice records |
| **Epic** | Epic 4 -- Invoice & Bill Management |
| **Milestone** | M2 |
| **Priority** | Must |
| **Size** | M |
| **Dependencies** | E4-S1, E7-S1 |

**Description**: Integrate file attachment system with invoices. The invoice form should prominently feature file upload since attaching the original invoice document is a primary use case.

**Backend tasks**: Already covered by E7-S1 (generic attachment endpoints with `entity_type=invoice`).

**Frontend tasks**:
- Update `InvoiceForm.tsx`:
  - Include `FileUpload` as a prominent section (not just a footer).
  - In the invoice detail/edit view, show `AttachmentList` with inline PDF preview (FA-08 / IV-10).
- For PDF attachments, clicking the preview should open the PDF in the browser's built-in viewer (use `<iframe>` or `<embed>` tag pointing to the preview endpoint).

**Acceptance criteria**:
- Given an invoice, when the user uploads the original invoice PDF, then it appears with an inline PDF preview.
- Given the user clicks the PDF preview, then the full PDF opens in the browser's viewer.

---

### E4-S7: Overdue Check Admin Endpoint

| Field | Value |
|---|---|
| **Ticket ID** | E4-S7 |
| **Title** | Add manual admin endpoint for overdue status check |
| **Epic** | Epic 4 -- Invoice & Bill Management |
| **Milestone** | M2 |
| **Priority** | Should |
| **Size** | XS |
| **Dependencies** | E4-S4 |

**Description**: Add an admin endpoint to manually trigger the overdue status transition implemented in E4-S4. The core overdue logic (`UpdateOverdueStatuses`) is already implemented and called on startup and list fetch in E4-S4; this ticket only adds a manual trigger endpoint and structured logging.

**Backend tasks**:
- Add `POST /api/invoices/check-overdue` (authenticated) -- calls `InvoiceService.UpdateOverdueStatuses(ctx)` and returns `{"updated_count": N}`.
- Ensure `UpdateOverdueStatuses` logs at INFO level: "Updated N invoices to overdue status" (if N > 0).

**Frontend tasks**: None (admin/debug endpoint only).

**Acceptance criteria**:
- Given 2 unpaid invoices past due, when `POST /api/invoices/check-overdue` is called, then the response is `{"updated_count": 2}` and the server logs the update.
- Given no invoices past due, then the response is `{"updated_count": 0}`.

---

### E5-S2: Create Custom Categories

| Field | Value |
|---|---|
| **Ticket ID** | E5-S2 |
| **Title** | Create custom expense and income categories |
| **Epic** | Epic 5 -- Categories |
| **Milestone** | M2 |
| **Priority** | Should |
| **Size** | S |
| **Dependencies** | E5-S1 |

**Description**: Allow the user to create their own categories for expenses and income, in addition to the default ones.

**Backend tasks**:
- Add to `CategoryService`:
  - `Create(ctx, name string, categoryType string) (*Category, error)` -- validate name is not empty, type is "expense" or "income", name+type is unique. Set `is_default = false`.
- Add to category handler:
  - `POST /api/categories` -- accepts `{"name": "Pets", "type": "expense"}`. Returns 201 with created category. Returns 409 if duplicate name+type.

**Frontend tasks**:
- Implement `src/pages/CategoriesPage.tsx`:
  - Two sections: Expense Categories and Income Categories.
  - List all categories (default marked with a badge).
  - "Add Category" button opens a simple form: name (text) and type (select, or determined by which section the button is in).
- Create `src/components/categories/CategoryForm.tsx`: simple form with name input and type select.

**Acceptance criteria**:
- Given the user creates a category "Pets" of type "expense", then it appears in the expense category list and is available in expense forms.
- Given a category with the same name and type already exists, then the API returns 409 and the form shows "Category already exists".
- Given the user creates a custom category, then it has `is_default = false`.

---

### E5-S3: Rename and Delete Custom Categories

| Field | Value |
|---|---|
| **Ticket ID** | E5-S3 |
| **Title** | Rename or delete custom categories |
| **Epic** | Epic 5 -- Categories |
| **Milestone** | M2 |
| **Priority** | Should |
| **Size** | S |
| **Dependencies** | E5-S2 |

**Description**: Allow the user to rename or delete custom categories. Default categories cannot be deleted or renamed.

**Backend tasks**:
- Add to `CategoryService`:
  - `Update(ctx, id int64, name string) (*Category, error)` -- validate category is not default (is_default = false), name is not empty, new name+type is unique.
  - `Delete(ctx, id int64) error` -- validate category is not default. See E5-S4 for reassignment logic.
- Add to category handler:
  - `PUT /api/categories/:id` -- accepts `{"name": "New Name"}`. Returns 200. Returns 403 if default category. Returns 409 if duplicate.
  - `DELETE /api/categories/:id` -- returns 204. Returns 403 if default category.

**Frontend tasks**:
- In `CategoriesPage.tsx`:
  - Add Edit (rename) and Delete buttons to each custom category row.
  - Default categories show a lock icon and no Edit/Delete buttons.
  - Edit opens an inline form or modal to rename.
  - Delete shows a confirmation dialog warning that transactions will be moved to "Uncategorized".

**Acceptance criteria**:
- Given a custom category "Pets", when the user renames it to "Pet Care", then the new name is reflected in all forms and transaction lists.
- Given a default category "Food", when the user tries to delete it, then the action is blocked with "Default categories cannot be deleted".
- Given a custom category with associated transactions, when the user clicks Delete, then the confirmation warns about reassignment.

---

### E5-S4: Reassign Transactions on Category Delete

| Field | Value |
|---|---|
| **Ticket ID** | E5-S4 |
| **Title** | Move transactions to Uncategorized when category is deleted |
| **Epic** | Epic 5 -- Categories |
| **Milestone** | M2 |
| **Priority** | Should |
| **Size** | S |
| **Dependencies** | E5-S3 |

**Description**: When a custom category is deleted, all expenses/incomes in that category must be reassigned to the "Uncategorized" category of the same type (CT-05). No transaction data may be lost.

**Backend tasks**:
- Update `CategoryService.Delete()`:
  - Begin transaction.
  - Find the "Uncategorized" category ID for the same type: `SELECT id FROM categories WHERE name = 'Uncategorized' AND type = ? AND is_default = 1`.
  - Reassign: `UPDATE expenses SET category_id = ?, updated_at = CURRENT_TIMESTAMP WHERE category_id = ?` (and same for incomes).
  - Delete the category.
  - Commit transaction.
- Ensure "Uncategorized" is never deletable (it is a default category).

**Frontend tasks**: No additional frontend work beyond E5-S3 (the confirmation dialog should already mention reassignment).

**Acceptance criteria**:
- Given a custom category "Pets" with 5 expenses, when the user deletes it, then all 5 expenses are moved to the "Uncategorized" expense category.
- Given the deletion, when the user views those expenses, then their category is "Uncategorized".
- Given the database, then no transaction data is lost (same count of expenses before and after).

---

### M2-01: Expense List Sorting

| Field | Value |
|---|---|
| **Ticket ID** | M2-01 |
| **Title** | Sort expense list by date or amount |
| **Epic** | Epic 2 -- Expense Tracking |
| **Milestone** | M2 |
| **Priority** | Should |
| **Size** | S |
| **Dependencies** | E2-S2 |

**Description**: Allow sorting the expense list by date or amount in ascending or descending order (EX-08).

**Backend tasks**:
- Extend `GET /api/expenses` to accept `sort_by=date|amount` and `sort_order=asc|desc` query params. Default: `sort_by=date`, `sort_order=desc`.
- Whitelist allowed sort columns to prevent SQL injection.

**Frontend tasks**:
- Make table column headers (Date, Amount) clickable to toggle sort.
- Show a sort direction indicator (arrow) on the active sort column.
- Update URL query params with sort state.

**Acceptance criteria**:
- Given the expense list, when the user clicks the "Amount" header, then expenses are sorted by amount descending.
- Given the user clicks "Amount" again, then the sort order toggles to ascending.

---

### M2-02: Income Filter by Category

| Field | Value |
|---|---|
| **Ticket ID** | M2-02 |
| **Title** | Filter income list by source/category |
| **Epic** | Epic 3 -- Income Management |
| **Milestone** | M2 |
| **Priority** | Should |
| **Size** | S |
| **Dependencies** | E3-S2 |

**Description**: Add category filtering to the income list, mirroring the expense category filter (IN-06).

**Backend tasks**: Already partially covered in E3-S1 (`category_id` param).

**Frontend tasks**:
- Reuse `CategoryFilter.tsx` in `IncomePage.tsx` configured for income categories.

**Acceptance criteria**:
- Given income records from Salary and Freelance, when the user filters by Salary, then only Salary income is shown.

---

### M2-03: Invoice Date Range Filter

| Field | Value |
|---|---|
| **Ticket ID** | M2-03 |
| **Title** | Filter invoices by date range |
| **Epic** | Epic 4 -- Invoice & Bill Management |
| **Milestone** | M2 |
| **Priority** | Should |
| **Size** | S |
| **Dependencies** | E4-S2 |

**Description**: Add date range filtering to the invoice list by issue date or due date (IV-09).

**Backend tasks**:
- Extend `GET /api/invoices` to accept `date_from`, `date_to`, and `date_field=issue_date|due_date` query params.

**Frontend tasks**:
- Reuse `DateRangeFilter.tsx` in `InvoicesPage.tsx`.
- Add a toggle or dropdown to switch between filtering by issue date vs. due date.

**Acceptance criteria**:
- Given invoices with various issue dates, when the user filters by issue date range, then only matching invoices are shown.
- Given the user switches to filter by due date, then the filter applies to the due_date column.

---

### M2-04: Income Running Total

| Field | Value |
|---|---|
| **Ticket ID** | M2-04 |
| **Title** | Display total income for filtered period |
| **Epic** | Epic 3 -- Income Management |
| **Milestone** | M2 |
| **Priority** | Should |
| **Size** | XS |
| **Dependencies** | E3-S3 |

**Description**: Show the total income amount for the current filter, same as E2-S8 for expenses (IN-08).

**Backend tasks**: Already included in E3-S1 (`total_amount` in list response).

**Frontend tasks**:
- Display total amount in the income list summary bar, same pattern as expense list.

**Acceptance criteria**:
- Given 3 filtered income records totaling 15,000,000, then "Total: 15,000,000 VND" is displayed.

---

## Milestone 3 -- Reports & Export

**Goal**: Add charts, period switching, and CSV export to the dashboard and transaction lists.

**Exit criteria**: Dashboard shows income vs. expense chart and category breakdown chart; user can switch periods; CSV export works on filtered views.

---

### E6-S2: Dashboard Period Switcher

| Field | Value |
|---|---|
| **Ticket ID** | E6-S2 |
| **Title** | Switch dashboard summary period |
| **Epic** | Epic 6 -- Dashboard & Reports |
| **Milestone** | M3 |
| **Priority** | Should |
| **Size** | S |
| **Dependencies** | E6-S1 |

**Description**: Allow the user to change the dashboard period to current month, last month, current year, or a custom date range (DB-03). All summary cards and charts recalculate for the selected period.

**Frontend tasks**:
- Create `src/components/dashboard/PeriodSelector.tsx`:
  - Preset buttons: "This Month", "Last Month", "This Year".
  - Custom date range picker (two date inputs with Apply button).
  - On selection, compute `date_from` and `date_to` and pass to all dashboard API calls.
- Update `DashboardPage.tsx` to use the period selector and pass dates to all data-fetching hooks.
- Persist the selected period in URL query params.

**Backend tasks**: Already supported by `date_from`/`date_to` params on `GET /api/dashboard/summary`.

**Acceptance criteria**:
- Given the user selects "Last Month", then all dashboard metrics recalculate for the previous calendar month.
- Given the user selects "This Year", then metrics cover Jan 1 to today.
- Given a custom date range, when the user applies it, then all dashboard metrics recalculate within 2 seconds (NF-02).

---

### E6-S3: Income vs. Expenses Chart

| Field | Value |
|---|---|
| **Ticket ID** | E6-S3 |
| **Title** | Bar/line chart showing income vs. expenses by month |
| **Epic** | Epic 6 -- Dashboard & Reports |
| **Milestone** | M3 |
| **Priority** | Should |
| **Size** | M |
| **Dependencies** | E6-S2 |

**Description**: Add a chart to the dashboard showing monthly income and expense totals side by side over the selected period (DB-04). Use a bar chart or line chart.

**Backend tasks**:
- Create `internal/services/report_service.go` (or extend `DashboardService`):
  - `GetMonthlyTrend(ctx, dateFrom, dateTo string) ([]MonthlyTrendItem, error)`:
    - Query: `SELECT strftime('%Y-%m', date) AS month, COALESCE(SUM(amount), 0) AS total FROM expenses WHERE date BETWEEN ? AND ? GROUP BY month ORDER BY month`.
    - Same for incomes.
    - Return `[]MonthlyTrendItem { Month string, TotalIncome int64, TotalExpenses int64 }`.
- Add to dashboard handler:
  - `GET /api/dashboard/monthly-trend?date_from=2026-01-01&date_to=2026-04-30` -- returns `{"data": [{"month": "2026-01", "total_income": ..., "total_expenses": ...}, ...]}`.

**Frontend tasks**:
- Install a chart library: `npm install recharts` (or `chart.js` + `react-chartjs-2`; Recharts recommended for React integration -- D5).
- Create `src/components/dashboard/MonthlyTrendChart.tsx`:
  - Bar chart with two series: Income (green bars) and Expenses (red bars), grouped by month.
  - X-axis: months (e.g., "Jan", "Feb", "Mar", "Apr").
  - Y-axis: amounts in major currency units.
  - Tooltip showing exact amounts on hover.
  - Responsive width.
- Add to `DashboardPage.tsx` below the summary cards.

**Acceptance criteria**:
- Given data exists for January through April 2026, when "This Year" is selected, then a bar chart shows 4 month groups with income and expense bars.
- Given a month with no data, then it shows as 0 in the chart.
- Given the user hovers over a bar, then a tooltip shows the exact amount.

---

### E6-S4: Expense Breakdown by Category Chart

| Field | Value |
|---|---|
| **Ticket ID** | E6-S4 |
| **Title** | Pie/donut chart showing expense breakdown by category |
| **Epic** | Epic 6 -- Dashboard & Reports |
| **Milestone** | M3 |
| **Priority** | Should |
| **Size** | M |
| **Dependencies** | E6-S2 |

**Description**: Add a pie or donut chart to the dashboard showing the proportion of expenses by category for the selected period (DB-05).

**Backend tasks**:
- Add to report/dashboard service:
  - `GetExpenseByCategory(ctx, dateFrom, dateTo string) ([]CategoryBreakdownItem, error)`:
    - Query: `SELECT c.name, COALESCE(SUM(e.amount), 0) AS total FROM expenses e JOIN categories c ON e.category_id = c.id WHERE e.date BETWEEN ? AND ? GROUP BY c.id ORDER BY total DESC`.
    - Return `[]CategoryBreakdownItem { CategoryName string, Total int64 }`.
- Add to dashboard handler:
  - `GET /api/dashboard/expense-by-category?date_from=...&date_to=...` -- returns `{"data": [{"category_name": "Food", "total": 500000}, ...]}`.

**Frontend tasks**:
- Create `src/components/dashboard/CategoryBreakdownChart.tsx`:
  - Donut chart with segments per category.
  - Legend showing category names, amounts, and percentages.
  - Colors per category (use category color if available, otherwise assign from a palette).
  - Tooltip showing category name, amount, and percentage on hover.
  - Responsive sizing.
- Add to `DashboardPage.tsx` alongside the monthly trend chart (two-column layout on desktop).

**Acceptance criteria**:
- Given expenses in Food (300,000), Transport (200,000), and Housing (500,000) for the selected period, then the donut chart shows three segments with correct proportions (30%, 20%, 50%).
- Given no expenses for the period, then the chart shows an empty state message.

---

### E6-S5: CSV Export

| Field | Value |
|---|---|
| **Ticket ID** | E6-S5 |
| **Title** | Export filtered transaction list as CSV |
| **Epic** | Epic 6 -- Dashboard & Reports |
| **Milestone** | M3 |
| **Priority** | Should |
| **Size** | L |
| **Dependencies** | E2-S3, E3-S3 |

**Description**: Allow the user to download the currently filtered transaction list (expenses and/or income) as a UTF-8 CSV file (DB-06). The export should respect all active filters.

**Backend tasks**:
- Create `internal/handlers/export_handler.go`:
  - `GET /api/export/transactions?type=expense&date_from=...&date_to=...&category_id=...` -- accepts the same filters as the list endpoints.
  - `type` param: `expense`, `income`, or `all` (both merged).
  - Stream CSV response with `Content-Type: text/csv; charset=utf-8` and `Content-Disposition: attachment; filename="transactions_2026-04-26.csv"`.
  - CSV columns: `date, type, category, description, amount` (DB-06 acceptance criteria).
  - Amount in the CSV should be in major currency units (divide by 100).
  - Use Go's `encoding/csv` package. Stream rows to avoid loading all data in memory.

**Frontend tasks**:
- Add an "Export CSV" button to both `ExpensesPage.tsx` and `IncomePage.tsx`.
- Create `src/components/ExportButton.tsx` -- on click, constructs the export URL with current filter params and triggers a download (open in new tab or programmatic download).
- Optionally add an "Export All Transactions" button on the dashboard.

**Acceptance criteria**:
- Given 50 filtered expenses, when the user clicks "Export CSV", then the browser downloads a CSV file with 50 data rows plus a header row.
- Given the CSV file, when opened in a spreadsheet, then columns are: date, type, category, description, amount -- all correctly populated.
- Given the export, then amounts are in major currency units (not minor units).
- Given UTF-8 characters in descriptions, then the CSV is correctly encoded.

---

## Milestone 4 -- Polish & Nice-to-Haves

**Goal**: Implement could-have features, polish UX, and add remaining should-have items.

**Exit criteria**: All should-have items complete; selected could-have items delivered based on available time.

---

### E1-S4: Change Password

| Field | Value |
|---|---|
| **Ticket ID** | E1-S4 |
| **Title** | Change password from settings page |
| **Epic** | Epic 1 -- Authentication |
| **Milestone** | M4 |
| **Priority** | Should |
| **Size** | S |
| **Dependencies** | E1-S1 |

**Description**: Allow the user to change their password from a settings page. Requires the current password for verification.

**Backend tasks**:
- Add to `AuthService`:
  - `ChangePassword(ctx, userID int64, currentPassword, newPassword string) error` -- verify current password matches hash, validate new password (minimum length), hash new password with bcrypt (cost >= 12 -- NF-08), update in database.
- Add to auth handler:
  - `PUT /api/auth/password` -- accepts `{"current_password": "...", "new_password": "..."}`. Returns 200 on success. Returns 400 if current password is wrong. Returns 400 if new password is too short.

**Frontend tasks**:
- Implement `src/pages/SettingsPage.tsx`:
  - Password change form: current password, new password, confirm new password.
  - Client-side validation: all fields required, new password minimum 8 characters, new password and confirm must match.
  - On success, show a success message. Optionally invalidate the current session and force re-login.

**Acceptance criteria**:
- Given valid current password and a new password >= 8 characters, when the user submits, then the password is updated and the user can log in with the new password.
- Given an incorrect current password, then the form shows "Current password is incorrect".
- Given a new password shorter than 8 characters, then the form shows "Password must be at least 8 characters".

---

### M4-01: Category Colors and Icons

| Field | Value |
|---|---|
| **Ticket ID** | M4-01 |
| **Title** | Assign colors and icons to categories |
| **Epic** | Epic 5 -- Categories |
| **Milestone** | M4 |
| **Priority** | Could |
| **Size** | S |
| **Dependencies** | E5-S2 |

**Description**: Allow the user to assign a color (hex code) to a category (CT-06). Colors are used in charts and category badges throughout the UI.

**Backend tasks**:
- The `categories` table already has a `color TEXT` column.
- Update `CategoryService.Create` and `CategoryService.Update` to accept an optional `color` parameter (hex string, validated format `#RRGGBB`).
- Update category list response to include `color`.

**Frontend tasks**:
- Add a color picker to `CategoryForm.tsx`.
- Use category colors in:
  - Category badges in expense/income tables.
  - Donut chart segments (E6-S4).
  - Category filter dropdown items.
- Default colors from a predefined palette for categories without a custom color.

**Acceptance criteria**:
- Given the user sets the "Food" category color to `#FF6B6B`, then all Food badges and chart segments use that color.
- Given a category without a custom color, then a default palette color is used.

---

### M4-02: Remember Me Option

| Field | Value |
|---|---|
| **Ticket ID** | M4-02 |
| **Title** | "Remember me" option extending token lifetime |
| **Epic** | Epic 1 -- Authentication |
| **Milestone** | M4 |
| **Priority** | Could |
| **Size** | XS |
| **Dependencies** | E1-S1 |

**Description**: Add a "Remember me" checkbox to the login form that extends the token lifetime (e.g., 30 days instead of 24 hours) (AU-07).

**Backend tasks**:
- Update `POST /api/auth/login` to accept an optional `remember_me: boolean` field.
- If true, set token expiry to 30 days. If false (default), use the standard `TOKEN_EXPIRY_HOURS`.

**Frontend tasks**:
- Add a "Remember me" checkbox to `LoginPage.tsx`.
- Pass `remember_me` in the login API call.

**Acceptance criteria**:
- Given the user checks "Remember me" and logs in, then the token is valid for 30 days.
- Given the user does not check it, then the token expires after the default period.

---

### M4-03: Due Date Browser Notification

| Field | Value |
|---|---|
| **Ticket ID** | M4-03 |
| **Title** | Browser notification for invoices approaching due date |
| **Epic** | Epic 4 -- Invoice & Bill Management |
| **Milestone** | M4 |
| **Priority** | Could |
| **Size** | M |
| **Dependencies** | E4-S7 |

**Description**: Show a browser notification when an invoice is within 3 days of its due date (IV-12). Requires the user to grant notification permission.

**Backend tasks**:
- Add to dashboard or invoice summary endpoint:
  - Include `approaching_due` field: list of invoices where `due_date BETWEEN date('now') AND date('now', '+3 days')` and `status = 'unpaid'`.

**Frontend tasks**:
- On dashboard load, check for approaching-due invoices.
- Request browser notification permission (using the Notification API).
- If permission is granted and there are approaching-due invoices, show a notification: "You have N invoices due in the next 3 days".
- Also show an in-app banner on the dashboard.

**Acceptance criteria**:
- Given an unpaid invoice with due_date 2 days from now, when the dashboard loads, then a browser notification is shown (if permission was granted).
- Given no invoices approaching due date, then no notification is shown.

---

### M4-04: Expense Tags

| Field | Value |
|---|---|
| **Ticket ID** | M4-04 |
| **Title** | Add free-text tags to expenses |
| **Epic** | Epic 2 -- Expense Tracking |
| **Milestone** | M4 |
| **Priority** | Could |
| **Size** | M |
| **Dependencies** | E2-S1 |

**Description**: Allow the user to add free-text tags to expenses in addition to the category (EX-10). Tags enable more flexible filtering and grouping.

**Backend tasks**:
- Create migration for `expense_tags` table:
  ```sql
  CREATE TABLE expense_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expense_id INTEGER NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  CREATE INDEX idx_expense_tags_expense_id ON expense_tags(expense_id);
  CREATE INDEX idx_expense_tags_tag ON expense_tags(tag);
  ```
- Update expense create/update endpoints to accept an optional `tags: string[]` field.
- Update expense list response to include `tags`.
- Add `tag` as a filter param on `GET /api/expenses?tag=vacation`.

**Frontend tasks**:
- Add a tag input to `ExpenseForm.tsx` (comma-separated or chip-style input).
- Display tags as badges in the expense table.
- Add tag filter to the expense list filters.

**Acceptance criteria**:
- Given the user creates an expense with tags ["vacation", "food"], then both tags are stored and visible on the expense.
- Given the user filters by tag "vacation", then only expenses tagged "vacation" are shown.

---

### M4-05: High Expense Warning

| Field | Value |
|---|---|
| **Ticket ID** | M4-05 |
| **Title** | Warn on unusually high expense amounts |
| **Epic** | Epic 2 -- Expense Tracking |
| **Milestone** | M4 |
| **Priority** | Could |
| **Size** | S |
| **Dependencies** | E2-S1 |

**Description**: Show a confirmation warning when the user enters an expense amount above a configurable threshold (EX-11). This helps catch data entry errors.

**Backend tasks**:
- Add `HIGH_EXPENSE_THRESHOLD` to config (default: 10,000,000 in minor units, i.e., 100,000 VND).
- The threshold is informational only -- the API still accepts the expense but includes a `warning` field in the response if the amount exceeds the threshold.

**Frontend tasks**:
- In `ExpenseForm.tsx`, check the amount against the threshold before submitting.
- If exceeded, show a warning dialog: "This amount (1,000,000 VND) is higher than your usual threshold. Are you sure?"
- On confirm, submit the expense. On cancel, return to the form.

**Acceptance criteria**:
- Given a threshold of 100,000 VND, when the user enters 150,000 VND, then a warning dialog is shown.
- Given the user confirms, then the expense is created normally.
- Given an amount below the threshold, then no warning is shown.

---

### M4-06: Storage Usage Display

| Field | Value |
|---|---|
| **Ticket ID** | M4-06 |
| **Title** | Display total storage usage on settings page |
| **Epic** | Epic 7 -- File Attachments |
| **Milestone** | M4 |
| **Priority** | Could |
| **Size** | S |
| **Dependencies** | E7-S1 |

**Description**: Show the total storage space used by uploaded files on the settings page (FA-09).

**Backend tasks**:
- Add to settings/admin handler:
  - `GET /api/settings/storage` -- returns `{"total_size_bytes": 52428800, "file_count": 15}` computed from `SELECT COUNT(*), COALESCE(SUM(size_bytes), 0) FROM attachments`.

**Frontend tasks**:
- In `SettingsPage.tsx`, add a "Storage" section showing: "15 files, 50.0 MB used".
- Format bytes into human-readable units (KB, MB, GB).

**Acceptance criteria**:
- Given 15 files totaling 50 MB, when the settings page loads, then it displays "15 files, 50.0 MB used".
- Given no files uploaded, then it displays "No files uploaded".

---

### M4-07: PDF Report Export

| Field | Value |
|---|---|
| **Ticket ID** | M4-07 |
| **Title** | Export summary report as PDF |
| **Epic** | Epic 6 -- Dashboard & Reports |
| **Milestone** | M4 |
| **Priority** | Could |
| **Size** | L |
| **Dependencies** | E6-S3, E6-S4 |

**Description**: Allow the user to export the dashboard summary (including charts) as a PDF report (DB-08).

**Backend tasks**:
- Evaluate Go PDF libraries (e.g., `jung-kurt/gofpdf` or `unidoc/unipdf`).
- Create `GET /api/export/report?date_from=...&date_to=...` that generates a PDF containing:
  - Title and date range.
  - Summary table: total income, total expenses, net balance.
  - Monthly trend data table.
  - Category breakdown table.
- Return with `Content-Type: application/pdf`.

**Frontend tasks**:
- Alternatively, use a client-side approach: use a library like `html2canvas` + `jspdf` to capture the dashboard as a PDF.
- Add an "Export PDF" button to the dashboard.

**Acceptance criteria**:
- Given the dashboard has data, when the user clicks "Export PDF", then a PDF file is downloaded containing the summary data.
- Given the selected period is "This Year", then the PDF reflects the full year's data.

---

### M4-08: Month-over-Month Trend Indicator

| Field | Value |
|---|---|
| **Ticket ID** | M4-08 |
| **Title** | Show month-over-month trend arrows on dashboard |
| **Epic** | Epic 6 -- Dashboard & Reports |
| **Milestone** | M4 |
| **Priority** | Could |
| **Size** | S |
| **Dependencies** | E6-S1 |

**Description**: Add up/down trend indicators to the dashboard summary cards showing whether income/expenses increased or decreased compared to the previous period (DB-07).

**Backend tasks**:
- Extend `GET /api/dashboard/summary` to include comparison data:
  - For "This Month" period, also compute totals for last month.
  - Return `previous_total_income`, `previous_total_expenses`, and computed `income_change_pct`, `expense_change_pct`.

**Frontend tasks**:
- Update `SummaryCard.tsx` to show a trend indicator: green up-arrow for income increase (good), red up-arrow for expense increase (bad), and vice versa.
- Display percentage change: "+12.5%" or "-8.3%".

**Acceptance criteria**:
- Given this month's expenses are 120,000 and last month's were 100,000, then the expense card shows a red up-arrow with "+20.0%".
- Given this month's income is higher than last month's, then the income card shows a green up-arrow.

---

### M4-09: Database Backup Endpoint

| Field | Value |
|---|---|
| **Ticket ID** | M4-09 |
| **Title** | Authenticated database backup download endpoint |
| **Epic** | Infrastructure |
| **Milestone** | M1 |
| **Priority** | Must |
| **Size** | S |
| **Dependencies** | E1-S1 |

**Description**: Expose an authenticated endpoint to download the SQLite database file for backup purposes (NF-13).

**Backend tasks**:
- Create `GET /api/backup` (authenticated):
  - Use SQLite's backup API or simply stream the database file.
  - For safety, use SQLite's `.backup` command or read a snapshot while in WAL mode.
  - Set `Content-Disposition: attachment; filename="moneyapp_backup_2026-04-26.db"`.
  - Set `Content-Type: application/octet-stream`.

**Frontend tasks**:
- Add a "Download Backup" button to `SettingsPage.tsx`.
- On click, trigger download from `GET /api/backup`.

**Acceptance criteria**:
- Given an authenticated user, when they click "Download Backup", then the SQLite database file is downloaded.
- Given the download, when the file is opened with an SQLite client, then all data is intact.
- Given an unauthenticated request, then the endpoint returns 401.

---

### M4-10: Recurring Invoice Marking

| Field | Value |
|---|---|
| **Ticket ID** | M4-10 |
| **Title** | Mark an invoice as recurring |
| **Epic** | Epic 4 -- Invoice & Bill Management |
| **Milestone** | M4 |
| **Priority** | Could |
| **Size** | S |
| **Dependencies** | E4-S1 |

**Description**: Allow the user to mark an invoice as recurring (monthly, quarterly, or yearly) for informational purposes (IV-13). Automated generation is out of scope (IV-14), but the label helps the user remember which bills recur.

**Backend tasks**:
- Add migration to add `recurrence` column to `invoices` table: `ALTER TABLE invoices ADD COLUMN recurrence TEXT CHECK(recurrence IN (NULL, 'monthly', 'quarterly', 'yearly'))`.
- Update invoice create/update endpoints to accept optional `recurrence` field.
- Update invoice list response to include `recurrence`.

**Frontend tasks**:
- Add a "Recurrence" dropdown (None/Monthly/Quarterly/Yearly) to `InvoiceForm.tsx`.
- Show a recurrence badge on invoice list rows.
- Optionally filter invoices by recurrence.

**Acceptance criteria**:
- Given the user creates an invoice and sets recurrence to "monthly", then the invoice shows a "Monthly" badge in the list.
- Given recurrence is not set, then no badge is shown.

---

## Appendix: Ticket Summary Table

| Ticket ID | Title | Milestone | Priority | Size | Dependencies |
|-----------|-------|-----------|----------|------|--------------|
| M0-01 | Database migration runner | M0 | Must | S | -- |
| M0-02 | Environment config and .env.example | M0 | Must | XS | -- |
| M0-03 | API error handling and response patterns | M0 | Must | S | -- |
| M0-04 | Frontend routing and app shell | M0 | Must | S | -- |
| M0-05 | Frontend API client | M0 | Must | S | M0-03 |
| M0-06 | Health check enhancement | M0 | Must | XS | M0-02, M0-03 |
| M0-07 | Docker Compose enhancement | M0 | Should | XS | M0-02 |
| M0-08 | CI pipeline setup | M0 | Should | S | M0-01, M0-04 |
| E1-S1 | User login with JWT | M1 | Must | S | M0-01, M0-02, M0-03, M0-05 |
| E1-S2 | Token expiry | M1 | Should | S | E1-S1 |
| E1-S3 | User logout | M1 | Must | XS | E1-S1 |
| E5-S1 | Default categories | M1 | Must | XS | M0-01 |
| E2-S1 | Create expense | M1 | Must | S | E1-S1, E5-S1 |
| E2-S2 | List expenses (paginated) | M1 | Must | S | E2-S1 |
| E2-S3 | Filter expenses by date range | M1 | Must | S | E2-S2 |
| E2-S4 | Filter expenses by category | M1 | Must | S | E2-S2 |
| E2-S5 | Edit expense | M1 | Must | XS | E2-S1 |
| E2-S6 | Delete expense | M1 | Must | XS | E2-S1 |
| E2-S8 | Running total for filtered expenses | M1 | Must | XS | E2-S3 |
| E3-S1 | Create income | M1 | Must | S | E1-S1, E5-S1 |
| E3-S2 | List income (paginated) | M1 | Must | S | E3-S1 |
| E3-S3 | Filter income by date range | M1 | Must | S | E3-S2 |
| E3-S4 | Edit and delete income | M1 | Must | XS | E3-S1 |
| E4-S1 | Create invoice/bill | M1 | Must | M | E1-S1, M0-01 |
| E4-S2 | List invoices with status filter | M1 | Must | S | E4-S1 |
| E4-S3 | Mark invoice as paid | M1 | Must | XS | E4-S1 |
| E4-S4 | Overdue invoice highlighting | M1 | Must | S | E4-S1 |
| E4-S6 | Edit and delete invoice | M1 | Must | XS | E4-S1 |
| E4-S8 | Outstanding invoice total | M1 | Must | XS | E4-S2 |
| E6-S1 | Dashboard summary cards | M1 | Must | M | E2-S1, E3-S1 |
| M1-01 | Service layer unit test coverage | M1 | Must | M | E2-S1, E3-S1, E4-S1, E5-S1 |
| E6-S6 | Invoice summary on dashboard | M1 | Must | S | E6-S1, E4-S1 |
| E7-S1 | File upload infrastructure | M2 | Must | L | M0-01, M0-02 |
| E7-S2 | File size validation | M2 | Must | XS | E7-S1 |
| E7-S3 | File download | M2 | Must | S | E7-S1 |
| E7-S4 | Image thumbnail preview | M2 | Should | S | E7-S1, E7-S3 |
| E7-S5 | Cascade delete attachments | M2 | Must | S | E7-S1, E2-S6, E3-S4, E4-S6 |
| E2-S7 | Attach receipts to expenses | M2 | Should | M | E2-S1, E7-S1 |
| E3-S5 | Attach documents to income | M2 | Should | M | E3-S1, E7-S1 |
| E4-S5 | Attach invoice PDFs | M2 | Must | M | E4-S1, E7-S1 |
| E4-S7 | Overdue check admin endpoint | M2 | Should | XS | E4-S4 |
| E5-S2 | Create custom categories | M2 | Should | S | E5-S1 |
| E5-S3 | Rename and delete custom categories | M2 | Should | S | E5-S2 |
| E5-S4 | Reassign transactions on category delete | M2 | Should | S | E5-S3 |
| M2-01 | Expense list sorting | M2 | Should | S | E2-S2 |
| M2-02 | Income filter by category | M2 | Should | S | E3-S2 |
| M2-03 | Invoice date range filter | M2 | Should | S | E4-S2 |
| M2-04 | Income running total | M2 | Should | XS | E3-S3 |
| E6-S2 | Dashboard period switcher | M3 | Should | S | E6-S1 |
| E6-S3 | Income vs. expenses chart | M3 | Should | M | E6-S2 |
| E6-S4 | Expense breakdown by category chart | M3 | Should | M | E6-S2 |
| E6-S5 | CSV export | M3 | Should | L | E2-S3, E3-S3 |
| E1-S4 | Change password | M4 | Should | S | E1-S1 |
| M4-01 | Category colors and icons | M4 | Could | S | E5-S2 |
| M4-02 | Remember me option | M4 | Could | XS | E1-S1 |
| M4-03 | Due date browser notification | M4 | Could | M | E4-S7 |
| M4-04 | Expense tags | M4 | Could | M | E2-S1 |
| M4-05 | High expense warning | M4 | Could | S | E2-S1 |
| M4-06 | Storage usage display | M4 | Could | S | E7-S1 |
| M4-07 | PDF report export | M4 | Could | L | E6-S3, E6-S4 |
| M4-08 | Month-over-month trend indicator | M4 | Could | S | E6-S1 |
| M4-09 | Database backup endpoint | M1 | Must | S | E1-S1 |
| M4-10 | Recurring invoice marking | M4 | Could | S | E4-S1 |

---

## Appendix: API Route Summary

| Method | Route | Ticket | Auth |
|--------|-------|--------|------|
| GET | `/api/health` | M0-06 | No |
| POST | `/api/auth/login` | E1-S1 | No |
| POST | `/api/auth/logout` | E1-S3 | Yes |
| PUT | `/api/auth/password` | E1-S4 | Yes |
| GET | `/api/categories?type=` | E5-S1 | Yes |
| POST | `/api/categories` | E5-S2 | Yes |
| PUT | `/api/categories/:id` | E5-S3 | Yes |
| DELETE | `/api/categories/:id` | E5-S3 | Yes |
| POST | `/api/expenses` | E2-S1 | Yes |
| GET | `/api/expenses?page=&per_page=&date_from=&date_to=&category_id=&sort_by=&sort_order=` | E2-S2, E2-S3, E2-S4, M2-01 | Yes |
| GET | `/api/expenses/:id` | E2-S5 | Yes |
| PUT | `/api/expenses/:id` | E2-S5 | Yes |
| DELETE | `/api/expenses/:id` | E2-S6 | Yes |
| POST | `/api/incomes` | E3-S1 | Yes |
| GET | `/api/incomes?page=&per_page=&date_from=&date_to=&category_id=` | E3-S1, E3-S3, M2-02 | Yes |
| GET | `/api/incomes/:id` | E3-S4 | Yes |
| PUT | `/api/incomes/:id` | E3-S4 | Yes |
| DELETE | `/api/incomes/:id` | E3-S4 | Yes |
| POST | `/api/invoices` | E4-S1 | Yes |
| GET | `/api/invoices?page=&per_page=&status=&date_from=&date_to=&date_field=` | E4-S2, M2-03 | Yes |
| GET | `/api/invoices/:id` | E4-S6 | Yes |
| PUT | `/api/invoices/:id` | E4-S6 | Yes |
| DELETE | `/api/invoices/:id` | E4-S6 | Yes |
| PATCH | `/api/invoices/:id/status` | E4-S3 | Yes |
| GET | `/api/invoices/stats` | E4-S8 | Yes |
| POST | `/api/attachments` | E7-S1 | Yes |
| GET | `/api/attachments?entity_type=&entity_id=` | E7-S1 | Yes |
| DELETE | `/api/attachments/:id` | E7-S1 | Yes |
| GET | `/api/attachments/:id/download` | E7-S3 | Yes |
| GET | `/api/attachments/:id/preview` | E7-S4 | Yes |
| GET | `/api/dashboard/summary?date_from=&date_to=` | E6-S1, E6-S6 | Yes |
| GET | `/api/dashboard/monthly-trend?date_from=&date_to=` | E6-S3 | Yes |
| GET | `/api/dashboard/expense-by-category?date_from=&date_to=` | E6-S4 | Yes |
| GET | `/api/export/transactions?type=&date_from=&date_to=&category_id=` | E6-S5 | Yes |
| GET | `/api/export/report?date_from=&date_to=` | M4-07 | Yes |
| GET | `/api/settings/storage` | M4-06 | Yes |
| GET | `/api/backup` | M4-09 | Yes |

---

## Appendix: Database Migration Order

| Migration File | Ticket | Description |
|---|---|---|
| `001_create_users.up.sql` | M0-01 | Users table |
| `002_seed_default_user.up.sql` | E1-S1 | Seed admin user |
| `003_create_categories.up.sql` | E5-S1 | Categories table |
| `004_seed_default_categories.up.sql` | E5-S1 | Default expense and income categories |
| `005_create_expenses.up.sql` | E2-S1 | Expenses table with indexes |
| `006_create_incomes.up.sql` | E3-S1 | Incomes table with indexes |
| `007_create_invoices.up.sql` | E4-S1 | Invoices table with indexes |
| `008_create_attachments.up.sql` | E7-S1 | Attachments table (polymorphic) |
| `009_create_expense_tags.up.sql` | M4-04 | Expense tags table |
| `010_add_invoice_recurrence.up.sql` | M4-10 | Add recurrence column to invoices |
