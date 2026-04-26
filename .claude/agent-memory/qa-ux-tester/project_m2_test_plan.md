---
name: M2 Test Plan — Coverage & Structure
description: Milestone 2 test plan summary — file attachments, custom categories, enhanced lists; 85 cases across TS-25 to TS-40
type: project
---

M2 test plan written at `docs/test-plans/milestone-2.md` (817 lines, 2026-04-26 draft).

**Why:** M2 covers the file attachment system (E7-S1 through E7-S5), entity-specific attachment UIs (E2-S7, E3-S5, E4-S5), overdue admin endpoint (E4-S7), custom category CRUD (E5-S2 through E5-S4), and enhanced list features (M2-01 through M2-04).

**How to apply:** Use TS-25 through TS-40 suite IDs when referencing M2 test cases. Both `STORAGE_TYPE=local` and `STORAGE_TYPE=s3` (MinIO) environments must be tested for all attachment tests. Run TS-29 cascade-delete tests in both environments.

Key notes for future runs:
- M1 defects D-01 (expense date default), D-03 (invoice PUT requires status), D-04 (no per_page cap) must be resolved before M2 sign-off — cascade delete depends on clean parent delete behavior.
- `entity_id` existence validation is not covered by AC — add once behavior confirmed.
- MIME sniffing (magic bytes) is not required by ticket; document as known limitation.
- `invoice/check-overdue` routing must not conflict with `/api/invoices/stats` — verify method-prefixed mux patterns.
- Traceability: 16 tickets → 85 total test cases (TS-25-01 through TS-40-03).
- Automation backlog: Playwright file upload tests (page.setInputFiles), Go integration tests for AttachmentService and CategoryService.Delete.
