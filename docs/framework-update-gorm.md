# Framework Update: Migrating to GORM

**Date:** 2026-04-26
**Status:** Draft
**Audience:** Backend developers, QA, Product

---

## 1. Executive Summary

MoneyApp's backend will replace the current `database/sql` + raw SQL approach with **GORM** — the most widely used ORM for Go. GORM sits on top of the same `mattn/go-sqlite3` driver already in use, so the database file, WAL mode, and connection semantics stay identical.

**What is changing:**

- The `*sql.DB` handle passed into every service constructor becomes a `*gorm.DB` handle.
- Hand-written SQL strings inside `internal/services/` are replaced with GORM's chainable query API.
- The `internal/database/db.go` `Open()` function is rewritten to open via GORM's SQLite dialector instead of `database/sql` directly.
- GORM's `AutoMigrate` is evaluated for struct-driven schema validation; the existing numbered `.sql` migration files remain the authoritative schema source (see §4 — Developer Workflow).
- Test helpers in `services_test` package are updated to use an in-memory GORM connection rather than a raw `*sql.DB`.

**What is NOT changing:**

- The app remains single-user and self-hosted.
- The database engine remains SQLite.
- All monetary amounts remain `int64` (minor currency units / cents) — GORM does not force floats.
- The REST API surface (routes, request/response shapes, status codes) is unchanged; handlers are unaffected.
- The `migrations/` SQL files and the `RunMigrations` runner are retained as the source of truth for schema versioning. GORM's `AutoMigrate` is used only as a safety check, not as the primary migration mechanism.
- MinIO / local storage integration is unaffected.

---

## 2. Current vs Target Architecture

### 2.1 Current (database/sql + raw SQL)

```
HTTP Request
    ↓
Handler  (internal/handlers/*.go)
    ↓  calls method on
Service  (internal/services/*.go)
    │  holds *sql.DB
    ↓  executes string SQL via
database/sql  (ExecContext / QueryRowContext / QueryContext)
    ↓
mattn/go-sqlite3  (CGO driver)
    ↓
SQLite file (moneyapp.db)
```

Each service carries a `*sql.DB` field. SQL statements are string literals scattered through methods such as `ExpenseService.Create`, `ExpenseService.List`, `ExpenseService.GetByID`, etc. Filtering logic (date range, category IDs, sorting) is assembled by helper functions (`buildExpenseWhere`, `buildExpenseOrder`) that concatenate SQL fragments and build positional `?` argument slices by hand.

Tests open an in-memory `*sql.DB`, apply DDL via a hard-coded list of `CREATE TABLE` statements in `test_helpers_test.go`, and pass the connection directly to service constructors.

### 2.2 Target (GORM + SQLite dialector)

```
HTTP Request
    ↓
Handler  (internal/handlers/*.go)  — unchanged
    ↓  calls method on
Service  (internal/services/*.go)
    │  holds *gorm.DB
    ↓  uses GORM query API
GORM core + SQLite dialector  (gorm.io/gorm + gorm.io/driver/sqlite)
    ↓
mattn/go-sqlite3  (same CGO driver)
    ↓
SQLite file (moneyapp.db)
```

Key changes at each layer:

| Layer | Before | After |
|---|---|---|
| `database.Open()` | `sql.Open("sqlite3", dsn)` | `gorm.Open(sqlite.Open(dsn), &gorm.Config{})` |
| Service struct fields | `db *sql.DB` | `db *gorm.DB` |
| Simple read (GetByID) | `QueryRowContext` + manual `Scan` | `db.First(&e, id)` |
| Filtered list | String concatenation helpers | `db.Where(...).Order(...).Limit(...).Offset(...).Find(&expenses)` |
| Insert | `ExecContext` + `LastInsertId` | `db.Create(&e)` — GORM populates `e.ID` |
| Update | `ExecContext` + `RowsAffected` | `db.Save(&e)` or `db.Model(&e).Updates(...)` |
| Delete | `ExecContext` + `RowsAffected` | `db.Delete(&e, id)` |
| Transactions (NF-11) | `db.Begin()` / `tx.Commit()` | `db.Transaction(func(tx *gorm.DB) error { ... })` |
| Test setup | Raw DDL slice in `test_helpers_test.go` | `gorm.Open(sqlite.Open(":memory:"), ...)` + `AutoMigrate` on model structs |
| Health check | `db.PingContext(ctx)` | `sqlDB, _ := db.DB(); sqlDB.PingContext(ctx)` |

---

## 3. Impacted Areas

### Files requiring direct changes

| File / Package | Nature of Change |
|---|---|
| `backend/go.mod` | Add `gorm.io/gorm` and `gorm.io/driver/sqlite`; remove or downgrade direct `database/sql` import if no longer needed at top level |
| `backend/internal/database/db.go` | Rewrite `Open()` to return `*gorm.DB` instead of `*sql.DB` |
| `backend/internal/database/migrate.go` | Update `RunMigrations` signature to accept `*gorm.DB`; extract `*sql.DB` via `db.DB()` for the migration executor (GORM wraps the underlying connection) |
| `backend/internal/services/expense_service.go` | Replace all `*sql.DB` usages, SQL strings, and scan loops |
| `backend/internal/services/income_service.go` | Same pattern as expense |
| `backend/internal/services/invoice_service.go` | Same pattern |
| `backend/internal/services/category_service.go` | Same pattern |
| `backend/internal/services/attachment_service.go` | Same pattern; transaction wrapping for multi-table deletes (NF-11) |
| `backend/internal/services/dashboard_service.go` | Aggregate queries — may use `db.Raw(...)` if GORM's API doesn't express the query cleanly |
| `backend/internal/services/auth_service.go` | Update DB field type |
| `backend/internal/services/test_helpers_test.go` | Replace `setupTestDB` with a GORM in-memory helper |
| `backend/internal/handlers/health.go` | Update `db *sql.DB` field; extract underlying `*sql.DB` via `gormDB.DB()` for `PingContext` |
| `backend/cmd/server/main.go` | Update wiring — `database.Open` now returns `*gorm.DB`, pass it to service constructors |

### Files not expected to change

- All handler files (`expense_handler.go`, etc.) — they call service methods; those signatures stay the same.
- Storage layer (`internal/storage/`) — orthogonal to database access.
- Frontend (`frontend/`) — REST API surface is frozen.
- Migration SQL files (`backend/migrations/*.sql`) — remain the authoritative schema.

---

## 4. Developer Workflow Changes

### 4.1 New dependencies

Run from `backend/`:

```bash
go get gorm.io/gorm
go get gorm.io/driver/sqlite
```

`gorm.io/driver/sqlite` still depends on `mattn/go-sqlite3`, so **CGO remains required**. The existing `CGO_ENABLED=1` requirement documented in `CLAUDE.md` does not change.

### 4.2 Migrations — no change in day-to-day flow

The numbered `.sql` files in `backend/migrations/` and the `RunMigrations` runner continue to drive schema evolution. When you need a new column or table:

1. Create `backend/migrations/NNN_description.up.sql` with the DDL.
2. Server startup applies it automatically on next run.

GORM `AutoMigrate` is used in tests for convenience (the test helper calls `db.AutoMigrate(&models.Expense{}, ...)` to create an in-memory schema from struct tags) but is **not** used against the production SQLite file.

### 4.3 Writing new queries

**Simple lookups — prefer GORM chainable API:**

```go
// Find one by primary key; populates e, returns gorm.ErrRecordNotFound if missing
var e models.Expense
result := s.db.WithContext(ctx).First(&e, id)
if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, ErrNotFound
}
```

**Filtered lists — chain conditions:**

```go
query := s.db.WithContext(ctx).
    Model(&models.Expense{}).
    Joins("LEFT JOIN categories ON categories.id = expenses.category_id").
    Where("expenses.date >= ?", params.DateFrom).
    Order("expenses.date DESC").
    Limit(params.PerPage).
    Offset((params.Page - 1) * params.PerPage)
query.Find(&expenses)
```

**Complex aggregates — drop to raw SQL when GORM becomes verbose:**

```go
s.db.WithContext(ctx).Raw(
    "SELECT COUNT(*), COALESCE(SUM(amount),0) FROM expenses WHERE ...",
    args...,
).Scan(&result)
```

This is a valid and idiomatic pattern; GORM does not require you to replace every query.

**Transactions (NF-11):**

```go
err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Delete(&models.Attachment{}, "entity_type = ? AND entity_id = ?", "expense", id).Error; err != nil {
        return err
    }
    return tx.Delete(&models.Expense{}, id).Error
})
```

**Recommended GORM references:**

- Official docs: [https://gorm.io/docs/](https://gorm.io/docs/)
- SQLite dialector: [https://gorm.io/docs/connecting_to_the_database.html#SQLite](https://gorm.io/docs/connecting_to_the_database.html#SQLite)
- GORM model conventions: [https://gorm.io/docs/models.html](https://gorm.io/docs/models.html)

### 4.4 Model struct tags (project convention)

GORM reads struct tags to map fields to columns. When column names match Go field names in snake_case, tags are optional. For non-obvious mappings, use explicit tags:

```go
// models/expense.go (illustrative — align with existing struct fields)
type Expense struct {
    ID          int64  `gorm:"primaryKey;autoIncrement"`
    Amount      int64  `gorm:"not null"`     // always minor units, never float
    Date        string `gorm:"type:date;not null"`
    CategoryID  int64  `gorm:"not null"`
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    // joined read-only field — not a DB column:
    CategoryName string `gorm:"-"`
}
```

Keep `CategoryName` (and similar computed fields) tagged `gorm:"-"` so GORM does not attempt to write them.

### 4.5 Running tests

No command change. Tests continue to use the in-memory SQLite dialect:

```bash
cd backend && go test ./...
```

The `setupTestDB` helper in `test_helpers_test.go` is rewritten to open `gorm.Open(sqlite.Open(":memory:"), ...)` and call `db.AutoMigrate` on all relevant model structs, replacing the hard-coded DDL slice.

---

## 5. Operations / CI

### 5.1 Build and run

No change to the `make dev-backend` shortcut or `go run ./cmd/server`. CGO is already expected.

### 5.2 Docker

No changes expected. The existing image build strategy (if any) already requires a C toolchain for `mattn/go-sqlite3`. `gorm.io/driver/sqlite` uses the same driver.

### 5.3 CI pipeline

- `go test ./...` continues to cover service-level tests with an in-memory DB.
- No new environment variables are needed.
- If the CI image does not have a C toolchain installed, it needs one. This is pre-existing: the raw `go-sqlite3` driver already required CGO. Flag this with your CI maintainer if tests currently pass with `CGO_ENABLED=0` (which would only be possible if tests are not importing the SQLite driver, a state that should be corrected regardless of this migration).

---

## 6. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| **Query regression** — GORM generates subtly different SQL than hand-written queries (e.g., extra columns in SELECT, different NULL handling) | Medium | Medium | Run the full integration test suite before and after; add golden-path tests for any query that uses `COALESCE` or multi-table joins |
| **`updated_at` management conflict** — GORM auto-updates `UpdatedAt` on every `Save`, but the current raw SQL sets `updated_at = CURRENT_TIMESTAMP` explicitly. Both approaches work but the column type must be consistent | Medium | Low | Align `UpdatedAt` field type to `time.Time` in model structs; verify existing rows (stored as text) parse correctly on first read |
| **Transaction scope for NF-11** — The current `attachment_service.go` deletes storage objects and DB rows in sequence, not in a DB transaction; NF-11 requires a transaction around multi-table DB writes | High (pre-existing gap) | Medium | Wrap multi-table DB operations in `db.Transaction(...)` as part of this migration; document that storage-object deletion remains outside the transaction (storage is not transactional) |
| **GORM `ErrRecordNotFound` vs `sql.ErrNoRows`** — Any code checking `errors.Is(err, sql.ErrNoRows)` will silently fail after migration | High | High | Grep for `sql.ErrNoRows` across the codebase and replace with `gorm.ErrRecordNotFound` before or during service rewrites |
| **Learning curve** — Developers unfamiliar with GORM may reach for `Raw` too eagerly or misuse `Save` (which updates all fields) vs `Updates` (which updates only non-zero fields) | Low–Medium | Low | Establish a brief convention note (see §4.3) and add to code review checklist |
| **Performance of complex dashboard queries** — GORM adds overhead and generates verbose SQL for aggregations | Low | Low | Dashboard queries may remain as `db.Raw(...)` — this is idiomatic and carries no penalty |
| **Migration drift** — Future developer uses `AutoMigrate` in production instead of the numbered SQL files | Low | Medium | Document explicitly (§4.2) that `AutoMigrate` is test-only; enforce via code review |

---

## 7. Definition of Done

QA and PM can use this checklist to verify the migration is complete and safe to merge.

### Code

- [ ] `backend/go.mod` and `backend/go.sum` include `gorm.io/gorm` and `gorm.io/driver/sqlite`
- [ ] `database.Open()` returns `*gorm.DB`; no `*sql.DB` is returned from the `database` package
- [ ] All service structs hold `*gorm.DB`, not `*sql.DB`
- [ ] No raw SQL string literals remain in `internal/services/` except inside explicit `db.Raw(...)` or `db.Exec(...)` calls that are documented with a comment explaining why the GORM API was insufficient
- [ ] `errors.Is(err, sql.ErrNoRows)` does not appear anywhere in `internal/services/`
- [ ] `handlers/health.go` extracts `*sql.DB` via `gormDB.DB()` to call `PingContext`; health endpoint returns correct status
- [ ] `test_helpers_test.go` uses `gorm.Open(sqlite.Open(":memory:"), ...)` — the hard-coded DDL slice is removed
- [ ] Model structs have `gorm:"-"` on all computed/joined fields not stored in the DB

### Tests

- [ ] `go test ./...` passes with no failures
- [ ] No test skips that were not present before the migration
- [ ] At minimum one test per service covers: Create, GetByID (found), GetByID (not found → `ErrNotFound`), Update, Delete, List with filters

### Transactions (NF-11)

- [ ] Every code path that writes to more than one table is wrapped in `db.Transaction(...)`
- [ ] There is a test that verifies a mid-transaction failure does not leave orphan rows

### Functional verification

- [ ] `GET /api/health` returns `{"status":"ok","database":"ok",...}` against a real SQLite file
- [ ] All expense, income, invoice, and category CRUD endpoints return the same response shapes as before
- [ ] Existing production SQLite database (if applicable) opens without error after the update

### Documentation

- [ ] `CLAUDE.md` (or an inline comment in `db.go`) is updated to note the GORM dependency
- [ ] This document status is updated from **Draft** to **Approved** once reviewed by at least one backend developer and the PM

---

## 8. Glossary

| Term | Definition |
|---|---|
| **ORM** (Object-Relational Mapper) | A library that maps Go structs to database tables, generating SQL automatically. You work with Go types; the ORM translates to/from SQL rows. |
| **Dialector** | GORM's adapter for a specific database engine. `gorm.io/driver/sqlite` is the SQLite dialector; it wraps `mattn/go-sqlite3` and translates GORM's generic operations to SQLite-compatible SQL. |
| **Migrator** | GORM's built-in tool (accessed via `db.Migrator()`) that can inspect and modify schema based on struct definitions. In this project it is used only in tests via `AutoMigrate`; numbered SQL files remain the production migration mechanism. |
| **Model tags** | Go struct field annotations (e.g., `` `gorm:"primaryKey"` ``) that tell GORM how to map a field: column name, constraints, whether to include it in INSERT/UPDATE, etc. A tag of `gorm:"-"` tells GORM to ignore the field entirely. |
| **`db.Transaction(...)`** | GORM helper that begins a transaction, passes a scoped `*gorm.DB` to the provided function, and automatically commits on success or rolls back on returned error. Replaces the manual `db.Begin()` / `tx.Commit()` / `tx.Rollback()` pattern. |
| **`gorm.ErrRecordNotFound`** | GORM's equivalent of `database/sql`'s `sql.ErrNoRows`. Returned by `First`, `Take`, and `Last` when no matching row exists. Must replace all `sql.ErrNoRows` checks in the codebase. |
| **WAL mode** | SQLite's Write-Ahead Logging journal mode. Allows concurrent reads with a single writer. Retained via the DSN parameter `?_journal_mode=WAL` passed to the GORM dialector's `Open` call — behavior is unchanged. |
