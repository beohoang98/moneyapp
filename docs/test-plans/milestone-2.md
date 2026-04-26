# Test Plan — Milestone 2: File Attachments & Enhanced Lists

**Milestone goal**: Add file upload/download via configurable storage backend (local or S3-compatible), inline previews, custom categories, and enhanced sorting/filtering. Auto-overdue for invoices via an admin endpoint.  
**Exit criteria (from tickets.md)**: User can attach files to any record, preview images inline, download files, manage custom categories, and invoices auto-transition to overdue status.  
**NF references covered**: FA-01, FA-02, FA-03, FA-04, FA-05, FA-06, FA-07, FA-08, NF-06, NF-07, NF-09, NF-10, NF-11, NF-12, NF-14, NF-15, NF-16, NF-17, NF-20  
**Date**: 2026-04-26  
**Status**: Draft

---

## Table of Contents

1. [Environment Setup](#environment-setup)
2. [Traceability Matrix](#traceability-matrix)
3. [E7-S1 — File Upload Infrastructure](#e7-s1--file-upload-infrastructure)
4. [E7-S2 — File Size Validation](#e7-s2--file-size-validation)
5. [E7-S3 — File Download](#e7-s3--file-download)
6. [E7-S4 — Image Thumbnail Preview](#e7-s4--image-thumbnail-preview)
7. [E7-S5 — Cascade Delete Attachments](#e7-s5--cascade-delete-attachments)
8. [E2-S7 — Attach Receipts to Expenses](#e2-s7--attach-receipts-to-expenses)
9. [E3-S5 — Attach Documents to Income](#e3-s5--attach-documents-to-income)
10. [E4-S5 — Attach Invoice PDFs](#e4-s5--attach-invoice-pdfs)
11. [E4-S7 — Overdue Check Admin Endpoint](#e4-s7--overdue-check-admin-endpoint)
12. [E5-S2 — Create Custom Categories](#e5-s2--create-custom-categories)
13. [E5-S3 — Rename and Delete Custom Categories](#e5-s3--rename-and-delete-custom-categories)
14. [E5-S4 — Reassign Transactions on Category Delete](#e5-s4--reassign-transactions-on-category-delete)
15. [M2-01 — Expense List Sorting](#m2-01--expense-list-sorting)
16. [M2-02 — Income Filter by Category](#m2-02--income-filter-by-category)
17. [M2-03 — Invoice Date Range Filter](#m2-03--invoice-date-range-filter)
18. [M2-04 — Income Running Total](#m2-04--income-running-total)
19. [Cross-Cutting Checks](#cross-cutting-checks)
20. [Automation Backlog](#automation-backlog)
21. [Coverage Gaps & Notes](#coverage-gaps--notes)

---

## Environment Setup

### Required environments

| Environment | Description | Notes |
|---|---|---|
| **m2-local** | `.env` with `STORAGE_TYPE=local`, `LOCAL_STORAGE_PATH=./data/storage`; no MinIO container | Primary environment for all M2 file attachment tests — fastest to spin up |
| **m2-minio** | `.env` with `STORAGE_TYPE=s3` and MinIO running via `docker compose up minio -d` | Required for S3-path attachment tests; use the same test cases as m2-local, run in both |
| **m2-seeded** | M2-local with M1 seed data (admin/changeme, default categories) and at least 3 expense/income/invoice records | Used for entity-specific attachment tests (E2-S7, E3-S5, E4-S5, E7-S5) |
| **m2-fresh** | Fresh checkout, no existing DB; migrations 001–008 applied on first start | Verifies migration 008 creates the `attachments` table on a clean install |
| **m2-storage-down** | `STORAGE_TYPE=local` with `LOCAL_STORAGE_PATH` pointing to a non-writable directory | Tests NF-12: graceful degradation when storage is unavailable |

### Prerequisites

#### Option A — Minimal (`STORAGE_TYPE=local`, no MinIO)

```bash
# Copy env; set local storage variables
cp .env.example .env
# Set / confirm:
#   STORAGE_TYPE=local
#   LOCAL_STORAGE_PATH=./data/storage
#   JWT_SECRET=test-secret-m2
#   TOKEN_EXPIRY_HOURS=24
#   BCRYPT_COST=12
#   CURRENCY=VND
#   PORT=8080
#   DB_PATH=./moneyapp.db

# Start backend (migrations 001-008 applied automatically)
cd backend && go run ./cmd/server

# Start frontend
cd frontend && npm run dev
```

Verify `attachments` table exists after startup:

```bash
sqlite3 backend/moneyapp.db ".tables"
# Expected output includes: attachments
sqlite3 backend/moneyapp.db ".schema attachments"
# Expected: CREATE TABLE attachments (id INTEGER PRIMARY KEY AUTOINCREMENT, ...)
```

#### Option B — With MinIO (`STORAGE_TYPE=s3`)

```bash
# Start MinIO container
docker compose up minio -d

# Confirm MinIO is healthy
curl -s http://localhost:9000/minio/health/live
# Expected: empty 200 response

# Set env for s3 storage
#   STORAGE_TYPE=s3
#   MINIO_ENDPOINT=localhost:9000
#   MINIO_ACCESS_KEY=minioadmin
#   MINIO_SECRET_KEY=minioadmin
#   MINIO_BUCKET=moneyapp
#   MINIO_USE_SSL=false

cd backend && go run ./cmd/server
cd frontend && npm run dev
```

MinIO console: `http://localhost:9001` (user: `minioadmin`, pass: `minioadmin`). Use the console to visually inspect uploaded objects and verify deletions.

#### Test file fixtures

Prepare the following test files before running attachment tests:

```bash
# Small valid JPEG (~50KB)
curl -o /tmp/test_receipt.jpg "https://via.placeholder.com/800x600.jpg"

# Valid PDF (~100KB) — or create a minimal one:
python3 -c "
import struct, zlib
# Minimal valid 1-page PDF
data = b'%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n3 0 obj<</Type/Page/MediaBox[0 0 612 792]/Parent 2 0 R>>endobj\nxref\n0 4\n0000000000 65535 f\n0000000009 00000 n\n0000000058 00000 n\n0000000115 00000 n\ntrailer<</Size 4/Root 1 0 R>>\nstartxref\n190\n%%EOF'
open('/tmp/test_invoice.pdf', 'wb').write(data)
print('PDF written')
"

# Valid PNG (~20KB)
python3 -c "
import struct, zlib
def make_png():
    sig = b'\x89PNG\r\n\x1a\n'
    w, h = 100, 100
    ihdr = struct.pack('>IIBBBBB', w, h, 8, 2, 0, 0, 0)
    ihdr_chunk = b'IHDR' + ihdr
    crc1 = struct.pack('>I', zlib.crc32(ihdr_chunk) & 0xffffffff)
    raw = b'\x00' + b'\xff\x00\x00' * w
    idat = zlib.compress(raw * h)
    idat_chunk = b'IDAT' + idat
    crc2 = struct.pack('>I', zlib.crc32(idat_chunk) & 0xffffffff)
    iend_chunk = b'IEND'
    crc3 = struct.pack('>I', zlib.crc32(iend_chunk) & 0xffffffff)
    def chunk(name, data):
        return struct.pack('>I', len(data)) + name + data + struct.pack('>I', zlib.crc32(name+data) & 0xffffffff)
    return sig + chunk(b'IHDR', ihdr) + chunk(b'IDAT', idat) + chunk(b'IEND', b'')
open('/tmp/test_image.png', 'wb').write(make_png())
print('PNG written')
"

# Oversized file (11MB) for rejection tests
dd if=/dev/urandom of=/tmp/test_oversized.bin bs=1M count=11

# Invalid file type
echo "MZ fake exe" > /tmp/test_malicious.exe
```

### Environment variables under test (M2-specific)

```
# Storage — local (Option A)
STORAGE_TYPE=local
LOCAL_STORAGE_PATH=./data/storage

# Storage — MinIO/S3 (Option B)
STORAGE_TYPE=s3
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=moneyapp
MINIO_USE_SSL=false

# Inherited from M1
JWT_SECRET=test-secret-m2
TOKEN_EXPIRY_HOURS=24
BCRYPT_COST=12
CURRENCY=VND
PORT=8080
DB_PATH=./moneyapp.db
```

### Seeded test credentials

| Username | Password | Notes |
|---|---|---|
| `admin` | `changeme` | Same seed user from M1 migration `002_seed_default_user.up.sql`. |

---

## Traceability Matrix

| Ticket | Title | Test Suite | Case Count | AC Bullets Covered |
|---|---|---|---|---|
| E7-S1 | File Upload Infrastructure | TS-25 | 12 | AC1 (local), AC2 (s3), AC3 (type reject), AC4 (cascade note), AC5 (orphan) |
| E7-S2 | File Size Validation | TS-26 | 5 | AC1 (client), AC2 (server 413) |
| E7-S3 | File Download | TS-27 | 5 | AC1 (download), AC2 (404) |
| E7-S4 | Image Thumbnail Preview | TS-28 | 5 | AC1 (JPEG thumb), AC2 (PDF icon), AC3 (click expand) |
| E7-S5 | Cascade Delete Attachments | TS-29 | 6 | AC1 (cascade), AC2 (storage down) |
| E2-S7 | Attach Receipts to Expenses | TS-30 | 5 | AC1 (upload), AC2 (indicator) |
| E3-S5 | Attach Documents to Income | TS-31 | 4 | AC1 (upload), AC2 (indicator) |
| E4-S5 | Attach Invoice PDFs | TS-32 | 4 | AC1 (upload+preview), AC2 (click open) |
| E4-S7 | Overdue Check Admin Endpoint | TS-33 | 4 | AC1 (updated_count), AC2 (zero count) |
| E5-S2 | Create Custom Categories | TS-34 | 6 | AC1 (create), AC2 (409 duplicate), AC3 (is_default=false) |
| E5-S3 | Rename and Delete Custom Categories | TS-35 | 7 | AC1 (rename), AC2 (403 default), AC3 (confirm dialog) |
| E5-S4 | Reassign Transactions on Category Delete | TS-36 | 5 | AC1 (reassign), AC2 (view uncategorized), AC3 (no data loss) |
| M2-01 | Expense List Sorting | TS-37 | 5 | AC1 (sort desc), AC2 (toggle asc) |
| M2-02 | Income Filter by Category | TS-38 | 4 | AC1 (filter), AC2 (clear) |
| M2-03 | Invoice Date Range Filter | TS-39 | 5 | AC1 (issue_date), AC2 (due_date) |
| M2-04 | Income Running Total | TS-40 | 3 | AC1 (total displayed) |

**Total planned test cases: 85**

---

## E7-S1 — File Upload Infrastructure

**Ticket**: E7-S1 | **NF**: FA-01, FA-02, FA-03, NF-06, NF-07, NF-09, NF-11, NF-12 | **Priority**: Must  
**Files under test**: `internal/storage/storage.go`, `internal/storage/local.go`, `internal/storage/minio.go`, `internal/services/attachment_service.go`, `internal/handlers/attachment_handler.go`, `internal/models/attachment.go`, `backend/migrations/008_create_attachments.up.sql`  
**Endpoints**: `POST /api/attachments`, `GET /api/attachments?entity_type=&entity_id=`, `DELETE /api/attachments/:id`

### Preconditions

- Backend running in `m2-seeded` environment (local or MinIO).
- Valid auth token available (login as `admin`/`changeme`).
- At least one expense with ID 1 exists.
- Test fixture files prepared (see Environment Setup).

### Test Suite TS-25

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-25-01 | Smoke | Upload JPEG to expense — local storage — record created | `curl -s -X POST -H "Authorization: Bearer <token>" -F "entity_type=expense" -F "entity_id=1" -F "file=@/tmp/test_receipt.jpg" http://localhost:8080/api/attachments` | HTTP 201; response body: `{"data":{"id":1,"entity_type":"expense","entity_id":1,"filename":"test_receipt.jpg","mime_type":"image/jpeg","size_bytes":<N>,"storage_key":"expense/1/<uuid>_test_receipt.jpg","created_at":"..."}}`. File written to `LOCAL_STORAGE_PATH/expense/1/`. DB row in `attachments` table. | FA-01, FA-03, AC1 |
| TS-25-02 | Smoke | Upload JPEG to expense — MinIO/S3 — object stored | (Run in **m2-minio** env.) Same curl as TS-25-01. After upload, verify via MinIO console or `mc ls myminio/moneyapp/expense/1/`. | HTTP 201; object present in MinIO bucket under key `expense/1/<uuid>_test_receipt.jpg`. DB row created. | FA-03, AC2 |
| TS-25-03 | Smoke | Upload PDF — accepted | `curl ... -F "file=@/tmp/test_invoice.pdf"` with `entity_type=invoice` and valid `entity_id`. | HTTP 201; `mime_type: "application/pdf"` in response. | FA-01 |
| TS-25-04 | Smoke | Upload PNG — accepted | `curl ... -F "file=@/tmp/test_image.png"` with `entity_type=expense`. | HTTP 201; `mime_type: "image/png"` in response. | FA-01 |
| TS-25-05 | Regression | **Negative**: Upload `.exe` file — rejected | `curl ... -F "file=@/tmp/test_malicious.exe"` | HTTP 400; body contains `{"error":"Only PDF, JPEG, and PNG files are allowed"}`. No file written to storage. No DB row created. | FA-01, NF-09 |
| TS-25-06 | Regression | List attachments for an entity | 1. Upload 2 files to expense ID 1. 2. `curl -s -H "Authorization: Bearer <token>" "http://localhost:8080/api/attachments?entity_type=expense&entity_id=1"` | HTTP 200; `data` array contains 2 items; each item has `id`, `filename`, `mime_type`, `size_bytes`, `storage_key`, `created_at`. | FA-06 |
| TS-25-07 | Regression | Delete single attachment | 1. Get attachment ID from TS-25-01. 2. `curl -s -X DELETE -H "Authorization: Bearer <token>" http://localhost:8080/api/attachments/<id>` | HTTP 204. File absent from storage path (local) or MinIO bucket. DB row deleted: `SELECT COUNT(*) FROM attachments WHERE id=<id>` = 0. | AC (delete) |
| TS-25-08 | Regression | Storage key format | Inspect `storage_key` from TS-25-01. | Format matches `expense/1/<uuid>_test_receipt.jpg` — entity type / entity ID / UUID underscore filename. UUID is a valid UUID v4. | FA-03 |
| TS-25-09 | Regression | **Negative**: Upload without auth | `curl -s -X POST -F "entity_type=expense" -F "entity_id=1" -F "file=@/tmp/test_receipt.jpg" http://localhost:8080/api/attachments` | HTTP 401. No storage write. | NF-07 |
| TS-25-10 | Regression | **Negative**: Invalid `entity_type` value | `curl ... -F "entity_type=account" -F "entity_id=1" -F "file=@/tmp/test_receipt.jpg"` | HTTP 400; error references allowed entity types (`expense`, `income`, `invoice`). The `attachments` table `CHECK` constraint also guards this. | NF-09 |
| TS-25-11 | Regression | Storage unavailable — graceful error (NF-12) | (In **m2-storage-down** env: set `LOCAL_STORAGE_PATH` to `/root/no-write-perm`.) Attempt `POST /api/attachments`. | HTTP 500 or 503 with `{"error":"storage unavailable"}` or similar. No orphaned DB row inserted. | NF-12 |
| TS-25-12 | Regression | `attachments` migration creates table and index | Start backend in **m2-fresh** env. `sqlite3 moneyapp.db ".schema attachments"` | Schema matches: `id INTEGER PRIMARY KEY AUTOINCREMENT`, `entity_type TEXT NOT NULL CHECK(...)`, `entity_id INTEGER NOT NULL`, `filename TEXT NOT NULL`, `mime_type TEXT NOT NULL CHECK(...)`, `size_bytes INTEGER NOT NULL`, `storage_key TEXT NOT NULL UNIQUE`, `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`. Index `idx_attachments_entity ON attachments(entity_type, entity_id)` exists. | NF-19 |

#### Security note

The `GET /api/attachments/:id/download` and `GET /api/attachments/:id/preview` endpoints must require a valid JWT. Raw storage paths (`LOCAL_STORAGE_PATH` filesystem directory or MinIO bucket objects) must never be directly accessible without authentication. Verify that setting `STORAGE_TYPE=local` with `LOCAL_STORAGE_PATH` inside the static file serving root does not inadvertently expose files — the storage path must be outside any public HTTP directory.

---

## E7-S2 — File Size Validation

**Ticket**: E7-S2 | **NF**: FA-02 | **Priority**: Must  
**Files under test**: `internal/handlers/attachment_handler.go` (`http.MaxBytesReader`), `internal/services/attachment_service.go` (`Upload` size check), `src/components/attachments/FileUpload.tsx`  
**Endpoint**: `POST /api/attachments`

### Preconditions

- Backend running in `m2-seeded` environment.
- `/tmp/test_oversized.bin` (11 MB) prepared.
- Valid auth token.

### Test Suite TS-26

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-26-01 | Smoke | **Client-side**: 11MB file rejected before upload | 1. Open expense edit page in browser. 2. Attempt to select `/tmp/test_oversized.bin` (≥ 11 MB) via the `FileUpload` component. | Client-side validation fires before any HTTP request; UI shows "File exceeds the 10 MB limit" or "File must be under 10 MB". No `POST /api/attachments` request appears in DevTools Network tab. | FA-02, AC1 |
| TS-26-02 | Smoke | **Server-side**: Crafted 15MB upload returns 413 | `curl -s -X POST -H "Authorization: Bearer <token>" -F "entity_type=expense" -F "entity_id=1" -F "file=@/tmp/test_oversized.bin" http://localhost:8080/api/attachments` | HTTP 413 (Payload Too Large); body `{"error":"file size exceeds 10 MB limit"}` or similar. No file written to storage. No DB row inserted. | FA-02, AC2 |
| TS-26-03 | Regression | Exactly 10MB file — accepted | Create a file exactly `10 * 1024 * 1024` bytes (`dd if=/dev/zero of=/tmp/test_exactly10mb.bin bs=1M count=10`). Upload via curl (with correct mime type workaround: use `-F "file=@/tmp/test_exactly10mb.bin;type=image/jpeg"`). | HTTP 201 or HTTP 400 depending on whether the size check is strict `<` or `<=`. Ticket AC says "over 10MB"; a file of exactly 10MB should be accepted (`<= 10*1024*1024`). Verify implemented boundary. | FA-02 (boundary) |
| TS-26-04 | Regression | File just over 10MB (10MB + 1 byte) rejected | `dd if=/dev/zero of=/tmp/test_10mb1.bin bs=1 count=$((10*1024*1024+1))`. Upload. | HTTP 413. | FA-02 |
| TS-26-05 | Regression | Client-side validation: exactly-at-limit file — no error | (Browser test.) Select a file of exactly 10MB. | FileUpload component does not show an error; upload proceeds to the server. | FA-02 (boundary) |

---

## E7-S3 — File Download

**Ticket**: E7-S3 | **NF**: FA-04, NF-07 | **Priority**: Must  
**Files under test**: `internal/handlers/attachment_handler.go` (`GET /api/attachments/:id/download`)  
**Endpoint**: `GET /api/attachments/:id/download`

### Preconditions

- At least one attachment uploaded for an expense (attachment ID known).
- Valid auth token.
- Run in both `m2-local` and `m2-minio` environments.

### Test Suite TS-27

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-27-01 | Smoke | Download JPEG attachment — correct headers and content | 1. Upload `test_receipt.jpg` (note attachment ID). 2. `curl -s -OJ -H "Authorization: Bearer <token>" http://localhost:8080/api/attachments/<id>/download` | HTTP 200; `Content-Type: image/jpeg`; `Content-Disposition: attachment; filename="test_receipt.jpg"`. Downloaded bytes match the original upload (MD5/SHA256 identical). | FA-04, AC1 |
| TS-27-02 | Smoke | Download PDF attachment — correct Content-Type | Upload `test_invoice.pdf`. Download via `/api/attachments/<id>/download`. | HTTP 200; `Content-Type: application/pdf`; `Content-Disposition: attachment; filename="test_invoice.pdf"`. | FA-04 |
| TS-27-03 | Regression | Download — storage proxied, not direct URL (NF-07) | Inspect the download response. Verify the file is streamed by the backend, not a redirect to a storage URL (no `Location` header pointing to MinIO/local path). | Response body is the raw file bytes served by the Go handler. No `Location` header unless pre-signed URL approach is deliberately chosen and the URL expires in ≤ 15 minutes (per ticket note). | NF-07 |
| TS-27-04 | Regression | **Negative**: Download non-existent attachment ID | `curl -s -H "Authorization: Bearer <token>" http://localhost:8080/api/attachments/99999/download` | HTTP 404; body `{"error":"attachment not found"}` or similar. | AC2 |
| TS-27-05 | Regression | **Negative**: Download without authentication | `curl -s http://localhost:8080/api/attachments/<id>/download` (no Authorization header) | HTTP 401. File bytes not returned. | NF-07 |

---

## E7-S4 — Image Thumbnail Preview

**Ticket**: E7-S4 | **NF**: FA-07, FA-08 | **Priority**: Should  
**Files under test**: `internal/handlers/attachment_handler.go` (`GET /api/attachments/:id/preview`), `src/components/attachments/AttachmentList.tsx`  
**Endpoint**: `GET /api/attachments/:id/preview`

### Preconditions

- At least one JPEG and one PDF attachment uploaded (attachment IDs known).
- Valid auth token.
- Browser available for frontend thumbnail checks.

### Test Suite TS-28

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-28-01 | Smoke | JPEG attachment — thumbnail `<img>` rendered in AttachmentList | 1. Upload a JPEG to an expense. 2. Open the expense edit view in browser. | `AttachmentList` shows an `<img>` tag pointing to `/api/attachments/<id>/preview`, rendered at thumbnail size (~80×80px). Image is visually correct (not broken). | FA-07, AC1 |
| TS-28-02 | Smoke | PDF attachment — PDF icon shown (not image thumbnail) | 1. Upload a PDF to an invoice. 2. Open invoice edit view in browser. | `AttachmentList` shows a PDF icon/glyph with the filename "test_invoice.pdf". No `<img>` tag with broken src is shown. | FA-08, AC2 |
| TS-28-03 | Regression | Preview endpoint — image served inline | `curl -s -I -H "Authorization: Bearer <token>" http://localhost:8080/api/attachments/<jpeg_id>/preview` | HTTP 200; `Content-Type: image/jpeg`; `Content-Disposition: inline` (not `attachment`). | FA-07 |
| TS-28-04 | Regression | Preview endpoint — PDF served inline for browser viewer | `curl -s -I -H "Authorization: Bearer <token>" http://localhost:8080/api/attachments/<pdf_id>/preview` | HTTP 200; `Content-Type: application/pdf`; `Content-Disposition: inline`. | FA-08 |
| TS-28-05 | Regression | Click image thumbnail — larger preview opens | 1. JPEG attachment visible in AttachmentList. 2. Click the thumbnail. | A modal or lightbox opens showing the full-size image. The image is served from `/api/attachments/<id>/preview`. | AC3 |

---

## E7-S5 — Cascade Delete Attachments

**Ticket**: E7-S5 | **NF**: FA-05, NF-11 | **Priority**: Must  
**Files under test**: `internal/services/expense_service.go` (`Delete`), `internal/services/income_service.go` (`Delete`), `internal/services/invoice_service.go` (`Delete`), `internal/services/attachment_service.go` (`DeleteByEntity`)  
**Endpoints**: `DELETE /api/expenses/:id`, `DELETE /api/incomes/:id`, `DELETE /api/invoices/:id`

### Preconditions

- Expense with ID `<exp_id>` has 2 uploaded attachments (storage keys noted).
- Income with ID `<inc_id>` has 1 uploaded attachment.
- Invoice with ID `<inv_id>` has 1 uploaded attachment.
- Valid auth token.
- Run in both `m2-local` (verify files absent from disk) and `m2-minio` (verify objects absent from bucket).

### Test Suite TS-29

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-29-01 | Smoke | Delete expense — 2 attachments removed from storage and DB | 1. Note expense `<exp_id>` with 2 attachment storage keys. 2. `curl -s -X DELETE -H "Authorization: Bearer <token>" http://localhost:8080/api/expenses/<exp_id>` | HTTP 204. Both files absent from `LOCAL_STORAGE_PATH/expense/<exp_id>/` (or MinIO bucket). `SELECT COUNT(*) FROM attachments WHERE entity_type='expense' AND entity_id=<exp_id>` = 0. | FA-05, NF-11, AC1 |
| TS-29-02 | Smoke | Delete income — attachment removed from storage and DB | Delete income `<inc_id>`. | HTTP 204. Storage key absent. `SELECT COUNT(*) FROM attachments WHERE entity_type='income' AND entity_id=<inc_id>` = 0. | FA-05, AC1 |
| TS-29-03 | Smoke | Delete invoice — attachment removed from storage and DB | Delete invoice `<inv_id>`. | HTTP 204. Storage key absent. DB attachment records for invoice deleted. | FA-05, AC1 |
| TS-29-04 | Regression | Delete expense with no attachments — works normally | Delete an expense that has zero attachments. | HTTP 204. No error (no crash from calling `DeleteByEntity` on an entity with 0 attachments). | AC1 (boundary) |
| TS-29-05 | Regression | Storage unreachable during delete — DB records still deleted, error logged | (In **m2-storage-down** env.) Delete an expense with attachments. | HTTP 204 or 500 (per ticket: "log error but still delete DB record"). DB attachment rows deleted. Server log contains "orphaned file" or "failed to delete from storage" warning. Expense record itself is deleted. | NF-11, NF-12, AC2 |
| TS-29-06 | Regression | Attachment DB rows are within same transaction as parent delete | Check implementation: `ExpenseService.Delete()` wraps in a DB transaction. | `sqlite3 moneyapp.db "BEGIN; DELETE FROM attachments WHERE entity_type='expense' AND entity_id=<id>; ROLLBACK;"` (manual transaction test). Unit test covers: if expense delete fails mid-transaction, attachment rows are not deleted. | NF-11 |

#### Storage orphan edge cases (E7-S5 AC)

The following scenarios may cause orphaned storage files and should be noted for the D3 audit (future milestone). Document these as known residual risks:

| Scenario | Risk |
|---|---|
| Storage upload succeeds, DB insert fails (TS-25-11) | Storage object exists with no DB record — orphaned object. Cleaned by future audit endpoint. |
| Storage delete fails during cascade (TS-29-05) | Storage object persists after DB record deleted — orphaned file. Logged but not auto-cleaned in M2. |
| Partial multi-attachment cascade (storage fails on 2nd of 3 files) | First file deleted, 2nd and 3rd may remain. Log must name which keys failed. |

---

## E2-S7 — Attach Receipts to Expenses

**Ticket**: E2-S7 | **NF**: FA-01, FA-06, FA-07 | **Priority**: Should  
**Files under test**: `src/components/expenses/ExpenseForm.tsx`, `src/components/attachments/AttachmentList.tsx`, `src/components/attachments/FileUpload.tsx`, `src/api/attachments.ts`  
**Backend**: Generic attachment endpoints from E7-S1 (`entity_type=expense`).

### Preconditions

- Expense with known ID exists.
- Browser and frontend dev server running.
- Valid auth token (logged in as `admin`).

### Test Suite TS-30

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-30-01 | Smoke | Upload receipt JPEG in expense edit view — appears in attachment list | 1. Navigate to expense edit modal/view for expense ID 1. 2. Locate `FileUpload` component below the form fields. 3. Select `test_receipt.jpg`. 4. Confirm upload. | Upload progress indicator shown during upload. After success, `AttachmentList` refreshes and shows "test_receipt.jpg" with filesize and actions (Download, Delete). Attachment icon count on expense list row updates. | AC1 |
| TS-30-02 | Smoke | Expense list shows attachment indicator | 1. Create/open an expense that now has ≥ 1 attachment. 2. Navigate to `/expenses` list. | A paperclip icon or badge (e.g., "📎 1") is shown in the expense list row for that expense. Expenses with 0 attachments show no indicator. | AC2 |
| TS-30-03 | Regression | Multiple files attached to one expense | Upload 3 separate files (JPEG, PNG, PDF) to expense ID 1 via the `FileUpload` component. | All 3 appear in `AttachmentList`. `GET /api/attachments?entity_type=expense&entity_id=1` returns 3 items. | FA-06 |
| TS-30-04 | Regression | Delete individual attachment from expense | From the AttachmentList in expense edit view, click Delete on one attachment. Confirm dialog. | That file is removed from the list and from storage. Other attachments remain. `DELETE /api/attachments/<id>` returns 204. | NF-15 (confirm) |
| TS-30-05 | Regression | Download receipt from expense attachment list | Click "Download" on a JPEG attachment in the expense edit view. | Browser initiates file download with original filename. File content matches the uploaded bytes. | FA-04 |

---

## E3-S5 — Attach Documents to Income

**Ticket**: E3-S5 | **NF**: FA-01, FA-06 | **Priority**: Should  
**Files under test**: `src/components/income/IncomeForm.tsx`, `src/components/attachments/AttachmentList.tsx`, `src/components/attachments/FileUpload.tsx`  
**Backend**: Generic attachment endpoints with `entity_type=income`.

### Preconditions

- Income record with known ID exists.
- Browser running.
- Valid auth token.

### Test Suite TS-31

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-31-01 | Smoke | Upload pay slip PDF in income edit view | 1. Navigate to income edit modal/view. 2. Select `test_invoice.pdf` via `FileUpload`. 3. Confirm upload. | `POST /api/attachments` called with `entity_type=income`. After upload, `AttachmentList` shows the PDF with filename, size, and Download/Delete actions. | AC1 |
| TS-31-02 | Smoke | Income list shows attachment indicator | Open income record that has ≥ 1 attachment. Navigate to `/income` list. | Attachment indicator visible on that row; rows without attachments show none. | AC2 |
| TS-31-03 | Regression | API correctly receives `entity_type=income` | Upload via the income form. Inspect `POST /api/attachments` request in browser DevTools Network tab. | Form field `entity_type` = `"income"` in the multipart request. Response `entity_type` = `"income"`. | FA-03 |
| TS-31-04 | Regression | `GET /api/attachments?entity_type=income&entity_id=<id>` returns correct records | After uploading to income ID 2, query: `curl -s -H "Authorization: Bearer <token>" "http://localhost:8080/api/attachments?entity_type=income&entity_id=2"` | Only attachments for income ID 2 returned; no expense or invoice attachments mixed in. | FA-03 |

---

## E4-S5 — Attach Invoice PDFs

**Ticket**: E4-S5 | **NF**: FA-04, FA-08 | **Priority**: Must  
**Files under test**: `src/components/invoices/InvoiceForm.tsx`, `src/components/attachments/AttachmentList.tsx`, `src/components/attachments/FileUpload.tsx`  
**Backend**: Generic attachment endpoints with `entity_type=invoice`.

### Preconditions

- Invoice record with known ID exists.
- Browser running.
- Valid auth token.

### Test Suite TS-32

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-32-01 | Smoke | Upload invoice PDF — prominent FileUpload section shown in InvoiceForm | 1. Open invoice edit/detail view. 2. Observe the `FileUpload` component placement. | `FileUpload` is a prominent section in the form (not a footer footnote). Select `test_invoice.pdf`. Upload progress shown. On completion, PDF attachment appears in `AttachmentList` with PDF icon and filename. | AC1 |
| TS-32-02 | Smoke | PDF preview visible inline in invoice attachment list | After upload, `AttachmentList` shows the PDF entry. | For the PDF attachment, an `<iframe>` or `<embed>` tag (or a link that opens the preview endpoint) is shown pointing to `/api/attachments/<id>/preview`. Clicking it opens the full PDF in the browser's built-in viewer. | FA-08, AC2 |
| TS-32-03 | Regression | JPEG attachment — thumbnail shown instead of PDF icon | Upload a JPEG receipt to an invoice. | In `AttachmentList`, JPEG renders as a thumbnail (not a PDF icon). | FA-07 |
| TS-32-04 | Regression | `entity_type=invoice` in API request | Inspect the `POST /api/attachments` multipart body when uploading via the invoice form. | Form data includes `entity_type=invoice` and the correct `entity_id`. | FA-03 |

---

## E4-S7 — Overdue Check Admin Endpoint

**Ticket**: E4-S7 | **NF**: NF-09 | **Priority**: Should  
**Files under test**: `internal/handlers/invoice_handler.go` (`POST /api/invoices/check-overdue`), `internal/services/invoice_service.go` (`UpdateOverdueStatuses`)  
**Endpoint**: `POST /api/invoices/check-overdue`

### Preconditions

- 2 invoices with `status=unpaid` and `due_date` in the past (e.g., `2020-01-01`) exist.
- 1 invoice with `status=unpaid` and `due_date` in the future exists.
- 1 invoice with `status=paid` and `due_date` in the past exists (must not be transitioned).
- Valid auth token.

### Test Suite TS-33

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-33-01 | Smoke | 2 past-due invoices — response `updated_count: 2` | `curl -s -X POST -H "Authorization: Bearer <token>" http://localhost:8080/api/invoices/check-overdue` | HTTP 200; body `{"updated_count": 2}`. Both invoices changed to `status=overdue` in DB. | AC1 |
| TS-33-02 | Smoke | No invoices past due — `updated_count: 0` | Ensure all unpaid invoices have future due dates. `curl -s -X POST -H "Authorization: Bearer <token>" http://localhost:8080/api/invoices/check-overdue` | HTTP 200; body `{"updated_count": 0}`. | AC2 |
| TS-33-03 | Regression | Server logs "Updated N invoices to overdue status" | After TS-33-01, inspect backend stdout/log output. | Log line at INFO level contains "Updated 2 invoices to overdue status" or equivalent with N = 2. | AC1 (logging) |
| TS-33-04 | Regression | **Negative**: Unauthenticated request rejected | `curl -s -X POST http://localhost:8080/api/invoices/check-overdue` (no Authorization header) | HTTP 401. No overdue update performed. | NF-07 |

#### Routing note

Verify `POST /api/invoices/check-overdue` does not conflict with the `GET /api/invoices/stats` route (both are fixed-path routes, not `:id`). The Go 1.22+ `net/http` mux should handle both without ambiguity when method-prefixed patterns are used.

```bash
# Verify check-overdue is not caught by GET /:id
curl -s -X POST -H "Authorization: Bearer <token>" http://localhost:8080/api/invoices/check-overdue
# Expected: {"updated_count": N}

# Verify stats still works
curl -s -H "Authorization: Bearer <token>" http://localhost:8080/api/invoices/stats
# Expected: {"total_outstanding": ..., "unpaid_count": ..., "overdue_count": ...}
```

---

## E5-S2 — Create Custom Categories

**Ticket**: E5-S2 | **NF**: NF-09, NF-14 | **Priority**: Should  
**Files under test**: `internal/services/category_service.go` (`Create`), `internal/handlers/category_handler.go`, `src/pages/CategoriesPage.tsx`, `src/components/categories/CategoryForm.tsx`  
**Endpoint**: `POST /api/categories`

### Preconditions

- Default categories present (M1 migration `004_seed_default_categories.up.sql`).
- Valid auth token.

### Test Suite TS-34

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-34-01 | Smoke | Create custom expense category — appears in list | `curl -s -X POST -H "Authorization: Bearer <token>" -H "Content-Type: application/json" -d '{"name":"Pets","type":"expense"}' http://localhost:8080/api/categories` | HTTP 201; body `{"data":{"id":<new_id>,"name":"Pets","type":"expense","is_default":false}}`. `GET /api/categories?type=expense` now includes "Pets". | AC1, AC3 |
| TS-34-02 | Smoke | Create custom income category | `curl ... -d '{"name":"Consulting","type":"income"}'` | HTTP 201; `type: "income"`, `is_default: false`. `GET /api/categories?type=income` includes "Consulting". | AC1, AC3 |
| TS-34-03 | Smoke | Custom category appears in expense form dropdown | 1. Create "Pets" category (TS-34-01). 2. Open "Add Expense" form in browser. | Category dropdown includes "Pets" in the expense category list. | AC1 (UI) |
| TS-34-04 | Regression | **Negative**: Duplicate name + type — 409 | `curl ... -d '{"name":"Pets","type":"expense"}'` (after TS-34-01) | HTTP 409; body `{"error":"Category already exists"}`. No new row inserted. | AC2 |
| TS-34-05 | Regression | **Negative**: Empty name rejected | `curl ... -d '{"name":"","type":"expense"}'` | HTTP 400; field-level error about name being required. | NF-14 |
| TS-34-06 | Regression | **Negative**: Invalid category type | `curl ... -d '{"name":"Test","type":"transaction"}'` | HTTP 400; error referencing allowed types (`expense`, `income`). | NF-09 |

---

## E5-S3 — Rename and Delete Custom Categories

**Ticket**: E5-S3 | **NF**: NF-09, NF-15 | **Priority**: Should  
**Files under test**: `internal/services/category_service.go` (`Update`, `Delete`), `internal/handlers/category_handler.go`, `src/pages/CategoriesPage.tsx`  
**Endpoints**: `PUT /api/categories/:id`, `DELETE /api/categories/:id`

### Preconditions

- Custom category "Pets" (expense, `is_default=false`) exists from TS-34-01.
- Default category "Food" (expense, `is_default=true`) exists from seed.
- Valid auth token.

### Test Suite TS-35

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-35-01 | Smoke | Rename custom category | `curl -s -X PUT -H "Authorization: Bearer <token>" -H "Content-Type: application/json" -d '{"name":"Pet Care"}' http://localhost:8080/api/categories/<pets_id>` | HTTP 200; body shows `name: "Pet Care"`. `GET /api/categories?type=expense` shows "Pet Care" (not "Pets"). | AC1 |
| TS-35-02 | Smoke | Renamed name reflected in expense form dropdown | After TS-35-01, open "Add Expense" form in browser. | Dropdown shows "Pet Care", not "Pets". Existing expenses categorized as "Pets" now show "Pet Care" in their category column. | AC1 (UI) |
| TS-35-03 | Smoke | Delete custom category — returns 204 | `curl -s -X DELETE -H "Authorization: Bearer <token>" http://localhost:8080/api/categories/<pets_id>` | HTTP 204. `GET /api/categories?type=expense` no longer includes "Pet Care". | AC2 (delete flow; see E5-S4 for reassignment) |
| TS-35-04 | Regression | **Negative**: Rename default category — 403 | `curl -s -X PUT -H "Authorization: Bearer <token>" -H "Content-Type: application/json" -d '{"name":"Junk Food"}' http://localhost:8080/api/categories/<food_id>` | HTTP 403; body `{"error":"Default categories cannot be renamed"}`. Category name unchanged. | AC2 |
| TS-35-05 | Regression | **Negative**: Delete default category — 403 | `curl -s -X DELETE -H "Authorization: Bearer <token>" http://localhost:8080/api/categories/<food_id>` | HTTP 403; body `{"error":"Default categories cannot be deleted"}`. Category still present. | AC2 |
| TS-35-06 | Regression | Default categories show lock icon (no Edit/Delete in UI) | Navigate to `/categories` page in browser. | Default categories (e.g., Food, Transport, Salary) display a lock icon or visual indicator. No Edit or Delete buttons visible for them. Custom categories show Edit/Delete buttons. | AC2 (UI) |
| TS-35-07 | Regression | Delete confirmation dialog warns about transaction reassignment | Click Delete on a custom category in the browser UI. | Confirmation dialog appears mentioning that associated transactions will be moved to "Uncategorized". User must confirm before deletion proceeds. (NF-15) | AC3, NF-15 |

---

## E5-S4 — Reassign Transactions on Category Delete

**Ticket**: E5-S4 | **NF**: NF-11 | **Priority**: Should  
**Files under test**: `internal/services/category_service.go` (`Delete` with reassignment), `backend/migrations/004_seed_default_categories.up.sql` ("Uncategorized" seed)

### Preconditions

- Custom expense category "Pets" with 5 associated expense records exists.
- Custom income category "Consulting" with 3 associated income records exists.
- "Uncategorized" expense and income default categories are seeded.
- Valid auth token.

### Test Suite TS-36

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-36-01 | Smoke | Delete "Pets" — 5 expenses moved to Uncategorized | 1. Note `<uncategorized_expense_id>` = ID of "Uncategorized" expense category. 2. `DELETE /api/categories/<pets_id>`. 3. `SELECT category_id FROM expenses WHERE category_id = <pets_id>` → empty. 4. `SELECT COUNT(*) FROM expenses WHERE category_id = <uncategorized_expense_id>` | Step 2 returns 204. Step 3 returns 0 rows. Step 4 shows count increased by 5 (all former "Pets" expenses now show "Uncategorized"). | AC1 |
| TS-36-02 | Smoke | Delete "Consulting" — 3 incomes moved to Uncategorized income | Same pattern for income category. | All 3 income records reassigned to "Uncategorized" income category. Income count unchanged. | AC1 |
| TS-36-03 | Regression | No data loss — expense count unchanged | Before delete: `SELECT COUNT(*) FROM expenses` → N. After delete: `SELECT COUNT(*) FROM expenses` → N. | Total expense count is identical. No records deleted. | AC3 |
| TS-36-04 | Regression | View reassigned expenses — category shows "Uncategorized" | After deleting "Pets", navigate to `/expenses` in browser. Filter by "Uncategorized" category. | Former "Pets" expenses appear in the filtered list under "Uncategorized". | AC2 |
| TS-36-05 | Regression | **Negative**: Cannot delete "Uncategorized" (it is default) | `curl -s -X DELETE -H "Authorization: Bearer <token>" http://localhost:8080/api/categories/<uncategorized_id>` | HTTP 403. "Uncategorized" category still present and is_default=true. | AC1 (integrity guard) |

---

## M2-01 — Expense List Sorting

**Ticket**: M2-01 | **NF**: NF-04 | **Priority**: Should  
**Files under test**: `internal/services/expense_service.go` (`List` — sort params), `internal/handlers/expense_handler.go`, `src/pages/ExpensesPage.tsx`  
**Endpoint**: `GET /api/expenses?sort_by=date|amount&sort_order=asc|desc`

### Preconditions

- At least 5 expense records with varying amounts and dates.
- Valid auth token.

### Test Suite TS-37

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-37-01 | Smoke | Default sort — date descending | `curl -s -H "Authorization: Bearer <token>" "http://localhost:8080/api/expenses?page=1&per_page=20"` | Items returned in descending date order (most recent first). No `sort_by` param required for default. | AC1 (default behavior) |
| TS-37-02 | Smoke | Sort by amount descending | `curl ... "?sort_by=amount&sort_order=desc"` | Items returned with highest amount first. Verify first item has amount ≥ second item ≥ third item. | AC1 |
| TS-37-03 | Smoke | Toggle sort to amount ascending | `curl ... "?sort_by=amount&sort_order=asc"` | Items returned with lowest amount first. Verify first item has amount ≤ second item ≤ third item. | AC2 |
| TS-37-04 | Regression | Sort by date ascending | `curl ... "?sort_by=date&sort_order=asc"` | Items returned oldest-first. | AC2 (toggle) |
| TS-37-05 | Regression | **Negative**: Invalid sort_by column — 400 or default | `curl ... "?sort_by=description&sort_order=desc"` | HTTP 400 with `{"error":"invalid sort column"}` — whitelisted columns only (SQL injection prevention), OR falls back to default sort without error. Document which approach is implemented. | AC (security: whitelist) |

#### Frontend checks (M2-01)

| Check | Steps | Expected |
|---|---|---|
| Clicking "Amount" header sorts by amount desc | Navigate to `/expenses`. Click "Amount" column header. | URL updates to `?sort_by=amount&sort_order=desc`. List reorders. Arrow indicator appears on Amount column. |
| Clicking "Amount" header again toggles to asc | Click "Amount" header a second time. | URL updates to `?sort_by=amount&sort_order=asc`. Arrow direction reverses. |
| Sort state persists in URL — survives page refresh | Apply sort `?sort_by=amount&sort_order=desc`. Refresh page. | Filter inputs and sort indicator restored. Same sort order shown. |

---

## M2-02 — Income Filter by Category

**Ticket**: M2-02 | **NF**: NF-04 | **Priority**: Should  
**Files under test**: `src/pages/IncomePage.tsx`, `src/components/filters/CategoryFilter.tsx` (reused), `internal/services/income_service.go` (`List` with `category_id`)  
**Endpoint**: `GET /api/incomes?category_id=<id>`

### Preconditions

- Income records in at least 2 different categories (e.g., Salary and Freelance).
- Valid auth token.

### Test Suite TS-38

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-38-01 | Smoke | Filter income by Salary category | `curl -s -H "Authorization: Bearer <token>" "http://localhost:8080/api/incomes?category_id=<salary_id>"` | Only Salary income records returned; Freelance records absent. `total_amount` reflects only Salary records. | AC1 |
| TS-38-02 | Smoke | Clear category filter — all income shown | Apply category filter in browser UI. Click Clear. | All income records shown; `?category_id=` removed from URL. | AC1 (clear) |
| TS-38-03 | Regression | Category filter in URL persists on refresh | Apply `?category_id=<salary_id>` in browser. Refresh page. | Filter is restored from URL; category dropdown shows "Salary" selected; only Salary records displayed. | Parity with E2-S4 AC3 |
| TS-38-04 | Regression | Combined category + date filter | `curl ... "?category_id=<salary_id>&date_from=2026-01-01&date_to=2026-01-31"` | Only January Salary income returned. `total_amount` reflects intersection. | M2-02 + E3-S3 parity |

---

## M2-03 — Invoice Date Range Filter

**Ticket**: M2-03 | **NF**: NF-04 | **Priority**: Should  
**Files under test**: `internal/services/invoice_service.go` (`List` — date params), `internal/handlers/invoice_handler.go`, `src/pages/InvoicesPage.tsx`, `src/components/filters/DateRangeFilter.tsx` (reused)  
**Endpoint**: `GET /api/invoices?date_from=...&date_to=...&date_field=issue_date|due_date`

### Preconditions

- Invoices with `issue_date` in January 2026 (≥ 2) and `issue_date` in March 2026 (≥ 2).
- Invoices with `due_date` in February 2026 (≥ 2).
- Valid auth token.

### Test Suite TS-39

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-39-01 | Smoke | Filter by issue date range — January invoices | `curl -s -H "Authorization: Bearer <token>" "http://localhost:8080/api/invoices?date_from=2026-01-01&date_to=2026-01-31&date_field=issue_date"` | Only invoices with `issue_date` in January 2026 returned. March invoices absent. | AC1 |
| TS-39-02 | Smoke | Filter by due date range — February invoices | `curl ... "?date_from=2026-02-01&date_to=2026-02-28&date_field=due_date"` | Only invoices with `due_date` in February 2026 returned. Filtering correctly uses `due_date` column, not `issue_date`. | AC2 |
| TS-39-03 | Regression | Default `date_field` — applies to issue_date when omitted | `curl ... "?date_from=2026-01-01&date_to=2026-01-31"` (no `date_field`) | Filters by `issue_date` (expected default). Document if implementation defaults to a different field. | AC1 (default) |
| TS-39-04 | Regression | Filter date field toggle in UI | In browser, apply date range filter. Switch toggle from "Issue Date" to "Due Date". | URL updates `date_field=due_date`. List refreshes with due-date-filtered results. | AC2 (UI) |
| TS-39-05 | Regression | Clear filter shows all invoices | Apply date range filter. Click Clear. | All invoices shown regardless of issue or due date. URL query params `date_from`, `date_to`, `date_field` removed. | AC1 (clear) |

---

## M2-04 — Income Running Total

**Ticket**: M2-04 | **NF**: NF-17 | **Priority**: Should  
**Files under test**: `src/pages/IncomePage.tsx` (summary bar component), `internal/services/income_service.go` (`List` — `total_amount` already in response from E3-S1)  
**Endpoint**: `GET /api/incomes` — `total_amount` field in response

### Preconditions

- 3 income records in January 2026 with amounts 5,000,000 + 5,000,000 + 5,000,000 = 15,000,000 total.
- 2 income records in February 2026.
- Valid auth token.

### Test Suite TS-40

| ID | Priority | Scenario | Steps | Expected Result | AC |
|---|---|---|---|---|---|
| TS-40-01 | Smoke | Summary bar shows total for all income | Navigate to `/income` in browser (no filter active). | Summary bar displays "Total: XX,XXX,XXX VND" (formatted per NF-17) where the amount matches `total_amount` from the API response. | AC1, NF-17 |
| TS-40-02 | Smoke | Total updates when date filter applied | Apply date range filter: January 2026. | Summary bar updates to show "15,000,000 VND" (January records only). February records excluded from total. | AC1 |
| TS-40-03 | Regression | Total updates when category filter applied | Apply category filter: "Salary" only. | Summary bar reflects sum of Salary income records only. | M2-02 + M2-04 combined |

---

## Cross-Cutting Checks

These checks apply across the entire M2 feature set and must be verified after all tickets are implemented.

### Auth protection on new M2 endpoints

| Route | Method | Check | Expected |
|---|---|---|---|
| `/api/attachments` | POST | No auth | HTTP 401 |
| `/api/attachments?entity_type=expense&entity_id=1` | GET | No auth | HTTP 401 |
| `/api/attachments/:id` | DELETE | No auth | HTTP 401 |
| `/api/attachments/:id/download` | GET | No auth | HTTP 401 |
| `/api/attachments/:id/preview` | GET | No auth | HTTP 401 |
| `/api/categories` | POST | No auth | HTTP 401 |
| `/api/categories/:id` | PUT | No auth | HTTP 401 |
| `/api/categories/:id` | DELETE | No auth | HTTP 401 |
| `/api/invoices/check-overdue` | POST | No auth | HTTP 401 |

```bash
# Bulk auth check for new M2 routes
TOKEN="" # intentionally empty
for route in \
  "POST http://localhost:8080/api/attachments" \
  "GET http://localhost:8080/api/attachments?entity_type=expense&entity_id=1" \
  "DELETE http://localhost:8080/api/attachments/1" \
  "GET http://localhost:8080/api/attachments/1/download" \
  "GET http://localhost:8080/api/attachments/1/preview" \
  "POST http://localhost:8080/api/categories" \
  "PUT http://localhost:8080/api/categories/1" \
  "DELETE http://localhost:8080/api/categories/1" \
  "POST http://localhost:8080/api/invoices/check-overdue"; do
  METHOD=$(echo $route | awk '{print $1}')
  URL=$(echo $route | awk '{print $2}')
  CODE=$(curl -s -o /dev/null -w "%{http_code}" -X $METHOD "$URL")
  echo "$METHOD $URL → $CODE"
done
# All should return 401
```

### Storage backend parity — local vs MinIO

Run the following test cases in **both** `m2-local` and `m2-minio` environments and confirm identical outcomes:

| Test | m2-local Expected | m2-minio Expected |
|---|---|---|
| TS-25-01 (upload JPEG) | File at `LOCAL_STORAGE_PATH/expense/1/<key>` | Object in MinIO bucket under `expense/1/<key>` |
| TS-27-01 (download JPEG) | File streamed from local disk | Object streamed from MinIO |
| TS-29-01 (cascade delete) | Files absent from disk path | Objects absent from MinIO bucket |
| TS-25-11 (storage unavailable) | Non-writable path → 500 or 503 | MinIO unreachable → 500 or 503 |

### File type enforcement double-check (FA-01)

| File Extension | MIME Heuristic | Expected Server Response |
|---|---|---|
| `.jpg` / `.jpeg` | `image/jpeg` | 201 |
| `.png` | `image/png` | 201 |
| `.pdf` | `application/pdf` | 201 |
| `.gif` | `image/gif` | 400 — type not allowed |
| `.exe` | `application/octet-stream` | 400 — type not allowed |
| `.svg` | `image/svg+xml` | 400 — type not allowed |
| `.jpg` with PDF content (MIME spoofing) | Client sends `image/jpeg`, content is PDF bytes | Behavior depends on sniff strategy — document and test if server checks magic bytes in addition to declared MIME type |

### Integer amounts — no regressions (NF-10)

Confirm all pre-existing amount fields in M1 endpoints are still returned as integers (not floats) after M2 changes. Spot-check:

```bash
TOKEN=$(curl -s -X POST -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"changeme"}' \
  http://localhost:8080/api/auth/login | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/expenses?page=1&per_page=5 \
  | python3 -c "import sys,json; d=json.load(sys.stdin); [print(e['amount'], type(e['amount']).__name__) for e in d['data']]"
# Expected: all amounts printed as int, not float
```

### Category "Uncategorized" invariant

Verify throughout all E5-S2, E5-S3, E5-S4 tests that:

```bash
# Uncategorized categories always exist and are marked default
sqlite3 moneyapp.db "SELECT id, name, type, is_default FROM categories WHERE name='Uncategorized'"
# Expected: 2 rows — one for type=expense, one for type=income, both is_default=1
```

### Transaction integrity — DB state after each delete

After every cascade-delete test (TS-29-01 through TS-29-05), verify no orphaned `attachments` rows remain linked to deleted parents:

```bash
# After deleting an expense with ID X:
sqlite3 moneyapp.db "SELECT COUNT(*) FROM attachments a
  LEFT JOIN expenses e ON a.entity_id = e.id AND a.entity_type = 'expense'
  WHERE a.entity_type = 'expense' AND e.id IS NULL"
# Expected: 0 (no orphaned attachment rows pointing to deleted expenses)
```

### Responsive layout checks (NF-16) — new M2 UI

| Page / Component | Viewport | Check |
|---|---|---|
| `/categories` | 375×812 (mobile) | Category list readable; Add Category button accessible; default lock icons visible |
| Expense edit modal — AttachmentList | 375×812 (mobile) | Attachment list not clipped; thumbnail images scale; Download/Delete buttons accessible (≥ 44×44px touch target) |
| Invoice edit — FileUpload section | 768×1024 (tablet) | FileUpload is prominently placed, not collapsed; upload button accessible |
| `/expenses` list with attachment indicators | 375×812 (mobile) | Paperclip indicator not hidden in table row; amount column not truncated |

---

## Automation Backlog

Candidates for Playwright or Go integration test automation in a future pass. Not required for M2 sign-off.

### Playwright — File upload flow (TS-30, TS-31, TS-32)

| Test name | Suite | Priority | Notes |
|---|---|---|---|
| `should upload JPEG receipt to expense and see it in AttachmentList` | TS-30 | High | Use `page.setInputFiles` on the hidden file input; assert attachment row appears |
| `should show attachment count indicator on expense list row` | TS-30 | High | After upload, navigate back to `/expenses`; assert paperclip icon visible on that row |
| `should reject oversized file client-side before upload` | TS-26 | High | `page.setInputFiles` with 11MB buffer; assert error text; assert no network request |
| `should reject non-PDF/JPEG/PNG file client-side` | TS-25 | Medium | `page.setInputFiles` with `.exe` file; assert type error message |
| `should download attachment and verify filename in header` | TS-27 | Medium | Intercept `GET /api/attachments/:id/download`; assert `Content-Disposition` header |
| `should show JPEG thumbnail and open lightbox on click` | TS-28 | Medium | Upload JPEG; assert `<img>` with `src` containing `preview`; click; assert modal |
| `should show PDF icon (not broken img) for PDF attachment` | TS-28 | Medium | Upload PDF; assert PDF icon element; assert no broken `<img>` |

### Playwright — Category management (TS-34, TS-35, TS-36)

| Test name | Suite | Priority | Notes |
|---|---|---|---|
| `should create custom category and see it in expense form dropdown` | TS-34 | High | Create via CategoriesPage form; navigate to expense form; assert dropdown option |
| `should show 409 error for duplicate category name` | TS-34 | Medium | Create same category twice; assert "Category already exists" error |
| `should block renaming a default category with 403 UI message` | TS-35 | Medium | Attempt to rename "Food"; assert error toast or inline message |
| `should show reassignment confirmation before deleting category` | TS-35 | High | Click Delete on custom category; assert confirmation dialog text mentions "Uncategorized" |
| `should move expenses to Uncategorized after category delete` | TS-36 | High | Create category; create expense with it; delete category; assert expense shows "Uncategorized" |

### Playwright — Enhanced list features (TS-37, TS-38, TS-39, TS-40)

| Test name | Suite | Priority | Notes |
|---|---|---|---|
| `should sort expense list by amount and toggle direction` | TS-37 | High | Click Amount header twice; assert URL params; assert row order |
| `should filter income by category and update total` | TS-38 | High | Select Salary filter; assert total changes; assert only Salary records shown |
| `should filter invoices by due date range` | TS-39 | Medium | Select "Due Date" toggle; apply range; assert date_field=due_date in URL |
| `should display income running total in summary bar` | TS-40 | High | Apply date filter; assert summary bar text contains formatted total |

### Go integration tests — attachment service (TS-25)

```
Package: internal/services
  - TestAttachmentService_Upload_LocalStorage_JPEG: in-memory SQLite + local temp dir; assert file written, DB row created
  - TestAttachmentService_Upload_LocalStorage_TypeReject: .exe file; assert error, no DB row
  - TestAttachmentService_Upload_SizeReject: file > 10MB; assert 413-equivalent error
  - TestAttachmentService_DeleteByEntity_RemovesFiles: upload 2 files; call DeleteByEntity; assert files gone, rows gone
  - TestAttachmentService_DeleteByEntity_StorageFail_DBStillClean: mock storage.Delete to fail; assert DB rows deleted, error logged
  - TestCategoryService_Delete_ReassignsToUncategorized: seed custom cat + 5 expenses; delete cat; assert all 5 expenses → Uncategorized
  - TestCategoryService_Delete_DefaultCategory_Rejected: attempt to delete default; assert error
  - TestCategoryService_Create_DuplicateNameType_409: create same name+type twice; assert conflict error
  - TestExpenseService_List_SortByAmount: insert 3 expenses; list sort_by=amount desc; assert order
  - TestInvoiceService_List_FilterByDueDate: insert invoices; filter by due_date range; assert correct subset
```

---

## Coverage Gaps & Notes

Issues and ambiguities identified while authoring this plan against the M2 ticket ACs:

| Gap | Ticket | Detail |
|---|---|---|
| **Storage key collision on re-upload with same filename** | E7-S1 | The storage key format `{entityType}/{entityID}/{uuid}_{filename}` includes a UUID, so collisions should be negligible. However, if the UUID generator is not seeded with entropy on startup, test: upload the same filename twice to the same entity. Verify different storage keys are generated. |
| **`entity_id` existence validation** | E7-S1 | Ticket ACs do not specify whether `POST /api/attachments` with a non-existent `entity_id` (e.g., `entity_id=99999`) returns a 404 or silently creates an orphaned attachment. The implementation should validate entity existence before inserting. TS-25 does not cover this case; it should be added once the behavior is confirmed. |
| **MIME type sniffing vs declared Content-Type** | E7-S1, E7-S2 | The file type check in `AttachmentService.Upload` is based on `header.Header.Get("Content-Type")` (client-declared MIME). A determined attacker can send a PDF renamed to `.jpg` with `Content-Type: image/jpeg`. The ticket does not require magic-byte sniffing; if it is not implemented, document this as a known limitation. |
| **Upload progress for large files** | E7-S1, E2-S7 | The ticket mentions "upload progress indicator" in `FileUpload.tsx`. Playwright's `page.setInputFiles` does not naturally demonstrate progress for small test files. Manual testing with a file in the 5–8MB range on a throttled connection (DevTools → Network throttling) is recommended to verify the progress bar renders correctly and does not freeze. |
| **Concurrent cascade delete + attachment read** | E7-S5 | If a user opens the attachment list for an expense while the expense is being deleted in another tab, the list fetch may return attachments that are subsequently deleted mid-request. This is an edge case not covered by existing ACs — document as a known race condition; no fix required for M2. |
| **`sort_by` SQL injection whitelist** | M2-01 | Ticket explicitly calls out "whitelist allowed sort columns to prevent SQL injection". TS-37-05 tests for this. The tester must confirm that the error response does not leak the SQL query or column name in the error body (NF-09). |
| **Income `category_id` filter — backend already partially done** | M2-02 | Ticket says backend is "already partially covered in E3-S1" but this must be verified. If the `category_id` param was stubbed but not implemented in the service query, M2-02 will fail silently (returning all records). TS-38-01 will surface this immediately. |
| **Invoice `date_field` default** | M2-03 | Ticket does not clearly specify the default for `date_field` when the param is omitted. TS-39-03 documents the test but leaves the expected default open — confirm with the implementing developer and update the test. |
| **M1 defects D-01, D-03, D-04 must be resolved before M2 sign-off** | M1 carryover | D-01 (`POST /api/expenses` missing date → 400 instead of defaulting to today), D-03 (`PUT /api/invoices/:id` requires status field), and D-04 (no `per_page` cap) from the M1 verification run should be fixed before M2 sign-off. Attachment cascade delete depends on clean expense/income/invoice delete behavior. |
| **MinIO bucket existence** | E7-S1 | The `m2-minio` environment assumes the bucket named in `MINIO_BUCKET` already exists. Verify whether the server auto-creates the bucket on startup (`storage.NewObjectStore`) or whether it must be created manually via the MinIO console before the first upload. Document in the `.env.example`. |
| **Attachment list performance** | E7-S1 | No performance AC is defined for `GET /api/attachments?entity_type=expense&entity_id=<id>`. For M2 sign-off, verify the `idx_attachments_entity` index is used: `sqlite3 moneyapp.db "EXPLAIN QUERY PLAN SELECT * FROM attachments WHERE entity_type='expense' AND entity_id=1"`. Expect to see `SEARCH attachments USING INDEX idx_attachments_entity`. |

---

*End of Milestone 2 Test Plan — Draft*
