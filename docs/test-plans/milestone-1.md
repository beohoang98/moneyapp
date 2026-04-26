# Test Plan — Milestone 1: MVP Core

**Milestone goal**: Fully functional core with auth, CRUD for all entities, basic filtering, and dashboard summary cards. No file attachments or charts yet.  
**Exit criteria (from tickets.md)**: User can log in, record expenses/income/invoices, filter lists, and see a dashboard summary. All data persists across restarts.  
**NF references covered**: AU-01, AU-02, AU-03, AU-04, AU-05, NF-01, NF-02, NF-04, NF-06, NF-08, NF-09, NF-10, NF-14, NF-15, NF-16, NF-17, NF-20  
**Date**: 2026-04-26  
**Status**: Draft

---

## Table of Contents

1. [Environment Setup](#environment-setup)
2. [Traceability Matrix](#traceability-matrix)
3. [E1-S1 — User Login](#e1-s1--user-login)
4. [E1-S2 — Token Expiry](#e1-s2--token-expiry)
5. [E1-S3 — User Logout](#e1-s3--user-logout)
6. [E5-S1 — Default Categories](#e5-s1--default-categories)
7. [E2-S1 — Create Expense](#e2-s1--create-expense)
8. [E2-S2 — List Expenses (Paginated)](#e2-s2--list-expenses-paginated)
9. [E2-S3 — Filter Expenses by Date Range](#e2-s3--filter-expenses-by-date-range)
10. [E2-S4 — Filter Expenses by Category](#e2-s4--filter-expenses-by-category)
11. [E2-S5 — Edit Expense](#e2-s5--edit-expense)
12. [E2-S6 — Delete Expense](#e2-s6--delete-expense)
13. [E2-S8 — Running Total for Filtered Expenses](#e2-s8--running-total-for-filtered-expenses)
14. [E3-S1 — Create Income](#e3-s1--create-income)
15. [E3-S2 — List Income (Paginated)](#e3-s2--list-income-paginated)
16. [E3-S3 — Filter Income by Date Range](#e3-s3--filter-income-by-date-range)
17. [E3-S4 — Edit and Delete Income](#e3-s4--edit-and-delete-income)
18. [E4-S1 — Create Invoice/Bill](#e4-s1--create-invoicebill)
19. [E4-S2 — List Invoices with Status Filter](#e4-s2--list-invoices-with-status-filter)
20. [E4-S3 — Mark Invoice as Paid](#e4-s3--mark-invoice-as-paid)
21. [E4-S4 — Overdue Invoice Highlighting](#e4-s4--overdue-invoice-highlighting)
22. [E4-S6 — Edit and Delete Invoice](#e4-s6--edit-and-delete-invoice)
23. [E4-S8 — Outstanding Invoice Total](#e4-s8--outstanding-invoice-total)
24. [E6-S1 — Dashboard Summary Cards](#e6-s1--dashboard-summary-cards)
25. [E6-S6 — Invoice Summary on Dashboard](#e6-s6--invoice-summary-on-dashboard)
26. [M1-01 — Service Layer Unit Test Coverage](#m1-01--service-layer-unit-test-coverage)
27. [Cross-Cutting Checks](#cross-cutting-checks)
28. [Automation Backlog](#automation-backlog)
29. [Coverage Gaps & Notes](#coverage-gaps--notes)

---

## Environment Setup

### Required environments

| Environment | Description | Notes |
|---|---|---|
| **m1-fresh** | Fresh checkout, no `.env`, no DB file | Verifies seed user creation and default categories on first run |
| **m1-local** | `.env` with `STORAGE_TYPE=local`; no MinIO container | Standard dev environment for all M1 tests (no file attachments in scope yet) |
| **m1-expired-token** | `.env` with `TOKEN_EXPIRY_HOURS=0` or manually crafted expired JWT | Tests token expiry handling |
| **m1-seeded** | M1-local with `admin`/`changeme` seed user present and default categories migrated | Most functional test cases run in this environment |

### Prerequisites

#### Option A — Minimal (no MinIO, `STORAGE_TYPE=local`)

```bash
# Copy env; local storage is the default
cp .env.example .env
# Confirm / set:
#   STORAGE_TYPE=local
#   LOCAL_STORAGE_PATH=./data/storage
#   JWT_SECRET=test-secret-m1
#   TOKEN_EXPIRY_HOURS=24
#   CURRENCY=VND

# Start backend (migrations run automatically, seed user created)
cd backend && go run ./cmd/server

# Start frontend
cd frontend && npm run dev
```

#### Option B — With S3/MinIO (optional for M1, no file attachments in scope)

S3/MinIO is not required for any M1 ticket. Use `STORAGE_TYPE=local` for all M1 test cases.

### Seeded test credentials

| Username | Password | Notes |
|---|---|---|
| `admin` | `changeme` | Created by migration `002_seed_default_user.up.sql` via bcrypt hash (cost 12). Present on any fresh DB after server start. |

> **Password storage**: The seed user's password hash is stored using bcrypt with cost factor 12 (NF-08). The migration contains a pre-computed hash literal — do not change `BCRYPT_COST` between runs if the seed hash was computed at a different cost.

### Environment variables under test (M1-specific)

```
# Auth
JWT_SECRET=test-secret-m1       # Must not be empty (NF-06)
TOKEN_EXPIRY_HOURS=24            # E1-S2; set to 1 to speed up expiry tests
BCRYPT_COST=12                   # NF-08

# Currency display
CURRENCY=VND                     # NF-17

# Storage (local; no MinIO needed for M1)
STORAGE_TYPE=local
LOCAL_STORAGE_PATH=./data/storage

# Core
PORT=8080
DB_PATH=./moneyapp.db
```

---

## Traceability Matrix

| Ticket | Title | Test Suite | Case Count | AC Bullets Covered |
|---|---|---|---|---|
| E1-S1 | User Login | TS-01 | 7 | AC1, AC2, AC3, AC4 |
| E1-S2 | Token Expiry | TS-02 | 4 | AC1, AC2 |
| E1-S3 | User Logout | TS-03 | 3 | AC1, AC2 |
| E5-S1 | Default Categories | TS-04 | 4 | AC1, AC2, AC3 |
| E2-S1 | Create Expense | TS-05 | 6 | AC1, AC2, AC3, AC4 |
| E2-S2 | List Expenses (Paginated) | TS-06 | 5 | AC1, AC2, AC3, AC4, AC5 |
| E2-S3 | Filter Expenses by Date Range | TS-07 | 5 | AC1, AC2, AC3 |
| E2-S4 | Filter Expenses by Category | TS-08 | 4 | AC1, AC2, AC3 |
| E2-S5 | Edit Expense | TS-09 | 4 | AC1, AC2, AC3 |
| E2-S6 | Delete Expense | TS-10 | 4 | AC1, AC2, AC3 |
| E2-S8 | Running Total | TS-11 | 3 | AC1, AC2 |
| E3-S1 | Create Income | TS-12 | 5 | AC1, AC2, AC3 |
| E3-S2 | List Income (Paginated) | TS-13 | 4 | AC1, AC2, AC3 |
| E3-S3 | Filter Income by Date Range | TS-14 | 3 | AC1, AC2 |
| E3-S4 | Edit and Delete Income | TS-15 | 4 | AC1, AC2, AC3 |
| E4-S1 | Create Invoice/Bill | TS-16 | 6 | AC1, AC2, AC3 |
| E4-S2 | List Invoices with Status Filter | TS-17 | 5 | AC1, AC2, AC3, AC4 |
| E4-S3 | Mark Invoice as Paid | TS-18 | 5 | AC1, AC2, AC3, AC4 |
| E4-S4 | Overdue Invoice Highlighting | TS-19 | 5 | AC1, AC2, AC3, AC4 |
| E4-S6 | Edit and Delete Invoice | TS-20 | 4 | AC1, AC2, AC3 |
| E4-S8 | Outstanding Invoice Total | TS-21 | 3 | AC1, AC2 |
| E6-S1 | Dashboard Summary Cards | TS-22 | 5 | AC1, AC2, AC3 |
| E6-S6 | Invoice Summary on Dashboard | TS-23 | 3 | AC1, AC2 |
| M1-01 | Service Layer Unit Tests | TS-24 | 5 | AC1, AC2, AC3 |

**Total planned test cases: 106**

---

## E1-S1 — User Login

**Ticket**: E1-S1 | **NF**: AU-01, AU-02, AU-03, NF-06, NF-08, NF-09, NF-14 | **Priority**: Must  
**Files under test**: `internal/services/auth_service.go`, `internal/handlers/auth_handler.go`, `src/pages/LoginPage.tsx`, `src/hooks/useAuth.ts`, `src/components/ProtectedRoute.tsx`  
**Endpoints**: `POST /api/auth/login`, all protected `/api/*` routes

### Preconditions

- Backend running in `m1-seeded` environment.
- Seed user `admin` / `changeme` is present in the DB.
- Frontend dev server running at `http://localhost:5173`.

### Test Suite TS-01

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-01-01 | Smoke | Valid credentials — JWT returned and redirect to dashboard | 1. Navigate to `http://localhost:5173/login`. 2. Enter `admin` / `changeme`. 3. Click Submit. | HTTP 200 from `POST /api/auth/login`; body contains `{"data":{"token":"...","expires_at":"..."}}`. Frontend stores token in `localStorage`. Browser redirects to `/`. Dashboard page renders. | AC1 |
| TS-01-02 | Smoke | Invalid password — 401 and error message shown | 1. Navigate to `/login`. 2. Enter `admin` / `wrongpassword`. 3. Click Submit. | Server returns HTTP 401. Form displays "Invalid username or password". No redirect. | AC2 |
| TS-01-03 | Regression | Blank username — client-side validation fires | 1. Navigate to `/login`. 2. Leave username empty, enter any password. 3. Click Submit. | No HTTP request made. Field-level error shown under username: required field message (NF-14). | AC3 |
| TS-01-04 | Regression | Blank password — client-side validation fires | 1. Navigate to `/login`. 2. Enter `admin`, leave password empty. 3. Click Submit. | No HTTP request made. Field-level error shown under password (NF-14). | AC3 |
| TS-01-05 | Regression | Protected route without token — redirect to login | 1. Clear `localStorage`. 2. Navigate directly to `http://localhost:5173/`. | Browser redirects to `/login`; dashboard content not rendered. | AC4 |
| TS-01-06 | Regression | Auth header injection on API requests | 1. Log in successfully (TS-01-01). 2. Navigate to any data page (e.g., `/expenses`). 3. Open DevTools → Network tab. Inspect any `GET /api/expenses` request. | Request includes `Authorization: Bearer <token>` header. | AU-02, AU-03 |
| TS-01-07 | Regression | **Negative**: curl protected route without token | `curl -s http://localhost:8080/api/expenses` | HTTP 401; body is `{"error":"..."}` (no stack trace, no internal detail per NF-09). | AU-03, NF-09 |

---

## E1-S2 — Token Expiry

**Ticket**: E1-S2 | **NF**: AU-05, NF-06 | **Priority**: Should  
**Files under test**: `internal/services/auth_service.go` (`Login`, `ValidateToken`), `internal/handlers/middleware.go`, `src/api/client.ts`

### Preconditions

- Backend running in `m1-local` environment.
- `TOKEN_EXPIRY_HOURS` env var accessible and configurable.

### Test Suite TS-02

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-02-01 | Smoke | Token carries `exp` claim — inspectable after login | 1. `POST /api/auth/login` with valid credentials. 2. Base64-decode the JWT payload. | Payload contains `exp` claim set to approximately `now + TOKEN_EXPIRY_HOURS * 3600`. | AC2 |
| TS-02-02 | Regression | Expired token — server returns 401 | 1. Obtain a token. 2. Wait for expiry OR manually craft a JWT with `exp` set in the past (use the same `JWT_SECRET`). 3. `curl -H "Authorization: Bearer <expired>" http://localhost:8080/api/expenses`. | HTTP 401; body contains `{"error":"token expired"}` or similar. | AC1 |
| TS-02-03 | Regression | Expired token in browser — redirect to `/login` | 1. Log in. 2. Replace `localStorage.token` with a crafted expired JWT (same secret). 3. Navigate to a protected page. | Frontend detects 401 (or proactively checks `exp`), clears localStorage, and redirects to `/login`. | AC1 |
| TS-02-04 | Regression | `TOKEN_EXPIRY_HOURS=1` — token expires after 1 hour | 1. Set `TOKEN_EXPIRY_HOURS=1`, restart backend. 2. Log in. 3. Decode JWT payload. | `exp − iat ≈ 3600` seconds. | AC2 |

---

## E1-S3 — User Logout

**Ticket**: E1-S3 | **NF**: AU-04 | **Priority**: Must  
**Files under test**: `src/components/Sidebar.tsx`, `src/hooks/useAuth.ts`  
**Endpoint**: `POST /api/auth/logout`

### Preconditions

- User is logged in (token present in `localStorage`).

### Test Suite TS-03

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-03-01 | Smoke | Logout clears token and redirects to `/login` | 1. Log in as `admin`. 2. Click the Logout button in the sidebar/navbar. | `localStorage.token` is removed. Browser redirects to `/login`. Dashboard is no longer accessible without re-login. | AC1 |
| TS-03-02 | Regression | After logout, protected route inaccessible | 1. Log out (TS-03-01). 2. Click browser Back or navigate to `/`. | Redirected to `/login`; no cached protected content shown. | AC2 |
| TS-03-03 | Regression | `POST /api/auth/logout` returns 200 (stateless JWT) | `curl -s -X POST -H "Authorization: Bearer <valid_token>" http://localhost:8080/api/auth/logout` | HTTP 200; server-side: no error. (Token is not invalidated server-side in MVP — stateless approach per ticket.) | AC1 |

---

## E5-S1 — Default Categories

**Ticket**: E5-S1 | **NF**: NF-09 | **Priority**: Must  
**Files under test**: `backend/migrations/003_create_categories.up.sql`, `004_seed_default_categories.up.sql`, `internal/services/category_service.go`, `internal/handlers/category_handler.go`  
**Endpoint**: `GET /api/categories?type=expense|income`

### Preconditions

- Backend started fresh (migrations applied).
- Valid auth token available.

### Test Suite TS-04

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-04-01 | Smoke | Default expense categories present after migration | `sqlite3 moneyapp.db "SELECT COUNT(*) FROM categories WHERE type='expense' AND is_default=1"` | Returns `9`. Categories include at minimum: Food, Transport, Housing, Health, Entertainment, Shopping, Utilities, Other, Uncategorized. | AC1 |
| TS-04-02 | Smoke | Default income categories present after migration | `sqlite3 moneyapp.db "SELECT COUNT(*) FROM categories WHERE type='income' AND is_default=1"` | Returns `6`. Categories include at minimum: Salary, Freelance, Investment, Gift, Other, Uncategorized. | AC1 |
| TS-04-03 | Smoke | `GET /api/categories?type=expense` returns expense categories | `curl -s -H "Authorization: Bearer <token>" http://localhost:8080/api/categories?type=expense` | HTTP 200; JSON array with 9 objects, each having `id`, `name`, `type=expense`, `is_default=true`. | AC2 |
| TS-04-04 | Regression | **Negative**: Unauthenticated request to categories endpoint | `curl -s http://localhost:8080/api/categories?type=expense` | HTTP 401. | AC3 |

---

## E2-S1 — Create Expense

**Ticket**: E2-S1 | **NF**: NF-10, NF-14 | **Priority**: Must  
**Files under test**: `internal/services/expense_service.go`, `internal/handlers/expense_handler.go`, `src/components/expenses/ExpenseForm.tsx`  
**Endpoint**: `POST /api/expenses`

### Preconditions

- Valid auth token present.
- Default categories seeded (`E5-S1` complete).

### Test Suite TS-05

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-05-01 | Smoke | Create expense with all valid fields | `curl -s -X POST -H "Authorization: Bearer <token>" -H "Content-Type: application/json" -d '{"amount":50000,"date":"2026-04-26","category_id":1,"description":"Lunch"}' http://localhost:8080/api/expenses` | HTTP 201; body contains `{"data":{"id":...,"amount":50000,"date":"2026-04-26","category_id":1,"description":"Lunch",...}}`. Row appears in `expenses` table. | AC1 |
| TS-05-02 | Smoke | Amount stored as minor units (integer) | 1. Create expense with `amount: 50000`. 2. `sqlite3 moneyapp.db "SELECT amount FROM expenses WHERE id=<id>"`. | Returns `50000` (integer, not float). (NF-10) | AC1, NF-10 |
| TS-05-03 | Regression | **Negative**: Zero amount rejected | `curl ... -d '{"amount":0,"date":"2026-04-26","category_id":1}'` | HTTP 400; body contains error referencing amount > 0. | AC2 |
| TS-05-04 | Regression | **Negative**: Negative amount rejected | `curl ... -d '{"amount":-100,"date":"2026-04-26","category_id":1}'` | HTTP 400; validation error. | AC2 |
| TS-05-05 | Regression | Missing date — defaults to today | `curl ... -d '{"amount":10000,"category_id":1}'` (no `date` field) | HTTP 201; response `date` field equals today's date (YYYY-MM-DD). | AC3 |
| TS-05-06 | Regression | **Negative**: Invalid category_id | `curl ... -d '{"amount":10000,"date":"2026-04-26","category_id":99999}'` | HTTP 400; descriptive error about invalid category. | AC4 |

#### Frontend checks (E2-S1)

| Check | Steps | Expected |
|---|---|---|
| Amount field converts from display to minor units | Enter `500` in the amount field (displayed as major units, e.g. `500 VND`). Submit. | API receives `amount: 500` (or `50000` if 1 VND = 100 minor units — verify against `CURRENCY` config). Body does not contain a float. |
| Category dropdown loads expense categories | Open "Add Expense" form. | Dropdown is populated from `GET /api/categories?type=expense`. |
| Client-side validation: empty amount | Click Submit with amount empty. | Field-level error visible without making an API call (NF-14). |

---

## E2-S2 — List Expenses (Paginated)

**Ticket**: E2-S2 | **NF**: NF-01, NF-04, NF-17 | **Priority**: Must  
**Files under test**: `internal/services/expense_service.go` (`List`), `internal/handlers/expense_handler.go`, `src/pages/ExpensesPage.tsx`  
**Endpoint**: `GET /api/expenses?page=1&per_page=20`

### Preconditions

- Auth token valid.
- At least 25 expense records exist (can be created via TS-05-01 loop).

### Test Suite TS-06

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-06-01 | Smoke | Default pagination — 20 items on page 1 | `curl -s -H "Authorization: Bearer <token>" "http://localhost:8080/api/expenses?page=1&per_page=20"` (25 records exist) | HTTP 200; `data` array has 20 items; `total=25`; `page=1`; `per_page=20`. Response includes `category_name` field on each item. | AC1 |
| TS-06-02 | Smoke | Page 2 — remaining items returned | `curl ... "?page=2&per_page=20"` | `data` array has 5 items; `page=2`. | AC1 |
| TS-06-03 | Regression | Empty state — no expenses exist | With zero expenses: `curl ... "?page=1&per_page=20"` | HTTP 200; `data` is empty array `[]`; `total=0`. Frontend shows empty state message "No expenses recorded yet". | AC3 |
| TS-06-04 | Regression | Loading indicator shown during fetch | Navigate to `/expenses` in browser (simulate slow network in DevTools). | A loading spinner or skeleton is visible while the API call is in flight. | AC4 |
| TS-06-05 | Regression | Performance — 10,000 records under 500ms | Insert 10,000 expense rows (script). `curl -o /dev/null -s -w "%{time_total}" "http://localhost:8080/api/expenses?page=1&per_page=20"` | Response time < 0.5 seconds. (NF-01, NF-04) | AC5, NF-01, NF-04 |

#### Frontend visual checklist (E2-S2)

- [ ] Amount displayed in major currency units with locale formatting (e.g., `50,000 VND`) — NF-17
- [ ] Pagination controls (Previous/Next, page number) visible and functional
- [ ] Category name column populated (not raw `category_id`)
- [ ] Date formatted consistently (YYYY-MM-DD or locale date string)

---

## E2-S3 — Filter Expenses by Date Range

**Ticket**: E2-S3 | **NF**: NF-04 | **Priority**: Must  
**Files under test**: `internal/services/expense_service.go` (`List` date params), `src/components/filters/DateRangeFilter.tsx`  
**Endpoint**: `GET /api/expenses?date_from=...&date_to=...`

### Preconditions

- Expenses created in January 2026 (≥ 3) and February 2026 (≥ 3) are present.
- Valid auth token.

### Test Suite TS-07

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-07-01 | Smoke | Filter returns only expenses in date range | `curl -s -H "Authorization: Bearer <token>" "http://localhost:8080/api/expenses?date_from=2026-01-01&date_to=2026-01-31"` | Only January 2026 expenses returned; February records absent. | AC1 |
| TS-07-02 | Smoke | Clear filter returns all expenses | Apply filter. Click Clear. `GET /api/expenses` (no date params). | All expenses returned; no date filter in request. | AC2 |
| TS-07-03 | Regression | URL persistence — filter survives page refresh | 1. Apply filter `date_from=2026-01-01&date_to=2026-01-31`. 2. Verify URL contains `?date_from=2026-01-01&date_to=2026-01-31`. 3. Refresh the page. | Filter inputs are pre-populated; API is called with same date params; filtered results are shown. | AC3 |
| TS-07-04 | Regression | **Negative**: `date_from` after `date_to` | `curl ... "?date_from=2026-02-01&date_to=2026-01-01"` | HTTP 400; error message about invalid date range. | AC1 (implicit validation) |
| TS-07-05 | Regression | **Negative**: Invalid date format | `curl ... "?date_from=26-01-01&date_to=26-01-31"` | HTTP 400; error message about date format (expected YYYY-MM-DD). | Ticket validation note |

---

## E2-S4 — Filter Expenses by Category

**Ticket**: E2-S4 | **NF**: — | **Priority**: Must  
**Files under test**: `internal/services/expense_service.go` (category filter), `src/components/filters/CategoryFilter.tsx`  
**Endpoint**: `GET /api/expenses?category_id=...` or `?category_ids=1,2,3`

### Preconditions

- Expenses created in at least two different categories (e.g., Food and Transport).
- Valid auth token.

### Test Suite TS-08

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-08-01 | Smoke | Filter by single category | `curl -s -H "Authorization: Bearer <token>" "http://localhost:8080/api/expenses?category_id=<food_id>"` | Only Food expenses returned; Transport expenses absent. | AC1 |
| TS-08-02 | Regression | Filter by multiple categories | `curl ... "?category_ids=<food_id>,<transport_id>"` | Expenses from both Food and Transport returned; no other categories in result. | AC2 |
| TS-08-03 | Regression | Clear category filter shows all expenses | Apply category filter. Click Clear. | All expenses shown; no category filter applied. | AC3 |
| TS-08-04 | Regression | Category filter in URL params | Apply category filter; verify URL includes `category_id=<id>`. Refresh page. | Filter is restored from URL; same filtered results shown. | AC3 (URL persistence) |

---

## E2-S5 — Edit Expense

**Ticket**: E2-S5 | **NF**: NF-14 | **Priority**: Must  
**Files under test**: `internal/services/expense_service.go` (`Update`), `internal/handlers/expense_handler.go`, `src/components/expenses/ExpenseForm.tsx`  
**Endpoints**: `GET /api/expenses/:id`, `PUT /api/expenses/:id`

### Preconditions

- At least one expense record exists (created via TS-05-01).
- Valid auth token.

### Test Suite TS-09

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-09-01 | Smoke | Edit expense — form pre-populated | 1. Navigate to `/expenses`. 2. Click Edit on an expense. | Form opens with all current values (amount, date, category, description) pre-filled. | AC1 |
| TS-09-02 | Smoke | Save edited amount — reflected in list and DB | 1. Open Edit form. 2. Change amount to `99999`. 3. Save. | `PUT /api/expenses/:id` returns 200 with updated amount `99999`. List row reflects new amount. DB `SELECT amount FROM expenses WHERE id=<id>` = 99999. `updated_at` is newer than original. | AC2 |
| TS-09-03 | Regression | **Negative**: Edit with invalid amount | Open Edit form. Set amount to `0`. Click Save. | HTTP 400 or client-side validation prevents submission; error message shown. | AC2 |
| TS-09-04 | Regression | **Negative**: Edit non-existent expense | `curl -s -X PUT -H "Authorization: Bearer <token>" -H "Content-Type: application/json" -d '{"amount":1000,"date":"2026-04-26","category_id":1}' http://localhost:8080/api/expenses/99999` | HTTP 404. | AC3 |

---

## E2-S6 — Delete Expense

**Ticket**: E2-S6 | **NF**: NF-15 | **Priority**: Must  
**Files under test**: `internal/services/expense_service.go` (`Delete`), `internal/handlers/expense_handler.go`, `src/components/ConfirmDialog.tsx`  
**Endpoint**: `DELETE /api/expenses/:id`

### Preconditions

- At least two expense records exist.
- Valid auth token.

### Test Suite TS-10

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-10-01 | Smoke | Delete expense — removed from list and DB | 1. Note expense ID. 2. Click Delete. 3. Confirm dialog shown. 4. Click Confirm. | `DELETE /api/expenses/:id` returns 204. Expense no longer appears in the list. `SELECT count(*) FROM expenses WHERE id=<id>` = 0. | AC1 |
| TS-10-02 | Regression | Cancel delete — no data modified | 1. Click Delete on an expense. 2. Click Cancel in the confirmation dialog. | No API call made. Expense remains in the list and DB. | AC2, NF-15 |
| TS-10-03 | Regression | Confirmation dialog required (NF-15) | Click Delete on any expense. | Confirmation dialog appears with message "Are you sure you want to delete this expense?" before any deletion occurs. | AC1, NF-15 |
| TS-10-04 | Regression | **Negative**: Delete non-existent expense | `curl -s -X DELETE -H "Authorization: Bearer <token>" http://localhost:8080/api/expenses/99999` | HTTP 404. | AC3 |

---

## E2-S8 — Running Total for Filtered Expenses

**Ticket**: E2-S8 | **NF**: NF-17 | **Priority**: Must  
**Files under test**: `internal/services/expense_service.go` (`List` — `total_amount`), `src/pages/ExpensesPage.tsx`  
**Endpoint**: `GET /api/expenses` — `total_amount` field in response

### Preconditions

- Three expenses with amounts 100,000, 200,000, and 200,000 exist dated in January 2026.

### Test Suite TS-11

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-11-01 | Smoke | `total_amount` field present in list response | `curl -s -H "Authorization: Bearer <token>" "http://localhost:8080/api/expenses?page=1&per_page=20"` | Response JSON contains `"total_amount": <sum_of_visible_amounts>`. Value matches the sum of all filtered expense amounts (integer). | AC1, NF-10 |
| TS-11-02 | Smoke | Total shown in UI and formatted correctly | Navigate to `/expenses` with no filters. | A summary bar above the table displays "Total: X VND" (NF-17) where X matches the API `total_amount`. | AC1, NF-17 |
| TS-11-03 | Regression | Total updates when date filter applied | Apply date range filter for January 2026. | Summary bar updates to show 500,000 VND (or appropriate sum for filtered data). Total reflects only filtered expenses, not all. | AC2 |

---

## E3-S1 — Create Income

**Ticket**: E3-S1 | **NF**: NF-10, NF-14 | **Priority**: Must  
**Files under test**: `internal/services/income_service.go`, `internal/handlers/income_handler.go`, `src/components/income/IncomeForm.tsx`  
**Endpoint**: `POST /api/incomes`

### Preconditions

- Valid auth token. Default income categories seeded (type=`income`).

### Test Suite TS-12

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-12-01 | Smoke | Create income with valid fields | `curl -s -X POST -H "Authorization: Bearer <token>" -H "Content-Type: application/json" -d '{"amount":10000000,"date":"2026-04-26","category_id":<salary_id>,"description":"April salary"}' http://localhost:8080/api/incomes` | HTTP 201; response contains `id`, `amount: 10000000`, `date: "2026-04-26"`. Persisted in `incomes` table. | AC1 |
| TS-12-02 | Regression | **Negative**: Zero amount rejected | `curl ... -d '{"amount":0,"date":"2026-04-26","category_id":<salary_id>}'` | HTTP 400; error message referencing amount > 0. | AC2 |
| TS-12-03 | Regression | **Negative**: Income category used for expense type rejected | `curl ... -d '{"amount":5000,"date":"2026-04-26","category_id":<expense_category_id>}'` | HTTP 400; error indicating category type mismatch ("income category required"). | AC3 |
| TS-12-04 | Regression | Income list has `total_amount` field | `GET /api/incomes?page=1&per_page=20` | Response includes `"total_amount"` summing income amounts for the current filter. | E3-S1 income list parity with expenses |
| TS-12-05 | Regression | Income supports date_from/date_to filter | `GET /api/incomes?date_from=2026-01-01&date_to=2026-01-31` | Only January income records returned. | E3-S3 parity |

---

## E3-S2 — List Income (Paginated)

**Ticket**: E3-S2 | **NF**: NF-01, NF-04, NF-17 | **Priority**: Must  
**Files under test**: `src/pages/IncomePage.tsx`  
**Endpoint**: `GET /api/incomes?page=1&per_page=20`

### Preconditions

- At least 25 income records exist.

### Test Suite TS-13

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-13-01 | Smoke | Default pagination — 20 items on page 1 | `GET /api/incomes?page=1&per_page=20` (25 records) | `data` has 20 items; `total=25`; `page=1`. | AC1 |
| TS-13-02 | Smoke | Page 2 — remaining items | `GET /api/incomes?page=2&per_page=20` | `data` has 5 items; `page=2`. | AC1 |
| TS-13-03 | Regression | Empty state message displayed | With zero income records, navigate to `/income`. | Empty state message shown (e.g., "No income records yet"). | AC2 |
| TS-13-04 | Regression | Performance — 10,000 records under 500ms | `curl -o /dev/null -s -w "%{time_total}" "http://localhost:8080/api/incomes?page=1&per_page=20"` | Response time < 0.5 seconds (NF-01, NF-04). | AC3, NF-01 |

---

## E3-S3 — Filter Income by Date Range

**Ticket**: E3-S3 | **NF**: — | **Priority**: Must  
**Files under test**: `src/pages/IncomePage.tsx` (reuses `DateRangeFilter.tsx`)  
**Endpoint**: `GET /api/incomes?date_from=...&date_to=...`

### Preconditions

- Income records in January 2026 and February 2026.

### Test Suite TS-14

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-14-01 | Smoke | Filter returns only January income | `GET /api/incomes?date_from=2026-01-01&date_to=2026-01-31` | Only January records returned. | AC1 |
| TS-14-02 | Regression | Clear filter shows all records | Apply filter; click Clear. | All income records shown. | AC2 |
| TS-14-03 | Regression | URL persistence — filter survives refresh | Apply filter; verify URL params; refresh page. | Filter restored from URL; same filtered results shown. | AC2 (parity with E2-S3 AC3) |

---

## E3-S4 — Edit and Delete Income

**Ticket**: E3-S4 | **NF**: NF-15 | **Priority**: Must  
**Files under test**: `internal/services/income_service.go` (`Update`, `Delete`), `src/components/income/IncomeForm.tsx`, `src/components/ConfirmDialog.tsx`  
**Endpoints**: `PUT /api/incomes/:id`, `DELETE /api/incomes/:id`

### Preconditions

- At least two income records exist.

### Test Suite TS-15

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-15-01 | Smoke | Edit income — changes reflected in list and DB | 1. Click Edit on an income. 2. Change amount to `12000000`. 3. Save. | `PUT /api/incomes/:id` returns 200; list row shows updated amount; DB updated. | AC1 |
| TS-15-02 | Regression | Delete income with confirm | Click Delete → Confirm. | `DELETE /api/incomes/:id` returns 204; record removed from list and DB. | AC2 |
| TS-15-03 | Regression | Cancel delete — no change | Click Delete → Cancel. | No API call. Record unchanged. | AC3 |
| TS-15-04 | Regression | **Negative**: PUT non-existent income | `curl -X PUT ... http://localhost:8080/api/incomes/99999` | HTTP 404. | AC1 (parity) |

---

## E4-S1 — Create Invoice/Bill

**Ticket**: E4-S1 | **NF**: NF-10, NF-14 | **Priority**: Must  
**Files under test**: `internal/services/invoice_service.go` (`Create`), `internal/handlers/invoice_handler.go`, `src/components/invoices/InvoiceForm.tsx`  
**Endpoint**: `POST /api/invoices`

### Preconditions

- Valid auth token. No categories dependency (invoices do not use categories).

### Test Suite TS-16

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-16-01 | Smoke | Create invoice with all valid fields | `curl -s -X POST -H "Authorization: Bearer <token>" -H "Content-Type: application/json" -d '{"vendor_name":"Acme Corp","amount":5000000,"issue_date":"2026-04-01","due_date":"2026-04-30","status":"unpaid","description":"April hosting"}' http://localhost:8080/api/invoices` | HTTP 201; response includes `id`, `vendor_name: "Acme Corp"`, `amount: 5000000`, `status: "unpaid"`. Row in `invoices` table. | AC1 |
| TS-16-02 | Regression | **Negative**: Missing vendor name | `curl ... -d '{"vendor_name":"","amount":5000000,"issue_date":"2026-04-01","due_date":"2026-04-30"}'` | HTTP 400; error referencing vendor_name required. | AC2 |
| TS-16-03 | Regression | **Negative**: Zero amount | `curl ... -d '{"vendor_name":"Corp","amount":0,"issue_date":"2026-04-01","due_date":"2026-04-30"}'` | HTTP 400; validation error. | AC2 |
| TS-16-04 | Regression | **Negative**: due_date before issue_date | `curl ... -d '{"vendor_name":"Corp","amount":1000,"issue_date":"2026-04-30","due_date":"2026-04-01","status":"unpaid"}'` | HTTP 400; error "Due date must be on or after issue date". | AC3 |
| TS-16-05 | Regression | due_date equal to issue_date is valid | `curl ... -d '{"vendor_name":"Corp","amount":1000,"issue_date":"2026-04-15","due_date":"2026-04-15","status":"unpaid"}'` | HTTP 201; invoice created successfully. | AC3 (boundary) |
| TS-16-06 | Regression | Amount stored as integer (NF-10) | Create invoice with `amount: 5000000`. Query DB. | `SELECT amount FROM invoices WHERE id=<id>` returns `5000000` as integer. | NF-10 |

---

## E4-S2 — List Invoices with Status Filter

**Ticket**: E4-S2 | **NF**: NF-01 | **Priority**: Must  
**Files under test**: `internal/services/invoice_service.go` (`List`), `src/pages/InvoicesPage.tsx`  
**Endpoint**: `GET /api/invoices?page=1&per_page=20&status=unpaid`

### Preconditions

- At least 3 unpaid, 2 paid, and 2 overdue invoices exist.
- Valid auth token.

### Test Suite TS-17

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-17-01 | Smoke | Filter by `status=unpaid` | `GET /api/invoices?status=unpaid` | Only unpaid invoices returned. Paid and overdue absent. | AC1 |
| TS-17-02 | Smoke | No status filter — all invoices returned | `GET /api/invoices` | All invoices returned regardless of status. | AC2 |
| TS-17-03 | Regression | Status filter in URL — persists on refresh | Select "Unpaid" tab on `/invoices` page. Verify URL: `/invoices?status=unpaid`. Refresh. | Filter restored; only unpaid invoices shown. | AC4 |
| TS-17-04 | Regression | Status badges present with correct colors | Navigate to `/invoices` (all tab). | Unpaid = yellow badge, Paid = green badge, Overdue = red badge. | AC1 (UI) |
| TS-17-05 | Regression | Performance — 10,000 records under 500ms | Insert 10,000 invoice rows. `curl -o /dev/null -s -w "%{time_total}" "http://localhost:8080/api/invoices?page=1&per_page=20"` | Response time < 0.5 seconds (NF-01). | AC3, NF-01 |

---

## E4-S3 — Mark Invoice as Paid

**Ticket**: E4-S3 | **NF**: NF-15 | **Priority**: Must  
**Files under test**: `internal/handlers/invoice_handler.go`, `src/pages/InvoicesPage.tsx`  
**Endpoint**: `PATCH /api/invoices/:id/status`

### Preconditions

- Unpaid invoice with known ID exists.
- Valid auth token.

### Test Suite TS-18

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-18-01 | Smoke | Mark unpaid invoice as paid | 1. Click "Mark as Paid" on an unpaid invoice. 2. Confirm dialog. | `PATCH /api/invoices/:id/status` with `{"status":"paid"}` returns 200. Invoice status in UI and DB changes to `paid`. Green badge shown. | AC1 |
| TS-18-02 | Regression | "Mark as Paid" button hidden on paid invoices | Navigate to paid invoices (status filter). | No "Mark as Paid" button visible on paid invoice rows. | AC2 |
| TS-18-03 | Regression | Confirmation dialog required (NF-15) | Click "Mark as Paid". | Dialog "Mark this invoice from [Vendor] as paid?" appears before any change. | NF-15 |
| TS-18-04 | Regression | **Negative**: Mark already-paid invoice as paid | `curl -s -X PATCH -H "Authorization: Bearer <token>" -H "Content-Type: application/json" -d '{"status":"paid"}' http://localhost:8080/api/invoices/<paid_id>/status` | HTTP 422; body `{"error":"Invoice is already paid"}`. | AC3 |
| TS-18-05 | Regression | **Negative**: Invalid target status via PATCH | `curl ... -d '{"status":"unpaid"}' http://localhost:8080/api/invoices/<id>/status` | HTTP 400; body `{"error":"Only 'paid' is a valid target status via this endpoint"}`. | AC4 |

---

## E4-S4 — Overdue Invoice Highlighting

**Ticket**: E4-S4 | **NF**: — | **Priority**: Must  
**Files under test**: `internal/services/invoice_service.go` (`UpdateOverdueStatuses`), `src/pages/InvoicesPage.tsx`

### Preconditions

- At least one invoice with `status=unpaid` and `due_date` in the past exists.
- At least one invoice with `status=unpaid` and `due_date` in the future exists.
- Valid auth token.

### Test Suite TS-19

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-19-01 | Smoke | Past-due unpaid invoice is shown as overdue after page load | 1. Ensure invoice with `due_date < today` and `status=unpaid`. 2. `GET /api/invoices`. | Invoice is returned with `status=overdue`. Frontend shows red badge with warning icon. | AC1 |
| TS-19-02 | Smoke | Future-due unpaid invoice remains unpaid | 1. Ensure invoice with `due_date > today` and `status=unpaid`. 2. `GET /api/invoices`. | Invoice returned with `status=unpaid`. | AC2 |
| TS-19-03 | Regression | Overdue transition persisted in DB (not computed on the fly) | After `GET /api/invoices` triggers `UpdateOverdueStatuses`: `sqlite3 moneyapp.db "SELECT status FROM invoices WHERE id=<past_due_id>"` | Returns `overdue` (not `unpaid`). DB was updated. | AC3 |
| TS-19-04 | Regression | Server startup updates overdue statuses | 1. Insert unpaid invoice with `due_date=2020-01-01` (clearly past). 2. Restart backend. 3. Check server logs. 4. `sqlite3 ... "SELECT status FROM invoices WHERE id=<id>"` | Logs contain "Updated N invoices to overdue status" (N ≥ 1). DB shows `status=overdue`. | AC4 |
| TS-19-05 | Regression | Paid invoice not transitioned to overdue | Create a paid invoice with `due_date` in the past. Call `GET /api/invoices`. | Invoice remains `paid`; `UpdateOverdueStatuses` does not affect paid invoices. | AC3 (negative) |

---

## E4-S6 — Edit and Delete Invoice

**Ticket**: E4-S6 | **NF**: NF-15 | **Priority**: Must  
**Files under test**: `internal/services/invoice_service.go` (`Update`, `Delete`), `src/components/invoices/InvoiceForm.tsx`  
**Endpoints**: `GET /api/invoices/:id`, `PUT /api/invoices/:id`, `DELETE /api/invoices/:id`

### Preconditions

- At least two invoice records exist.

### Test Suite TS-20

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-20-01 | Smoke | Edit invoice vendor name — reflected in list | 1. Click Edit on invoice. 2. Change vendor name to "New Vendor". 3. Save. | `PUT /api/invoices/:id` returns 200. List row shows "New Vendor". | AC1 |
| TS-20-02 | Regression | Delete invoice with confirm | Click Delete → Confirm. | `DELETE /api/invoices/:id` returns 204. Invoice removed from list and DB. | AC2 |
| TS-20-03 | Regression | Cancel delete — no change | Click Delete → Cancel. | No API call. Invoice unchanged. (NF-15) | AC3 |
| TS-20-04 | Regression | **Negative**: GET/PUT/DELETE non-existent invoice | `curl -s -H "Authorization: Bearer <token>" http://localhost:8080/api/invoices/99999` | HTTP 404 for GET. `PUT` returns 404. `DELETE` returns 404. | AC2 (parity) |

---

## E4-S8 — Outstanding Invoice Total

**Ticket**: E4-S8 | **NF**: NF-17 | **Priority**: Must  
**Files under test**: `internal/handlers/invoice_handler.go` (`GET /api/invoices/stats`), `src/pages/InvoicesPage.tsx`

> **Routing note**: The endpoint uses `/api/invoices/stats` (not `/api/invoices/summary`) to avoid the Go `net/http` mux matching "summary" as the `:id` path parameter. When testing, ensure the request goes to `/api/invoices/stats` — any implementation using `/api/invoices/summary` would be a routing bug.

### Preconditions

- 3 unpaid invoices (total 10,000,000) and 2 overdue invoices (total 5,000,000) exist.

### Test Suite TS-21

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-21-01 | Smoke | `/api/invoices/stats` returns correct outstanding total | `curl -s -H "Authorization: Bearer <token>" http://localhost:8080/api/invoices/stats` | HTTP 200; body `{"total_outstanding":15000000,"unpaid_count":3,"overdue_count":2}`. | AC1 |
| TS-21-02 | Smoke | UI summary bar shows correct values | Navigate to `/invoices` (All tab). | Summary bar: "Outstanding: 15,000,000 VND (3 unpaid, 2 overdue)" or equivalent formatting. (NF-17) | AC1, NF-17 |
| TS-21-03 | Regression | Total updates after marking invoice as paid | Mark one unpaid invoice as paid. Reload `/invoices`. | Summary bar decreases by that invoice's amount; unpaid count decreases by 1. | AC2 |

---

## E6-S1 — Dashboard Summary Cards

**Ticket**: E6-S1 | **NF**: NF-02, NF-10, NF-17 | **Priority**: Must  
**Files under test**: `internal/services/dashboard_service.go` (`GetSummary`), `internal/handlers/dashboard_handler.go`, `src/pages/DashboardPage.tsx`, `src/components/dashboard/SummaryCard.tsx`  
**Endpoint**: `GET /api/dashboard/summary?date_from=...&date_to=...`

### Preconditions

- Income records for April 2026 totaling 20,000,000 exist.
- Expense records for April 2026 totaling 5,000,000 exist.
- Valid auth token.

### Test Suite TS-22

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-22-01 | Smoke | Dashboard summary for current month — correct totals | `curl -s -H "Authorization: Bearer <token>" "http://localhost:8080/api/dashboard/summary?date_from=2026-04-01&date_to=2026-04-30"` | HTTP 200; `total_income=20000000`, `total_expenses=5000000`, `net_balance=15000000`. | AC1 |
| TS-22-02 | Smoke | Three summary cards visible on dashboard | Navigate to `http://localhost:5173/` after login. | Three cards visible: "Total Income" (green), "Total Expenses" (red), "Net Balance" (green when positive, red when negative). Amounts formatted per NF-17. | AC1, NF-17 |
| TS-22-03 | Regression | Default period — current calendar month (no date params) | `curl -s -H "Authorization: Bearer <token>" "http://localhost:8080/api/dashboard/summary"` | Response defaults to current month (e.g., April 2026-04-01 to 2026-04-30); `date_from` and `date_to` fields in response reflect current month bounds. | AC1 |
| TS-22-04 | Regression | Zero records — all cards show 0 | Navigate to dashboard with no records for current month. | All three cards show 0 (not null, not empty). | AC2 |
| TS-22-05 | Regression | Loading skeleton shown during fetch | Navigate to `/` on slow connection (throttle in DevTools). | Skeleton or placeholder visible during API fetch (NF-02 implies < 3 seconds total). | AC3 |

---

## E6-S6 — Invoice Summary on Dashboard

**Ticket**: E6-S6 | **NF**: NF-17 | **Priority**: Must  
**Files under test**: `internal/services/dashboard_service.go` or separate invoice summary, `src/components/dashboard/InvoiceSummaryCard.tsx`

### Preconditions

- 3 unpaid invoices totaling 10,000,000 and 1 overdue invoice totaling 5,000,000 exist.

### Test Suite TS-23

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-23-01 | Smoke | Invoice summary card shows correct counts and totals | Navigate to `/` (dashboard). | Invoice summary section shows "4 outstanding invoices - 15,000,000 VND (3 unpaid, 1 overdue)" or equivalent. | AC1, NF-17 |
| TS-23-02 | Regression | Zero outstanding invoices — shows "no outstanding" message | Ensure all invoices are paid. Navigate to `/`. | Invoice summary shows "No outstanding invoices". | AC2 |
| TS-23-03 | Regression | Link on invoice summary navigates to `/invoices` | Click the invoice summary card or link. | Browser navigates to `/invoices` page. | AC1 (UX) |

---

## M1-01 — Service Layer Unit Test Coverage

**Ticket**: M1-01 | **NF**: NF-20 | **Priority**: Must  
**Files under test**: `internal/services/*_test.go`  
**Command**: `cd backend && go test ./internal/services/... -cover`

### Preconditions

- Go toolchain available.
- No external dependencies (tests use SQLite `:memory:`).

### Test Suite TS-24

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-24-01 | Smoke | All service tests pass | `cd backend && go test ./internal/services/... -count=1` | Exit code 0; all tests listed as PASS; no FAIL or panic lines. | AC1 |
| TS-24-02 | Smoke | Coverage ≥ 70% for `internal/services/` | `cd backend && go test ./internal/services/... -cover` | Output line: `coverage: XX.X% of statements` with XX.X ≥ 70. | AC2 |
| TS-24-03 | Regression | Expense amount validation test catches regression | Run tests; find test covering `ExpenseService.Create` with zero amount. | A dedicated test (e.g., `TestExpenseService_Create_ZeroAmount`) exists and fails if the `amount > 0` check is removed. | AC3 |
| TS-24-04 | Regression | Invoice date ordering test | Find test covering `InvoiceService.Create` with `due_date < issue_date`. | Test exists and passes; it would fail if the date validation were removed. | AC1 |
| TS-24-05 | Regression | `UpdateOverdueStatuses` does not affect paid invoices | Find test covering this negative case. | Test seeds a paid invoice with past due_date; asserts `status` remains `paid` after calling `UpdateOverdueStatuses`. | AC1 |

---

## Cross-Cutting Checks

These checks apply across the entire M1 feature set and should be verified after all tickets are implemented.

### Auth protection on all routes

| Check | Command | Expected |
|---|---|---|
| Every `/api/` route except `/api/auth/login` and `/api/health` requires auth | `for route in /api/expenses /api/incomes /api/invoices /api/categories /api/dashboard/summary /api/invoices/stats; do echo "$route: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080$route)"; done` | All return 401. |
| `GET /api/health` accessible without auth | `curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/health` | 200. |
| `POST /api/auth/login` accessible without auth | `curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"changeme"}' http://localhost:8080/api/auth/login` | 200. |

### Amount / integer formatting (NF-10)

| Check | Method | Expected |
|---|---|---|
| No float amounts in API responses | Create records; inspect all list/detail JSON responses for any float value (e.g., `50000.0`). | All `amount`, `total_amount`, `net_balance`, `total_income`, `total_expenses`, `total_outstanding` fields are JSON integers. |
| Frontend displays amounts formatted per NF-17 | View expense, income, invoice, and dashboard pages. | Amounts shown with locale thousands separator and currency symbol (e.g., `50,000 VND`), never as raw integers. |

### Pagination limits

| Check | Command | Expected |
|---|---|---|
| `per_page` capped at 100 | `curl ... "GET /api/expenses?page=1&per_page=9999"` | Response `per_page` is capped at 100; does not return more than 100 rows. |
| `per_page` defaults to 20 | `curl ... "GET /api/expenses?page=1"` (no per_page) | Response `per_page=20`. |

### URL persistence for filters

| Route | Filters that must persist in URL |
|---|---|
| `/expenses` | `date_from`, `date_to`, `category_id`, `page` |
| `/income` | `date_from`, `date_to`, `page` |
| `/invoices` | `status`, `page` |

**Test**: Apply each filter combination, copy the URL, open in a new tab. Verify the same filtered results are shown and the filter UI elements are populated.

### Invoice overdue routing sanity check

```bash
# Verify /api/invoices/stats is routed correctly (not caught as /:id)
curl -s -H "Authorization: Bearer <token>" http://localhost:8080/api/invoices/stats
# Expected: JSON with total_outstanding, unpaid_count, overdue_count

# Verify /api/invoices/:id still works with a numeric ID
curl -s -H "Authorization: Bearer <token>" http://localhost:8080/api/invoices/1
# Expected: Single invoice object (or 404 if id=1 doesn't exist)
```

### Error response quality (NF-09)

| Check | Steps | Expected |
|---|---|---|
| 400 responses include field-level errors | Submit invalid form data. | Body contains `{"error":"..."}` with a descriptive message; no Go error types, file paths, or stack traces. |
| 404 responses have consistent shape | GET non-existent expense, income, and invoice. | All return `{"error":"not found"}` (or similar); shape is consistent with other error responses. |
| No internal error leakage | Trigger any validation error. Inspect full response body. | No SQL statements, column names, or Go struct tags visible. |

### bcrypt password hashing (NF-08)

```bash
# Verify seed password is hashed, not plaintext
sqlite3 moneyapp.db "SELECT password_hash FROM users WHERE username='admin'"
# Expected: starts with '$2a$12$' or '$2b$12$' (bcrypt cost 12)
```

### Responsive layout checks (NF-16)

| Page | Viewport | Check |
|---|---|---|
| `/expenses` | 375×812 (mobile) | Table readable; pagination accessible; filter inputs usable; amounts not truncated. |
| `/invoices` | 768×1024 (tablet) | Status badges visible; "Mark as Paid" button accessible. |
| `/` (dashboard) | 375×812 (mobile) | All 3 summary cards visible (may stack vertically); invoice summary card visible. |

---

## Automation Backlog

Candidates for Playwright or Go integration test automation in a future pass. Not required for M1 sign-off.

### Playwright — login and auth flow (TS-01, TS-03)

| Test name | Suite | Priority | Notes |
|---|---|---|---|
| `should redirect to /login when unauthenticated` | TS-01 | High | `page.goto('/')` → assert URL ends with `/login` |
| `should login with valid credentials and reach dashboard` | TS-01 | High | Fill form, submit, assert `page.url()` is `/`, assert 3 summary cards visible |
| `should show error for invalid credentials` | TS-01 | High | Fill wrong password, assert alert/error text |
| `should clear token and redirect on logout` | TS-03 | High | Click logout button, assert URL `/login`, assert `localStorage.getItem('token')` is null |
| `should redirect to /login on expired token (401)` | TS-02 | Medium | Intercept API with `route.fulfill({ status: 401 })`, assert redirect to `/login` |

### Playwright — expense CRUD smoke (TS-05, TS-09, TS-10)

| Test name | Suite | Priority | Notes |
|---|---|---|---|
| `should create an expense and see it in the list` | TS-05 | High | Full form fill, submit, assert row in table |
| `should edit expense and see updated amount` | TS-09 | High | Click Edit, change amount, save, assert updated row |
| `should delete expense with confirmation` | TS-10 | High | Click Delete, confirm dialog, assert row removed |
| `should show empty state when no expenses exist` | TS-06 | Medium | Navigate with clean DB, assert empty state text |

### Playwright — invoice flow (TS-16, TS-18, TS-19)

| Test name | Suite | Priority | Notes |
|---|---|---|---|
| `should create invoice and see it under Unpaid tab` | TS-16 | High | Create via form; click Unpaid tab; assert row |
| `should mark invoice as paid via quick action` | TS-18 | High | Click "Mark as Paid", confirm dialog, assert green badge |
| `should show overdue invoices in red` | TS-19 | Medium | Requires past-due test data; assert red badge |
| `should filter invoices by status and persist in URL` | TS-17 | Medium | Click status tab; assert URL param; refresh; assert filter restored |

### Playwright — dashboard (TS-22, TS-23)

| Test name | Suite | Priority | Notes |
|---|---|---|---|
| `should show correct monthly totals on dashboard` | TS-22 | High | Seed data; navigate to `/`; assert card amounts |
| `should show invoice summary card on dashboard` | TS-23 | Medium | Seed invoices; assert summary count and amount |

### Go integration tests — service layer (TS-24)

```
Package: internal/services
  - TestExpenseService_Create_ValidAmount: in-memory DB; assert row created with correct amount
  - TestExpenseService_Create_ZeroAmount: assert error returned
  - TestExpenseService_Create_WrongCategoryType: income category for expense; assert error
  - TestIncomeService_Create_ZeroAmount: assert error
  - TestInvoiceService_Create_DueDateBeforeIssueDate: assert error
  - TestInvoiceService_UpdateOverdueStatuses_OnlyUnpaid: paid invoice not transitioned
  - TestInvoiceService_UpdateOverdueStatuses_PastDue: unpaid past-due → overdue
  - TestDashboardService_GetSummary_CurrentMonth: assert correct sums
  - TestAuthService_Login_ValidCredentials: returns token
  - TestAuthService_Login_WrongPassword: returns error
  - TestAuthService_ValidateToken_Expired: returns error
```

---

## Coverage Gaps & Notes

Issues and ambiguities identified while authoring this plan against the M1 ticket ACs:

| Gap | Ticket | Detail |
|---|---|---|
| **`CURRENCY` config and minor unit factor** | E2-S1, NF-10, NF-17 | Tickets specify amounts stored as "minor currency units" but VND has no subdivisions (1 VND = 1 unit). The frontend conversion (multiply by 100) may be incorrect for VND. The spec does not clarify whether `CURRENCY=VND` means the stored amount equals the display amount. Test TS-05-02 should be run against the actual `CURRENCY` config to confirm the factor. |
| **`/api/invoices/stats` vs. list `total_amount`** | E4-S8 | The ticket offers two implementation options: a dedicated `/api/invoices/stats` endpoint OR `total_amount` in the list response when filtered by status. Only one must be implemented; TS-21-01 tests the dedicated endpoint path. If the team chose the list-response path, TS-21-01 needs adjustment. |
| **E6-S6 endpoint shape** | E6-S6 | The ticket says to extend `GET /api/dashboard/summary` OR add `GET /api/dashboard/invoices`. The test plan (TS-23) tests via the dashboard page UI rather than a specific endpoint to remain implementation-agnostic. The tester should verify which endpoint the frontend calls. |
| **`UpdateOverdueStatuses` call during `List` vs. separate cron** | E4-S4 | The ticket says `UpdateOverdueStatuses` is called at the start of every `InvoiceService.List()`. This means concurrent reads could trigger unnecessary DB writes. TS-19 covers the functional outcome; teams should also monitor for write contention under load. |
| **Logout token invalidation** | E1-S3 | E1-S3 explicitly notes JWTs are stateless and there is no server-side token blacklist in MVP. TS-03-03 only verifies the API returns 200. A test for "token still usable after logout" is not included since that behavior is expected by design. This is a known security limitation to revisit in M2+. |
| **bcrypt cost factor for seed hash** | E1-S1 | Migration `002_seed_default_user.up.sql` contains a pre-computed bcrypt hash. If `BCRYPT_COST` in the environment differs from the cost used when generating the hash, login will still work (bcrypt encodes the cost in the hash), but the config value is misleading. Document which cost the seed hash was generated at. |
| **Dashboard date range defaults** | E6-S1 | The ticket says "default to current month if no dates provided" but does not specify whether the backend computes the current month or the frontend always sends explicit dates. TS-22-03 covers both: the API with no params should return current month data. |
| **Confirm dialog message text not specified for Income and Invoice delete** | E3-S4, E4-S6 | Ticket specifies the exact dialog message only for E2-S6 ("Are you sure you want to delete this expense?"). Income and invoice delete dialogs should use analogous messages. Tests TS-15 and TS-20 check for confirmation dialog presence but do not assert specific message text; tester should note if the message is missing or generic. |

---

*End of Milestone 1 Test Plan — Draft*

---

## M1 verification run (implementation)

**Date**: 2026-04-26  
**Commit**: `d21fe23`  
**Environment**: `m1-local` — `STORAGE_TYPE=local`, fresh DB (all migrations applied on startup), `BCRYPT_COST=12`, `TOKEN_EXPIRY_HOURS=24`, `CURRENCY=VND`  
**Runner**: qa-ux-tester agent (curl smoke + `go test` + `npm run lint && npm run build`)

### Automated checks

| Check | Result | Detail |
|---|---|---|
| `go test ./...` | **PASS** | 55/55 service tests pass; `internal/services` coverage **78.1%** (≥ 70% AC met) |
| `npm run lint` | **PASS** | No ESLint errors |
| `npm run build` | **PASS** | `tsc -b` clean; Vite bundle 266 kB JS / 11 kB CSS |

> Note: the backend binary running on port 8080 at test start was built before the latest commit (started at 16:51, code last modified at 19:27). It was replaced by a fresh `go run ./cmd/server` before smoke tests began. Always restart the server when verifying a new build.

---

### Summary table

| Area / Ticket | Status | Notes |
|---|---|---|
| **E1-S1 — Login** | **PASS** | Valid → 200 + JWT; invalid → 401; unauth API call → 401 |
| **E1-S2 — Token expiry** | PASS (partial) | JWT `exp` claim present; crafted expired token → 401. Browser redirect (TS-02-03) not tested — requires browser. |
| **E1-S3 — Logout** | PASS (partial) | `POST /api/auth/logout` → 200. UI redirect (TS-03-01/02) blocked (browser). |
| **E5-S1 — Default categories** | **PASS** | 9 expense + 6 income default categories seeded; `GET /api/categories?type=expense` → 200 with 9 items; unauth → 401 |
| **E2-S1 — Create expense** | **FAIL** | TS-05-01/03/04/06 pass. **TS-05-05 FAIL**: missing `date` field returns 400 `{"error":"date is required"}` — AC3 says default to today, not reject. |
| **E2-S2 — List expenses** | PASS (partial) | Page 1 = 20 items, `total_amount` present, `category_name` populated. Performance + browser loading state blocked. |
| **E2-S3 — Date range filter** | PASS (partial) | January filter correct; inverted range → 400. **TS-07-05 FAIL**: invalid date format `26-01-01` → 200 empty results, expected 400. URL persistence blocked (browser). |
| **E2-S4 — Category filter** | PASS (partial) | Single-category filter correct. Multi-category and URL persistence blocked (browser). |
| **E2-S5 — Edit expense** | PASS (partial) | Save → 200, amount updated, non-existent → 404. Form pre-populated (TS-09-01) blocked (browser). |
| **E2-S6 — Delete expense** | PASS (partial) | 204 on delete, 404 on non-existent. Confirm dialog / cancel UI (TS-10-02/03) blocked (browser). |
| **E2-S8 — Running total** | **PASS** | `total_amount` present in list response; value reflects filtered sum |
| **E3-S1 — Create income** | **PASS** | Create → 201; zero amount → 400; wrong category type → 400 (`category is not of type "income"`); `total_amount` in list |
| **E3-S2 — List income** | PASS (partial) | Pagination structure matches expense list. Empty state + performance blocked (browser/perf). |
| **E3-S3 — Filter income by date** | PASS (partial) | Date filter logic tested via service tests. Browser URL persistence blocked. |
| **E3-S4 — Edit and delete income** | **PASS** | Edit → 200; delete → 204; non-existent PUT → 404 |
| **E4-S1 — Create invoice** | **PASS** | Create → 201; missing vendor → 400; zero amount → 400; due_date < issue_date → 400; equal dates → 201; amount stored as integer |
| **E4-S2 — List invoices / status filter** | PASS (partial) | `status=unpaid` filter correct. Badge colors, URL persistence, performance blocked (browser). |
| **E4-S3 — Mark invoice as paid** | **PASS** | PATCH → 200, status=paid; already-paid → error; invalid status target → 400. Confirmation dialog (TS-18-03) blocked (browser). |
| **E4-S4 — Overdue highlighting** | **PASS** | Past-due unpaid → `overdue` status in API and DB (`SELECT status FROM invoices WHERE id=3` → `overdue`). Paid invoice not transitioned. |
| **E4-S6 — Edit and delete invoice** | **FAIL** | Delete → 204; non-existent GET/DELETE → 404. **TS-20-01 FAIL**: `PUT /api/invoices/:id` without `status` field → 400 — `status` is validated as required in `InvoiceService.Update()` but is not logically required for a general edit. |
| **E4-S8 — Outstanding invoice total** | **PASS** | `GET /api/invoices/stats` → 200 with `total_outstanding`, `unpaid_count`, `overdue_count`; routing correctly hits stats not `:id` |
| **E6-S1 — Dashboard summary** | PASS (partial) | `total_income`, `total_expenses`, `net_balance` correct; defaults to current month. **Minor defect**: `unpaid_amount` and `overdue_amount` are hardcoded 0. Three summary card UI blocked (browser). |
| **E6-S6 — Invoice summary on dashboard** | Blocked | Dashboard API includes `unpaid_count`, `overdue_count`, `total_outstanding`; frontend card rendering blocked (browser). |
| **M1-01 — Service layer unit tests** | **PASS** | 78.1% coverage; zero-amount, date-ordering, overdue-only-unpaid tests present and passing |
| **Cross-cut: Auth on all routes** | **PASS** | All `/api/expenses`, `/api/incomes`, `/api/invoices`, `/api/categories`, `/api/dashboard/summary`, `/api/invoices/stats` return 401 without token; `/api/health` → 200 |
| **Cross-cut: NF-10 integers** | **PASS** | No float notation in any API response |
| **Cross-cut: NF-08 bcrypt** | **PASS** | Seed hash `$2a$12$...` — cost factor 12 confirmed |
| **Cross-cut: NF-09 error format** | **PASS** | All error responses are `{"error":"..."}`, no stack traces, SQL, or Go type names |
| **Cross-cut: Pagination cap** | **FAIL** | `GET /api/expenses?per_page=9999` returns `per_page: 9999` and all items — cap at 100 not enforced (applies to expenses, incomes, and invoices handlers) |

---

### Defects found

| # | Severity | Area | Description | Location |
|---|---|---|---|---|
| D-01 | **Major** | E2-S1 (TS-05-05) | `POST /api/expenses` with no `date` field returns HTTP 400 `{"error":"date is required"}`. AC3 specifies the date should default to today. | `backend/internal/services/expense_service.go` `Create()` — `date` is required but AC expects empty = today |
| D-02 | **Minor** | E2-S3 (TS-07-05) | Invalid date format (`26-01-01`) in `date_from`/`date_to` params silently returns HTTP 200 with 0 results instead of HTTP 400 validation error. Same issue exists for incomes. | `backend/internal/handlers/expense_handler.go` `handleList()` — no date format validation on query params |
| D-03 | **Major** | E4-S6 (TS-20-01) | `PUT /api/invoices/:id` without a `status` field in the body returns HTTP 400. `InvoiceService.Update()` validates `status` as required, but a general field-level edit (vendor, amount, dates) should not need to re-specify status. The empty-string check `inv.Status != "unpaid" && inv.Status != "paid" && inv.Status != "overdue"` rejects the empty default. | `backend/internal/services/invoice_service.go:Update()` line ~`if inv.Status != ...` |
| D-04 | **Major** | Cross-cutting (NF-01) | `per_page` query parameter is not capped at 100. Sending `per_page=9999` returns all rows with no limit enforcement. Applies to `/api/expenses`, `/api/incomes`, `/api/invoices`. | `backend/internal/handlers/expense_handler.go`, `income_handler.go`, `invoice_handler.go` — no max per_page check |
| D-05 | **Minor** | E6-S1 | `GET /api/dashboard/summary` returns `unpaid_amount: 0` and `overdue_amount: 0` unconditionally. The `DashboardService.GetSummary()` calls `invoiceService.GetStats()` which doesn't populate per-status amounts; those fields are hardcoded to `0` in the return struct. | `backend/internal/services/dashboard_service.go:GetSummary()` — `UnpaidAmount: 0, OverdueAmount: 0` |

---

### Blocked / not tested

The following test cases require a live browser and were not executed in this run:

- **TS-01-03/04/05/06** — Client-side form validation and Auth header injection (DevTools)
- **TS-02-03/04** — Expired token browser redirect; `exp−iat` verification
- **TS-03-01/02** — Logout UI button, Back-button after logout
- **TS-06-02/03/04/05** — Page 2 navigation, empty state UI, loading spinner, 10 k-record performance
- **TS-07-02/03**, **TS-08-02/03/04** — Clear filter, URL persistence on refresh
- **TS-09-01/03** — Edit form pre-populated, client-side invalid amount
- **TS-10-02/03** — Cancel delete, confirm dialog message text
- **TS-13-03/04**, **TS-14-01/02/03** — Income empty state, performance; income date filter UI
- **TS-15-03** — Cancel delete income
- **TS-17-02/03/04/05** — Invoice list (all), URL persistence, status badge colors, performance
- **TS-18-02/03** — "Mark as Paid" hidden on paid rows, confirmation dialog
- **TS-19-02/04** — Future-due invoice remains unpaid UI; startup overdue update log check
- **TS-20-03** — Cancel delete invoice
- **TS-21-02/03** — Outstanding total shown in UI; total updates after paid
- **TS-22-02/04/05** — Three summary card UI; zero-record state; loading skeleton
- **TS-23-01/02/03** — Invoice summary dashboard card

---

### Residual risk / CI note

- **CGO**: `go test ./...` and `go build` require CGO for `mattn/go-sqlite3`. The `.github/workflows/ci.yml` backend jobs must include `CGO_ENABLED=1` and a C toolchain (e.g. `sudo apt-get install gcc` on Ubuntu runners). If CI runs without CGO, the build will silently produce an error or a broken binary.
- **Stale binary risk**: The backend binary was stale at test start (built 2.5 h before test run). CI should build fresh from source; local devs should not rely on a running `go run` from a previous session.
- **D-01 / D-03** are most likely to surface as user-facing friction: expense creation without a date fails silently from the frontend perspective, and invoice edits require callers to always re-send the current status.
- **D-04** (no per_page cap) is a latent performance risk — a malicious or buggy client could request the entire table in one call.

---

**Counts: Pass 43 / Fail 5 defects (spanning 8 test case IDs) / Blocked ~40 (browser/UI/perf)**  
**lint+build+tests: GREEN**

---

### Browser MCP pass

**Date**: 2026-04-26  
**Environment**: `m1-local` — backend `http://localhost:8080`, frontend `http://localhost:5173` (Vite dev server)  
**Runner**: qa-ux-tester agent — Cursor IDE Browser MCP (`cursor-ide-browser`)  
**Browser View ID**: `d49870`

#### URLs tested

- `http://localhost:5173/login`
- `http://localhost:5173/dashboard`
- `http://localhost:5173/expenses` (filtered: `?category_id=1`)
- `http://localhost:5173/income`
- `http://localhost:5173/invoices` (filtered: `?status=unpaid`, `?status=overdue`)

#### Pass / Fail / Blocked summary

| Area | Status | What was verified |
|---|---|---|
| **E1-S1 — Login (valid credentials)** | **PASS** | `admin`/`changeme` → redirected to `/dashboard`; "Signing in…" disabled-button state visible |
| **E1-S1 — Login (invalid credentials)** | **FAIL** | Wrong password clears form and stays on `/login` but **no error message is displayed** — user gets no feedback (defect B-01) |
| **E1-S3 — Logout** | **PASS** | Logout button → `/login`; subsequent navigation to `/dashboard` redirects back to `/login` |
| **Route protection** | **PASS** | Navigating to `/dashboard` unauthenticated redirects to `/login` |
| **E6-S1 — Dashboard summary cards** | **PASS** | Three cards visible: Total Income 27.000.000 ₫, Total Expenses 99.999 ₫, Net Balance 26.900.001 ₫; current month period shown |
| **E6-S6 — Invoice summary on dashboard** | **PASS** | "Outstanding Invoices" card shows 8.000.000 ₫, "2 outstanding (1 unpaid, 1 overdue)"; "View invoices →" link present |
| **E5-S1 — Default categories** | **PASS** | Category dropdown in Add Expense form shows 9 categories (Food, Transport, Housing, Health, Entertainment, Shopping, Utilities, Other, Uncategorized) |
| **E2-S1 — Add Expense form opens** | **PASS** | Modal opens; date pre-filled with today; categories load (slight async delay); Amount validation shows "Amount must be greater than zero" on empty submit |
| **E2-S1 — Add Expense submit (via API)** | **PASS** | API `POST /api/expenses` → 201; new record immediately visible at top of list on reload |
| **E2-S2 — List expenses** | **PASS** | List loads with DATE / CATEGORY / DESCRIPTION / AMOUNT / ACTIONS columns; pagination controls present (Previous disabled on page 1, Next active) |
| **E2-S3 — Filter by date** | Blocked | Date inputs present (FROM/TO) but `<input type="date">` could not be filled via Browser MCP (React controlled input — see tooling notes) |
| **E2-S4 — Filter by category** | **PASS** | Category combobox → "Food" → URL updates to `?category_id=1`; list shows only Food expenses; total updates (924.999 ₫ → 649.999 ₫) |
| **E2-S5 — Edit expense** | **PASS** | Edit modal opens with pre-populated amount, category, description; "Saving…" button shown on submit; modal closes and list refreshes |
| **E2-S6 — Delete expense** | **PASS** | Confirmation dialog ("Delete Expense / Are you sure?"); confirm deletes; record removed from list; total updated |
| **E2-S8 — Running total** | **PASS** | Total updates on filter change and on create/delete |
| **E3-S2 — Income list** | **PASS** | 4 income records; total 27.000.000 ₫ (matches dashboard); filter controls visible |
| **E4-S2 — Invoices status filter** | **PASS** | Tabs: All / Unpaid / Paid / Overdue; URL updates to `?status=unpaid`, `?status=overdue`; list filtered correctly |
| **E4-S3 — Mark invoice as paid** | **PASS** | Confirmation dialog shown; after confirm: Unpaid count 1→0, Outstanding 8.000.000→3.000.000 ₫; "No invoices found." shown in Unpaid tab |
| **E4-S4 — Overdue highlighting** | **PASS** | Overdue invoice ("Old Vendor", due 2020-01-31) shows red "⚠ OVERDUE" badge; "✓ Paid" button available |
| **E4-S8 — Outstanding total** | **PASS** | Invoices page header shows OUTSTANDING, UNPAID, OVERDUE counts; updates after mark-as-paid |

#### Defects found in browser run

| # | Severity | Area | Description | Steps to reproduce |
|---|---|---|---|---|
| B-01 | **Major** | E1-S1 | **No error message on failed login** — wrong credentials silently clears the form; user has no feedback. | 1. Enter `admin` / `wrongpassword`. 2. Click Sign In. 3. Observe: form resets to empty with no error text. |
| B-02 | **Major** | E2, E3, E4 | **Date column shows raw ISO datetime** — all list views (expenses, income, invoices) show `2026-04-26T00:00:00Z` instead of a formatted date like `26/04/2026`. | Open `/expenses`; observe DATE column values. |
| B-03 | **Major** | E2-S5, E3-S4, E4-S6 | **Edit form date field not pre-populated** — the edit modal passes the ISO datetime string `2026-04-26T00:00:00Z` to `<input type="date">` which expects `YYYY-MM-DD`; the field shows `dd/mm/yyyy` placeholder (empty). Users cannot see or edit the original date. | Open any expense edit modal; observe the Date field is blank. |

#### Tooling notes (Cursor IDE Browser MCP)

- **Ref staleness with React controlled inputs**: The browser MCP assigns integer refs (`e0`, `e1`, …) per DOM snapshot. React re-renders between `browser_snapshot` and `browser_fill`/`browser_type` calls invalidate these refs. This affected all `<input type="number">` and `<textarea>` elements in the expense/income forms (amount field could not be filled via the MCP). Workaround used: API-side creation verified by direct `curl`, then reloaded page in browser to confirm list update.
- **`browser_press_key` does not trigger React onChange**: Sending digit key events to a focused `<input type="number">` produced no visible state update. React controlled inputs require a synthetic `input` event, which `browser_press_key` does not generate.
- **`browser_select_option` works for native `<select>`**: The category combobox and invoice status tabs both responded correctly to `browser_select_option`/`browser_click`.
- **`browser_mouse_click_xy` is the reliable fallback for focused fields**: After clicking a `<button>` or `<select>` with coordinates from a fresh screenshot, the click lands correctly. Modal open/close, pagination, tab switching, and confirmation dialogs all worked via this method.
- **Screenshot vs snapshot timing**: The first `browser_navigate` screenshot may show the previous page content (React routing async). Always call `browser_snapshot` after navigation to confirm the correct URL before interacting.

#### Browser MCP pass counts

**19 checks executed | 16 PASS | 2 FAIL | 1 Blocked**

- PASS: Login (valid), Logout, Route protection, Dashboard cards, Invoice summary card, Default categories, Add Expense form validation, Add Expense (API-verified), List Expenses, Category filter, Edit Expense, Delete Expense, Running total, Income list, Invoice status filter, Mark as Paid, Overdue highlighting, Outstanding total
- FAIL: Login error message missing (B-01); previously blocked TS-03-01/02 now PASS
- Blocked: Date range filter inputs (React number/date input limitation with Browser MCP)
