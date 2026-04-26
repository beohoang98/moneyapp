# MoneyApp — Product Requirements Document

**Version**: 1.1
**Date**: 2026-04-26
**Status**: Draft

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Scope & Assumptions](#2-scope--assumptions)
3. [Functional Requirements](#3-functional-requirements)
   - 3.1 [Authentication & User Management](#31-authentication--user-management)
   - 3.2 [Expense Tracking](#32-expense-tracking)
   - 3.3 [Income Management](#33-income-management)
   - 3.4 [Invoice & Bill Management](#34-invoice--bill-management)
   - 3.5 [Categories & Tags](#35-categories--tags)
   - 3.6 [Financial Dashboard & Reports](#36-financial-dashboard--reports)
   - 3.7 [File Attachments](#37-file-attachments)
   - 3.8 [Document Scanning / OCR](#38-document-scanning--ocr)
4. [Non-Functional Requirements](#4-non-functional-requirements)
5. [Data Model Overview](#5-data-model-overview)
6. [Epics & User Stories](#6-epics--user-stories)
7. [Milestones & Timeline](#7-milestones--timeline)
8. [Dependencies & Risks](#8-dependencies--risks)
9. [Open Questions](#9-open-questions)
10. [Definition of Done](#10-definition-of-done)

---

## 1. Executive Summary

MoneyApp is a single-user personal finance web application. It allows the user to record and categorize expenses and incomes, manage recurring and one-time invoices/bills with attached files, and gain a financial overview through summaries and basic reports.

The application is built for a solo developer or small team. Requirements are scoped to be achievable incrementally, with a clear MVP that delivers core value quickly and optional enhancements deferred to later milestones.

**Primary user**: A single individual managing their own finances (no multi-tenancy in MVP).

---

## 2. Scope & Assumptions

### In Scope

- CRUD operations for expenses, incomes, and invoices/bills
- File upload and storage for receipts and invoice documents
- Category management for transactions
- A financial dashboard with summaries and basic filtering
- REST API backend in Go with a React TypeScript frontend
- SQLite as the primary database; MinIO for file storage

### Out of Scope (deferred)

- Multi-user or household sharing
- Bank account sync / Open Banking integrations
- Mobile native applications (PWA-friendly design is acceptable)
- Currency exchange / multi-currency support
- Tax reporting or accountant exports
- Automated recurring transaction creation (may be added post-MVP)
- Budget planning / goal setting (post-MVP)

### Assumptions

- A1: The app is self-hosted by the user; a single login credential is sufficient for MVP.
- A2: SQLite is adequate for the expected data volume (single user, several years of records).
- A3: MinIO is available locally via Docker Compose; object keys do not need to be publicly accessible.
- A4: The user accesses the app from a desktop browser; mobile responsiveness is a should-have.
- A5: Monetary amounts are stored and displayed in a single configured currency (e.g., VND or USD).
- A6: "Invoice" and "bill" are used interchangeably — both refer to a payable document from a vendor.

---

## 3. Functional Requirements

Priority notation follows MoSCoW:
- **[M]** Must-have — required for MVP
- **[S]** Should-have — high value, target Milestone 2
- **[C]** Could-have — nice to have, Milestone 3+
- **[W]** Won't-have — explicitly deferred

---

### 3.1 Authentication & User Management

| ID    | Priority | Requirement                                                                 |
|-------|----------|-----------------------------------------------------------------------------|
| AU-01 | [M]      | The system must allow the user to log in with a username and password.      |
| AU-02 | [M]      | The system must issue a JWT (or session token) upon successful login.       |
| AU-03 | [M]      | All API endpoints except the login endpoint must require a valid token.     |
| AU-04 | [M]      | The system must allow the user to log out, invalidating their session.      |
| AU-05 | [S]      | Tokens must expire after a configurable idle period (default: 24 hours).   |
| AU-06 | [S]      | The system should allow the user to change their password.                  |
| AU-07 | [C]      | The system could support a "remember me" option extending token lifetime.   |
| AU-08 | [W]      | User registration / multi-user accounts are explicitly out of scope.        |

**Acceptance Criteria — AU-01:**
- Given valid credentials, when the user submits the login form, then the system returns a token and redirects to the dashboard.
- Given invalid credentials, when the user submits the login form, then the system returns HTTP 401 and displays an error message.
- Given a blank username or password, when the user submits the form, then the form validates client-side before submission.

---

### 3.2 Expense Tracking

| ID    | Priority | Requirement                                                                              |
|-------|----------|------------------------------------------------------------------------------------------|
| EX-01 | [M]      | The user must be able to create an expense record with: amount, date, category, and optional description. |
| EX-02 | [M]      | The user must be able to view a paginated list of all expenses.                          |
| EX-03 | [M]      | The user must be able to edit any field of an existing expense.                          |
| EX-04 | [M]      | The user must be able to delete an expense record.                                       |
| EX-05 | [M]      | Expense list must be filterable by date range.                                           |
| EX-06 | [S]      | Expense list should be filterable by category.                                           |
| EX-07 | [S]      | The user should be able to attach one or more receipts (image or PDF) to an expense.     |
| EX-08 | [S]      | The expense list should be sortable by date or amount.                                   |
| EX-09 | [S]      | The system should display a running total for the currently filtered view.               |
| EX-10 | [C]      | The user could add free-text tags to an expense in addition to a category.               |
| EX-11 | [C]      | The system could warn if an expense amount is unusually high (configurable threshold).   |

**Acceptance Criteria — EX-01:**
- Given a valid amount (> 0), date, and category, when the user saves the expense, then it appears immediately in the expense list and is persisted to the database.
- Given a negative or zero amount, when the user tries to save, then validation rejects it with a descriptive error.
- Given a missing date, when the user tries to save, then the date defaults to today.

**Acceptance Criteria — EX-04:**
- Given an existing expense, when the user confirms deletion, then the record and any associated file attachments are removed from both the database and MinIO.
- Given a delete action, when the user is shown a confirmation dialog and cancels, then no data is modified.

---

### 3.3 Income Management

| ID    | Priority | Requirement                                                                              |
|-------|----------|------------------------------------------------------------------------------------------|
| IN-01 | [M]      | The user must be able to create an income record with: amount, date, source/category, and optional description. |
| IN-02 | [M]      | The user must be able to view a paginated list of all income records.                    |
| IN-03 | [M]      | The user must be able to edit any field of an existing income record.                    |
| IN-04 | [M]      | The user must be able to delete an income record.                                        |
| IN-05 | [M]      | Income list must be filterable by date range.                                            |
| IN-06 | [S]      | Income list should be filterable by source/category.                                     |
| IN-07 | [S]      | The user should be able to attach supporting documents (e.g., pay slip) to an income record. |
| IN-08 | [S]      | The system should display a total income for the currently filtered period.              |

**Acceptance Criteria — IN-01:**
- Given a valid amount and date, when the user saves the income record, then it appears in the income list and is reflected in the dashboard totals.
- Given an amount of zero, when the user tries to save, then the system rejects it with an error message.

---

### 3.4 Invoice & Bill Management

| ID    | Priority | Requirement                                                                              |
|-------|----------|------------------------------------------------------------------------------------------|
| IV-01 | [M]      | The user must be able to create an invoice/bill record with: vendor name, amount, issue date, due date, status (unpaid / paid / overdue), and optional description. |
| IV-02 | [M]      | The user must be able to view a paginated list of invoices/bills.                        |
| IV-03 | [M]      | The user must be able to update the status of an invoice (e.g., mark as paid).          |
| IV-04 | [M]      | The user must be able to edit all fields of an invoice record.                           |
| IV-05 | [M]      | The user must be able to delete an invoice record.                                       |
| IV-06 | [M]      | The user must be able to attach at least one PDF or image file to an invoice.            |
| IV-07 | [M]      | Invoice list must be filterable by status (unpaid, paid, overdue).                       |
| IV-08 | [S]      | The system should automatically set status to "overdue" when due date has passed and status is still "unpaid". |
| IV-09 | [S]      | Invoice list should be filterable by date range (issue date or due date).               |
| IV-10 | [S]      | The user should be able to view a preview of an attached PDF or image inline.            |
| IV-11 | [S]      | The system should display the total outstanding (unpaid + overdue) amount.               |
| IV-12 | [C]      | The system could send a browser notification when an invoice is approaching its due date (within 3 days). |
| IV-13 | [C]      | The user could mark an invoice as recurring (monthly / quarterly / yearly).              |
| IV-14 | [W]      | Automated recurring invoice generation is out of scope for MVP.                         |

**Acceptance Criteria — IV-01:**
- Given all required fields (vendor, amount, issue date, due date, status), when the user saves the invoice, then it appears in the invoice list under the correct status tab.
- Given a missing vendor name or amount, when the user tries to save, then validation rejects the form with descriptive field-level errors.
- Given a due date earlier than the issue date, when the user tries to save, then the system rejects it with an error message.

**Acceptance Criteria — IV-03:**
- Given an unpaid invoice, when the user marks it as paid, then the status updates immediately in the list and the outstanding total decreases accordingly.

**Acceptance Criteria — IV-08:**
- Given an invoice with status "unpaid" whose due date is in the past, when the page loads (or a scheduled check runs), then the system displays it with status "overdue".
- The transition to "overdue" must not require manual user action.

---

### 3.5 Categories & Tags

| ID    | Priority | Requirement                                                                              |
|-------|----------|------------------------------------------------------------------------------------------|
| CT-01 | [M]      | The system must ship with a set of default expense categories (e.g., Food, Transport, Housing, Health, Entertainment, Shopping, Utilities, Other). |
| CT-02 | [M]      | The system must ship with a set of default income sources (e.g., Salary, Freelance, Investment, Gift, Other). |
| CT-03 | [S]      | The user should be able to create custom categories for both expenses and income.        |
| CT-04 | [S]      | The user should be able to rename or delete custom categories.                           |
| CT-05 | [S]      | Deleting a category must reassign its transactions to a default "Uncategorized" category, not delete them. |
| CT-06 | [C]      | The user could assign a color or icon to a category.                                     |

**Acceptance Criteria — CT-05:**
- Given a custom category with associated transactions, when the user deletes the category, then all transactions previously in that category are moved to "Uncategorized" and no transaction data is lost.

---

### 3.6 Financial Dashboard & Reports

| ID    | Priority | Requirement                                                                              |
|-------|----------|------------------------------------------------------------------------------------------|
| DB-01 | [M]      | The dashboard must display total income, total expenses, and net balance for the current month. |
| DB-02 | [M]      | The dashboard must display a count and total amount of unpaid/overdue invoices.          |
| DB-03 | [M]      | The dashboard must allow the user to switch the summary period (current month, last month, current year, custom date range). |
| DB-04 | [S]      | The dashboard should include a bar or line chart showing income vs. expenses over the selected period broken down by month. |
| DB-05 | [S]      | The dashboard should include a pie or donut chart showing expense breakdown by category. |
| DB-06 | [S]      | The user should be able to export the transaction list (filtered) as a CSV file.        |
| DB-07 | [C]      | The dashboard could show a month-over-month trend indicator (up/down arrow) for expenses. |
| DB-08 | [C]      | The user could export a summary report as PDF.                                           |

**Acceptance Criteria — DB-01:**
- Given the current month, when the dashboard loads, then it shows three summary cards: Total Income, Total Expenses, and Net Balance (Income minus Expenses), all computed from records dated within the current calendar month.
- Given a custom date range, when the user applies it, then all dashboard metrics recalculate within 2 seconds.

**Acceptance Criteria — DB-06:**
- Given a filtered transaction list, when the user clicks Export CSV, then the browser downloads a UTF-8 encoded CSV with columns: date, type, category, description, amount.

---

### 3.7 File Attachments

| ID    | Priority | Requirement                                                                              |
|-------|----------|------------------------------------------------------------------------------------------|
| FA-01 | [M]      | The system must accept file uploads in PDF, JPEG, and PNG formats.                      |
| FA-02 | [M]      | The system must enforce a maximum file size of 10 MB per file.                          |
| FA-03 | [M]      | Uploaded files must be stored in MinIO and referenced by a key in the database.         |
| FA-04 | [M]      | The user must be able to download any attached file.                                     |
| FA-05 | [M]      | Deleting a parent record (expense, income, invoice) must also delete its attached files from MinIO. |
| FA-06 | [S]      | The user should be able to attach multiple files to a single record.                    |
| FA-07 | [S]      | The system should display an inline preview for image attachments (thumbnail).          |
| FA-08 | [S]      | The system should display an inline PDF preview using the browser's built-in PDF viewer.|
| FA-09 | [C]      | The system could show storage usage (total MB used) in a settings page.                 |

**Acceptance Criteria — FA-02:**
- Given a file larger than 10 MB, when the user attempts to upload, then the system rejects it client-side before any network request is made, with a clear size-limit error message.

**Acceptance Criteria — FA-05:**
- Given a record with two attached files, when the user deletes the record, then both objects are removed from MinIO storage (verified by the absence of those keys in the bucket).

---

### 3.8 Document Scanning / OCR

The user uploads a photo or scan of a receipt, bill, or invoice; the system extracts structured data from it and pre-fills the relevant form so the user only needs to review and confirm.

The primary processing approach is a vision-capable LLM API (e.g., Claude Vision / GPT-4o Vision) called server-side. This avoids building a custom OCR pipeline and handles the wide variety of receipt formats a personal app will encounter. Raw images are never permanently stored on a third-party service — only the extracted text/JSON result is retained.

| ID    | Priority | Requirement                                                                                              |
|-------|----------|----------------------------------------------------------------------------------------------------------|
| SC-01 | [M]      | The user must be able to upload an image (JPEG, PNG, WEBP, HEIC) from their file system to trigger a scan. |
| SC-02 | [M]      | The system must send the uploaded image to a server-side vision API and return extracted data as structured JSON. |
| SC-03 | [M]      | The system must extract at minimum: total amount, currency, vendor/merchant name, and transaction date from the image. |
| SC-04 | [M]      | Extracted data must be presented in a pre-filled review form before any record is saved, allowing the user to correct every field. |
| SC-05 | [M]      | The system must handle unreadable or low-confidence results gracefully — if extraction fails or the image is not a financial document, the user is informed and the form is left blank for manual entry. |
| SC-06 | [M]      | A scan attempt must not automatically create any expense, income, or invoice record — the user must explicitly submit the reviewed form. |
| SC-07 | [S]      | The system should extract individual line items (description + amount) from itemized receipts and display them in the review UI. |
| SC-08 | [S]      | The system should allow the user to choose the record type (expense / invoice) before or after scanning. |
| SC-09 | [S]      | The system should display a confidence indicator or highlight fields where the extraction was uncertain so the user knows what to check. |
| SC-10 | [S]      | The uploaded image should automatically be attached to the resulting record (reusing the FA-* attachment flow) so the user has the original on file. |
| SC-11 | [S]      | Scan processing must complete and return a result within 15 seconds under normal network conditions. |
| SC-12 | [C]      | The system could allow the user to capture an image directly from the device camera (via browser `<input capture>`) rather than only selecting from a file. |
| SC-13 | [C]      | The system could persist scan results (extracted JSON + source image key) in a `scan_results` table for audit and re-processing purposes. |
| SC-14 | [C]      | The system could suggest a category based on the vendor name using simple keyword matching. |
| SC-15 | [W]      | On-device OCR (e.g., Tesseract.js in the browser) is not in scope — server-side vision API is the chosen approach. |
| SC-16 | [W]      | Batch scanning of multiple receipts in a single operation is not in scope for MVP. |

**Acceptance Criteria — SC-02 / SC-03:**
- Given a clear photo of a receipt, when the user submits it for scanning, then the system returns a JSON object containing at minimum: `amount` (integer, minor units), `currency`, `vendor_name`, and `date` within 15 seconds.
- Given a blurry or non-receipt image, when the system cannot extract meaningful data, then the API returns a structured error response and the frontend displays a human-readable message (e.g., "We couldn't read this image. Please try a clearer photo or enter details manually.").

**Acceptance Criteria — SC-04:**
- Given a successful scan result, when the review form is displayed, then every extracted field is editable and the user can change any value before saving.
- Given a partially extracted result (e.g., amount found but date missing), when the form is displayed, then blank fields are clearly indicated and the user can fill them in manually.

**Acceptance Criteria — SC-06:**
- Given a completed scan and filled review form, when the user does NOT click save, then no record is written to the database and no attachment is stored in MinIO.

---

## 4. Non-Functional Requirements

### 4.1 Performance

| ID     | Requirement                                                                                   |
|--------|-----------------------------------------------------------------------------------------------|
| NF-01  | API responses for list endpoints must complete within 500 ms for up to 10,000 records under local network conditions. |
| NF-02  | The dashboard summary page must load and render within 3 seconds on initial page load.        |
| NF-03  | File uploads up to 10 MB must complete without timeout under a local network.                 |
| NF-04  | SQLite queries on the main transaction tables must use indexed columns for date and category.  |

### 4.2 Security

| ID     | Requirement                                                                                   |
|--------|-----------------------------------------------------------------------------------------------|
| NF-05  | All API communication must occur over HTTPS in production.                                    |
| NF-06  | JWT secrets and MinIO credentials must be loaded from environment variables, never hardcoded. |
| NF-07  | File download URLs must be pre-signed or access-controlled — raw MinIO bucket objects must not be publicly accessible. |
| NF-08  | User passwords must be hashed using bcrypt (cost factor >= 12) before storage.                |
| NF-09  | The API must return appropriate HTTP error codes and must not leak internal error details to the client. |

### 4.3 Reliability & Data Integrity

| ID     | Requirement                                                                                   |
|--------|-----------------------------------------------------------------------------------------------|
| NF-10  | All monetary amounts must be stored as integers (minor currency units, e.g., cents) to avoid floating-point errors. |
| NF-11  | Database writes must use transactions where multiple tables are affected (e.g., record + attachments). |
| NF-12  | The application must handle MinIO unavailability gracefully — file operations must fail with a user-friendly error without corrupting database state. |
| NF-13  | The SQLite database file must be backed up by copying the file; the app should expose a `GET /api/backup` endpoint (authenticated) that streams the DB file. |

### 4.4 Usability

| ID     | Requirement                                                                                   |
|--------|-----------------------------------------------------------------------------------------------|
| NF-14  | All forms must provide real-time client-side validation with field-level error messages.      |
| NF-15  | Destructive actions (delete, status change) must require explicit confirmation.               |
| NF-16  | The UI must be responsive for screens >= 768 px (tablet and desktop).                        |
| NF-17  | The application must display currency amounts in a locale-consistent format.                  |

### 4.5 Document Scanning / OCR

| ID     | Requirement                                                                                   |
|--------|-----------------------------------------------------------------------------------------------|
| NF-22  | Vision API calls must be made server-side only; the API key must never be exposed to the browser. |
| NF-23  | Images submitted for scanning must not be stored permanently on the vision API provider's infrastructure — the request must use a stateless, non-training API tier where available. |
| NF-24  | Scan processing (upload to backend + vision API round-trip + response to client) must complete within 15 seconds under a local or residential network connection. |
| NF-25  | The backend must validate the uploaded image MIME type and file size (max 10 MB, same as FA-02) before forwarding it to the vision API. |
| NF-26  | Vision API credentials must be loaded from environment variables and documented in `.env.example`. |
| NF-27  | If the vision API is unavailable or returns an error, the system must degrade gracefully — the user can still enter data manually without any application crash. |

### 4.6 Maintainability

| ID     | Requirement                                                                                   |
|--------|-----------------------------------------------------------------------------------------------|
| NF-18  | The backend must expose a `GET /api/health` endpoint returning service and database status.   |
| NF-19  | Database schema changes must be applied via versioned migration files (not ad-hoc ALTER TABLE). |
| NF-20  | Backend code must have unit test coverage for service-layer business logic.                   |
| NF-21  | Environment configuration (port, DB path, MinIO endpoint/credentials, JWT secret) must be documented in a `.env.example` file. |

---

## 5. Data Model Overview

This is a conceptual overview — exact schema is defined in migration files.

### Entities

**users**
- id, username, password_hash, created_at

**categories**
- id, name, type (expense | income), is_default (bool), color (optional), created_at

**expenses**
- id, amount (integer, minor units), date, category_id (FK), description, created_at, updated_at

**incomes**
- id, amount (integer, minor units), date, category_id (FK), description, created_at, updated_at

**invoices**
- id, vendor_name, amount (integer, minor units), issue_date, due_date, status (unpaid | paid | overdue), description, created_at, updated_at

**attachments**
- id, entity_type (expense | income | invoice), entity_id, filename, mime_type, size_bytes, storage_key (MinIO object key), created_at

**scan_results** *(could-have — SC-13; only needed if scan audit/history is implemented)*
- id, source_image_key (MinIO object key of the uploaded image), raw_response (JSON text returned by the vision API), extracted_amount (integer, minor units, nullable), extracted_currency (nullable), extracted_vendor (nullable), extracted_date (nullable), extracted_line_items (JSON array, nullable), status (success | partial | failed), created_at
- Note: if SC-13 is deferred, the backend can process scans statelessly without writing to this table; the table only becomes necessary for audit or re-processing use cases.

### Key Relationships

- Each expense/income belongs to one category.
- Each attachment belongs to one entity (polymorphic via entity_type + entity_id).
- Deleting an expense/income/invoice cascades to its attachments (in both DB and MinIO).
- A scan_results row is optionally linked to the attachment created from the source image (foreign key to attachments.id, nullable — the image may be discarded if the user does not save the record).

---

## 6. Epics & User Stories

### Epic 1 — Authentication

**Goal**: Secure access to the application.

| Story ID | User Story                                                                 | Size | Dependencies |
|----------|----------------------------------------------------------------------------|------|--------------|
| E1-S1    | As the user, I want to log in with a username and password so that only I can access my financial data. | S    | —            |
| E1-S2    | As the user, I want to be automatically logged out after 24 hours of inactivity so that my data stays secure. | S    | E1-S1        |
| E1-S3    | As the user, I want to manually log out from the app so that I can secure my session on shared computers. | XS   | E1-S1        |
| E1-S4    | As the user, I want to change my password from a settings page so that I can maintain account security. | S    | E1-S1        |

---

### Epic 2 — Expense Tracking

**Goal**: Record, view, and manage all outgoing transactions.

| Story ID | User Story                                                                                                          | Size | Dependencies |
|----------|---------------------------------------------------------------------------------------------------------------------|------|--------------|
| E2-S1    | As the user, I want to add an expense with amount, date, category, and description so that I have a record of my spending. | S    | Epic 5 (categories) |
| E2-S2    | As the user, I want to view all my expenses in a paginated list so that I can browse my spending history.           | S    | E2-S1        |
| E2-S3    | As the user, I want to filter expenses by date range so that I can focus on a specific period.                      | S    | E2-S2        |
| E2-S4    | As the user, I want to filter expenses by category so that I can see how much I spent in a particular area.         | S    | E2-S2        |
| E2-S5    | As the user, I want to edit an existing expense so that I can correct mistakes.                                     | XS   | E2-S1        |
| E2-S6    | As the user, I want to delete an expense so that I can remove erroneous records.                                    | XS   | E2-S1        |
| E2-S7    | As the user, I want to attach a receipt image or PDF to an expense so that I have proof of the transaction.         | M    | E2-S1, Epic 7 (files) |
| E2-S8    | As the user, I want to see the total amount for my current filtered view so that I know my spending at a glance.    | XS   | E2-S3        |

---

### Epic 3 — Income Management

**Goal**: Record and track all incoming money.

| Story ID | User Story                                                                                             | Size | Dependencies |
|----------|--------------------------------------------------------------------------------------------------------|------|--------------|
| E3-S1    | As the user, I want to add an income record with amount, date, source, and description so that I can track what I earn. | S    | Epic 5       |
| E3-S2    | As the user, I want to view all income records in a paginated list so that I have a clear earnings history. | S    | E3-S1        |
| E3-S3    | As the user, I want to filter income by date range so that I can analyze earnings for a period.        | S    | E3-S2        |
| E3-S4    | As the user, I want to edit or delete an income record so that I can keep my data accurate.            | XS   | E3-S1        |
| E3-S5    | As the user, I want to attach a pay slip or document to an income record so that I have supporting evidence. | M    | E3-S1, Epic 7 |

---

### Epic 4 — Invoice & Bill Management

**Goal**: Manage payable documents and track payment status.

| Story ID | User Story                                                                                                       | Size | Dependencies |
|----------|------------------------------------------------------------------------------------------------------------------|------|--------------|
| E4-S1    | As the user, I want to create an invoice record with vendor, amount, issue date, due date, and status so that I can track my bills. | M    | —            |
| E4-S2    | As the user, I want to view all invoices in a list, filterable by status, so that I know what is unpaid.         | S    | E4-S1        |
| E4-S3    | As the user, I want to mark an invoice as paid so that my outstanding balance stays accurate.                    | XS   | E4-S1        |
| E4-S4    | As the user, I want to see overdue invoices highlighted so that I can act on them promptly.                      | S    | E4-S1        |
| E4-S5    | As the user, I want to attach the actual invoice PDF to the record so that I can retrieve the original document. | M    | E4-S1, Epic 7 |
| E4-S6    | As the user, I want to edit or delete an invoice record so that I can correct mistakes or remove duplicates.     | XS   | E4-S1        |
| E4-S7    | As the user, I want the app to automatically flag an invoice as overdue when its due date has passed so that I don't miss payments. | S    | E4-S1        |
| E4-S8    | As the user, I want to see the total outstanding amount for unpaid and overdue invoices so that I know what I owe. | XS   | E4-S2        |

---

### Epic 5 — Categories

**Goal**: Organize transactions for meaningful reporting.

| Story ID | User Story                                                                                              | Size | Dependencies |
|----------|---------------------------------------------------------------------------------------------------------|------|--------------|
| E5-S1    | As the user, I want the app to have sensible default categories pre-loaded so that I can start recording immediately. | XS   | —            |
| E5-S2    | As the user, I want to create custom categories so that I can tailor categorization to my lifestyle.    | S    | E5-S1        |
| E5-S3    | As the user, I want to rename or delete custom categories so that I can keep my list tidy.              | S    | E5-S2        |
| E5-S4    | As the user, I want transactions from a deleted category to move to "Uncategorized" so that I never lose data. | S    | E5-S3        |

---

### Epic 6 — Dashboard & Reports

**Goal**: Provide a financial overview and actionable insights.

| Story ID | User Story                                                                                                   | Size | Dependencies |
|----------|--------------------------------------------------------------------------------------------------------------|------|--------------|
| E6-S1    | As the user, I want to see my total income, total expenses, and net balance for the current month so that I know my financial position at a glance. | M    | Epics 2, 3   |
| E6-S2    | As the user, I want to switch the dashboard period (current month / last month / this year / custom range) so that I can compare different periods. | S    | E6-S1        |
| E6-S3    | As the user, I want a chart showing income vs. expenses by month so that I can spot trends over time.        | M    | E6-S1        |
| E6-S4    | As the user, I want a chart showing my expense breakdown by category so that I know where my money goes.     | M    | E6-S1        |
| E6-S5    | As the user, I want to export my filtered transaction list as a CSV file so that I can use the data in a spreadsheet. | L    | Epics 2, 3   |
| E6-S6    | As the user, I want to see a count and total of unpaid/overdue invoices on the dashboard so that outstanding bills are always visible. | S    | Epic 4       |

---

### Epic 7 — File Attachments

**Goal**: Store and retrieve supporting documents.

| Story ID | User Story                                                                                                 | Size | Dependencies |
|----------|------------------------------------------------------------------------------------------------------------|------|--------------|
| E7-S1    | As the user, I want to upload PDF, JPEG, or PNG files to a record so that I can keep supporting documents alongside transactions. | L    | MinIO setup  |
| E7-S2    | As the user, I want the app to reject files over 10 MB so that I am warned before wasting bandwidth.       | XS   | E7-S1        |
| E7-S3    | As the user, I want to download any attached file so that I can retrieve the original document.            | S    | E7-S1        |
| E7-S4    | As the user, I want to see a thumbnail preview of image attachments so that I can confirm I uploaded the right file. | S    | E7-S1        |
| E7-S5    | As the user, I want attached files to be deleted when I delete a record so that I don't accumulate orphaned files in storage. | S    | E7-S1        |

---

### Epic 8 — Document Scanning

**Goal**: Let the user photograph a receipt, bill, or invoice and have key fields extracted automatically, reducing manual data entry.

| Story ID | User Story                                                                                                                          | Size | Dependencies              |
|----------|-------------------------------------------------------------------------------------------------------------------------------------|------|---------------------------|
| E8-S1    | As the user, I want to upload a receipt or invoice image and have the system extract the amount, vendor, and date automatically so that I don't have to type everything by hand. | L    | Epic 7 (file upload), Vision API configured |
| E8-S2    | As the user, I want to review all extracted fields in a pre-filled form before saving so that I can correct any mistakes the scan made. | M    | E8-S1                     |
| E8-S3    | As the user, I want to be told clearly when a scan fails or the image is unreadable so that I know to enter data manually instead. | S    | E8-S1                     |
| E8-S4    | As the user, I want the scanned image to be automatically attached to the record I save so that I have the original receipt on file without a second upload step. | S    | E8-S1, E8-S2              |
| E8-S5    | As the user, I want the scan to show extracted line items from an itemized receipt so that I can see the breakdown before deciding how to categorize the expense. | M    | E8-S1                     |
| E8-S6    | As the user, I want to choose whether the scanned document becomes an expense or an invoice so that I can handle both receipt types from the same scan flow. | S    | E8-S1, E8-S2              |
| E8-S7    | As the user, I want to see which fields the system is uncertain about so that I know where to focus my review. | M    | E8-S2                     |
| E8-S8    | As the user, I want to use my device camera directly from the app to capture a receipt so that I don't have to transfer photos to my computer first. | S    | E8-S1                     |

**Notes on sizing:**
- E8-S1 is L because it requires a new backend endpoint, vision API integration, structured JSON prompt engineering, error handling, and frontend upload flow.
- E8-S2 is M because the review form is a new UI component that mirrors the expense/invoice form but is pre-populated and must handle nullable/partial data.
- E8-S5 is M because line item rendering requires a dynamic list component and the prompt must reliably return itemized data.

---

## 7. Milestones & Timeline

Estimates assume a solo developer working part-time (~10–15 hours/week). Adjust multipliers for a full-time developer or a pair.

### Milestone 0 — Project Foundation (Week 1–2)

**Goal**: Runnable skeleton, no features yet.

- [ ] Initialize Go module, folder structure (`cmd/`, `internal/`, `migrations/`)
- [ ] Initialize React + TypeScript app (Vite), ESLint, Prettier
- [ ] Docker Compose with MinIO service
- [ ] Database migration runner (e.g., `golang-migrate`)
- [ ] `.env.example` with all required variables
- [ ] `GET /api/health` endpoint
- [ ] CI pipeline: lint + build checks (GitHub Actions or equivalent)

**Exit criteria**: `go run ./cmd/server` and `npm run dev` start without errors; MinIO is reachable; health endpoint returns 200.

---

### Milestone 1 — MVP Core (Week 3–8)

**Goal**: Fully functional core with no charts or file attachments. Delivers the most essential user value.

Epics covered: Auth (full), Categories (defaults only), Expenses (CRUD + basic filter), Income (CRUD + basic filter), Invoices (CRUD + status filter), Dashboard (summary cards only).

Key stories: E1-S1 through E1-S3, E2-S1 through E2-S6, E2-S8, E3-S1 through E3-S4, E4-S1 through E4-S4, E4-S6 through E4-S8, E5-S1, E6-S1, E6-S6.

**Exit criteria**: User can log in, record expenses/income/invoices, filter lists, and see a dashboard summary card. All data persists across restarts.

---

### Milestone 2 — File Attachments & Enhanced Lists (Week 9–13)

**Goal**: Add file storage and improve list UX.

Epics covered: File Attachments (full Epic 7), Expenses E2-S7, Income E3-S5, Invoices E4-S5, Categories E5-S2 through E5-S4, Invoice auto-overdue E4-S7.

**Exit criteria**: User can attach files to any record, preview images inline, download files, and custom categories are manageable.

---

### Milestone 3 — Reports, Export & Document Scanning (Week 14–20)

**Goal**: Add charts, data export, and the receipt scanning / OCR feature.

Epics covered:
- Dashboard E6-S2 through E6-S5 (charts, period switcher, CSV export)
- Document Scanning E8-S1 through E8-S7 (core scan flow: upload, extract, review, save, error handling)

Prerequisite: Milestone 2 (file attachments) must be complete — scanning reuses the attachment upload and storage flow.

Key decisions before starting M3 scanning work:
- Confirm choice of vision API (see OQ7)
- Confirm API key is available and billing is set up
- Add `VISION_API_KEY` and `VISION_API_PROVIDER` to `.env.example`

**Exit criteria**: Dashboard shows income vs. expense chart and category breakdown chart; CSV export works on all filtered views; user can scan a receipt image and have amount/vendor/date pre-filled in a review form before saving.

---

### Milestone 4 — Polish, Camera Capture & Nice-to-Haves (Week 21+)

Remaining could-have items: category colors/icons, due date notifications, password change, storage usage display, PDF export, camera capture from browser (E8-S8), scan audit log (SC-13), category suggestion from vendor name (SC-14), and any deferred backlog items.

---

## 8. Dependencies & Risks

| ID  | Item                             | Type        | Impact | Mitigation                                                                 |
|-----|----------------------------------|-------------|--------|----------------------------------------------------------------------------|
| D1  | MinIO availability during dev    | External    | High   | Docker Compose setup must be documented and easy to start; health check should detect MinIO status. |
| D2  | SQLite concurrency limitations   | Technical   | Low    | Single-user app; concurrent writes are not a concern. Use WAL mode for slightly better performance. |
| D3  | File storage orphan accumulation | Data        | Medium | Implement cleanup in delete handlers and add an optional audit/cleanup admin endpoint. |
| D4  | JWT secret rotation              | Security    | Medium | Document rotation process; existing tokens become invalid on rotation (acceptable for personal app). |
| D5  | Chart library choice (frontend)  | Technical   | Low    | Evaluate Recharts or Chart.js early to avoid late-stage refactoring. Decide before Milestone 3. |
| D6  | Amount precision (floating point)| Data        | High   | Enforce integer storage (minor units) at the model level from day one. |
| D7  | Missing MinIO cleanup on failure | Reliability | Medium | Use database transactions; only write MinIO key to DB after successful upload. Roll back DB if MinIO write fails. |
| D8  | Vision API availability / cost   | External    | Medium | Design the scan endpoint so it degrades gracefully (returns a clear error) if the API is unreachable. Monitor token usage; a personal app should stay well within free-tier or low-cost limits given infrequent scans. |
| D9  | Vision API result quality        | Technical   | Medium | Prompt must request structured JSON output with explicit field names. Implement response validation; if required fields are missing, treat as a partial result rather than a hard failure. |
| D10 | Image data privacy               | Security    | Medium | Image bytes are sent to a third-party API. Document this in a user-facing note. Prefer API tiers that do not use data for model training (see NF-23). Never log image contents server-side. |
| D11 | Scan depends on file attachment Epic | Technical | High  | Epic 8 cannot start until Epic 7 (file attachments, MinIO upload) is functional. Plan M3 accordingly. |

---

## 9. Open Questions

| ID  | Question                                                                                       | Owner     | Priority |
|-----|------------------------------------------------------------------------------------------------|-----------|----------|
| OQ1 | What currency should the app default to? Will the user ever need to switch currencies?         | Developer | High     |
| OQ2 | Should "overdue" status be calculated on every page load, or via a background job?             | Developer | Medium   |
| OQ3 | Is a pre-signed URL approach acceptable for file downloads, or is a proxy download route preferred? | Developer | Medium   |
| OQ4 | Should CSV export include file attachment metadata (filename, size)?                           | Developer | Low      |
| OQ5 | Is password change (E1-S4) required before the app can be considered shippable for personal use? | Developer | Medium   |
| OQ6 | What is the intended deployment environment — local-only, or self-hosted behind a reverse proxy (e.g., Nginx, Traefik)? | Developer | High     |
| OQ7 | Which vision API should be used for OCR/scanning? Options: (a) Claude Vision (Anthropic API — strong structured output, familiar SDK), (b) GPT-4o Vision (OpenAI), (c) Google Cloud Vision (traditional OCR, less flexible on complex layouts), (d) a self-hosted model (Ollama + LLaVA — no cost/privacy concern, but lower accuracy). Recommended default: Claude Vision for structured JSON extraction. | Developer | High     |
| OQ8 | Should scan results be persisted to a `scan_results` table (SC-13) for audit, or processed statelessly with no DB write? Stateless is simpler for MVP; persisting enables re-processing and history. | Developer | Medium   |
| OQ9 | What should happen to the uploaded image after a scan if the user does NOT save the resulting record? Options: (a) delete immediately from MinIO, (b) keep temporarily for 24 hours then purge, (c) always discard (process in memory without writing to MinIO). Option (c) simplifies cleanup but prevents attaching the image to the record. | Developer | Medium   |
| OQ10 | Is the 15-second scan timeout (SC-11, NF-24) acceptable UX, or should a loading state with progress indication be prioritized? Vision API latency is typically 3–8 seconds for a receipt image; 15 s is a hard ceiling, not an expected average. | Developer | Low      |

---

## 10. Definition of Done

A story is considered **Done** when all of the following are true:

- [ ] All acceptance criteria defined for the story pass.
- [ ] Backend: the relevant API endpoint(s) are implemented, return correct HTTP status codes, and validate inputs.
- [ ] Frontend: the UI reflects the feature and handles both success and error states.
- [ ] Data is persisted correctly and survives an application restart.
- [ ] No regression in previously working features (manual smoke test at minimum).
- [ ] Code is committed, with no linting errors.
- [ ] Any new environment variables are added to `.env.example` with a comment.
- [ ] If file storage is involved: MinIO interaction is tested with a real MinIO instance (not mocked).
- [ ] If vision API is involved: the scan endpoint is tested with both a valid receipt image (happy path) and an unreadable image (error path); no record is written if the user does not submit the review form.

A milestone is considered **Done** when:

- [ ] All must-have stories for that milestone are Done.
- [ ] The application can be started from a clean clone using documented steps.
- [ ] `GET /api/health` returns 200 with database connectivity confirmed.
