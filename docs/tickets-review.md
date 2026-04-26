# Ticket Breakdown Review — MoneyApp

**Reviewer**: QA/UX Agent
**Review Date**: 2026-04-26
**Documents Reviewed**:
- `docs/requirements.md` v1.0
- `docs/tickets.md` (draft, generated 2026-04-26)

---

## Summary Assessment

**Overall Quality Grade: B+**

The ticket breakdown is thorough and well-structured. Infrastructure tickets are detailed and technically sound. The separation into Milestone 0 foundation tickets is excellent practice. The API design is largely RESTful and internally consistent. Most Must-have requirements have corresponding tickets.

However, there are notable gaps in acceptance criteria quality, several dependency errors, a routing conflict that will cause a bug in production, missing tickets for must-have non-functional requirements, and UX state coverage is uneven across tickets. These issues should be resolved before development begins, particularly the routing conflict and the missing transaction rollback specification in the attachment delete path.

---

## Issues Found

### Critical

---

**[CRITICAL] [DEPENDENCY / BUG] API routing conflict between `/api/invoices/summary` and `/api/invoices/:id`**

- Ticket(s): E4-S8, E4-S6
- Description: The API route table lists both `GET /api/invoices/summary` and `GET /api/invoices/:id`. With Go's standard `net/http` mux (and most routers that do prefix matching), the path `/api/invoices/summary` will be captured by the `/:id` handler with `id = "summary"` unless the router has explicit literal-segment priority. The ticket does not specify which router library is used, nor does it address this conflict.
- Impact: `GET /api/invoices/summary` will return a 404 (not found for id="summary") or a 400 (failed int parse) in production. This blocks E4-S8 entirely.
- Recommendation: Either (a) rename the summary endpoint to `GET /api/invoices/stats` or `GET /api/dashboard/invoice-summary` so there is no path collision, or (b) explicitly document that the chosen router (e.g., `chi`, `gorilla/mux`) gives literal path segments priority over parameter segments, and add a test case that verifies the route resolution order. Option (a) is simpler and safer.

---

**[CRITICAL] [COMPLETENESS] NF-13 database backup endpoint is placed in M4 but requirements mark it as a must-have non-functional requirement**

- Ticket: M4-09
- Description: NF-13 states "the app should expose a `GET /api/backup` endpoint (authenticated) that streams the DB file." The NFR section uses "should" (lower-case English sense, not MoSCoW), but the backup endpoint is the only documented data-recovery mechanism for a self-hosted app. Deferring it to M4 (the polish milestone, week 18+) leaves the app without any backup capability for months.
- Impact: If a user's disk fails or the SQLite file is corrupted between M1 and M4, there is no supported recovery path.
- Recommendation: Move M4-09 to M1 or early M2. It is an XS-to-S backend-only ticket with no UI dependencies beyond a settings button. The risk-to-effort ratio is very unfavorable for deferral.

---

**[CRITICAL] [TECHNICAL] E4-S4 and E4-S7 define overlapping, potentially conflicting overdue transition strategies without clear precedence**

- Ticket(s): E4-S4, E4-S7
- Description: E4-S4 (M1) says "In `InvoiceService.List()`, before returning results, check each invoice: if status == 'unpaid' and due_date < today, update its status to 'overdue' in the database." E4-S7 (M2) then claims to "enhance the overdue detection from E4-S4" by adding `UpdateOverdueStatuses()` called on startup and every GET. In practice both tickets implement the same database UPDATE, so the code produced will either duplicate the logic or the M2 ticket will be a no-op. The M1 implementation already satisfies IV-08 fully; the M2 ticket adds no new observable behavior.
- Impact: Developer confusion during M2 — wasted time re-investigating already-done logic; or worse, two separate UPDATE paths running concurrently that interfere.
- Recommendation: Consolidate into a single approach, chosen explicitly in E4-S4. Remove E4-S7 or redefine it as "add on-startup call to UpdateOverdueStatuses and add logging," which is an XS task, not S.

---

### Major

---

**[MAJOR] [DEPENDENCY] E2-S4 (filter by category) is scheduled in M1 but listed as "Should" priority, while its parent story E2-S2 is "Must"**

- Ticket: E2-S4
- Description: The milestone plan for M1 in the requirements doc includes "E2-S4" in the key stories list for Milestone 1. The ticket correctly marks E2-S4 as "Should" priority. This inconsistency means the milestone exit criteria are unclear — is M1 done without category filtering working?
- Impact: Sprint planning ambiguity. A developer could declare M1 complete without E2-S4, which contradicts the requirements milestone plan.
- Recommendation: Either (a) remove E2-S4 from the M1 key story list in the requirements (and confirm the ticket's M1 placement is intentional but optional), or (b) change E2-S4 priority to "Must" if the product owner considers it required for M1. The ticket and requirements doc must agree.

---

**[MAJOR] [COMPLETENESS] No ticket exists for `NF-04` (indexed queries) or `NF-01` (500ms list performance)**

- Description: NF-04 requires that SQLite queries on main transaction tables use indexed columns for date and category. The migration SQL in E2-S1, E3-S1, E4-S1 does create the indexes, but there is no test or acceptance criterion that verifies index usage. NF-01 (500ms under 10,000 records) has no ticket and no acceptance criterion in any ticket.
- Impact: Performance regressions can silently accumulate across milestones with no test gate.
- Recommendation: Add an acceptance criterion to E2-S2, E3-S2, and E4-S2: "Given 10,000 records, when the list endpoint is called, then the response time is under 500ms (local network)." Alternatively add a dedicated performance smoke-test ticket in M2. At minimum, add `EXPLAIN QUERY PLAN` verification to the CI backend-test job.

---

**[MAJOR] [COMPLETENESS] No ticket covers `NF-20` — unit test coverage for service-layer business logic**

- Description: M0-08 sets up CI with `go test ./...` but no ticket defines what must be tested or sets a coverage expectation. NF-20 explicitly requires unit test coverage for service-layer business logic. This means the CI `backend-test` job runs an empty test suite and always passes.
- Impact: The most critical business logic (amount validation, category type enforcement, overdue status transition, cascade delete) ships with zero automated verification.
- Recommendation: Add a ticket (M1 or M2) specifying which service methods require unit tests and a minimum coverage threshold (e.g., 70% for `internal/services/`). Key test cases to mandate: `ExpenseService.Create` rejects zero/negative amounts, `CategoryService.Delete` reassigns transactions, `InvoiceService.UpdateOverdueStatuses` only updates unpaid invoices past due date.

---

**[MAJOR] [DEPENDENCY] E7-S5 (cascade delete) depends on `E2-S6` but not on `E4-S6` or `E3-S4`**

- Ticket: E7-S5
- Description: The cascade delete ticket lists dependencies as `E7-S1, E2-S6`. It must also update `IncomeService.Delete()` and `InvoiceService.Delete()`, which live in E3-S4 and E4-S6 respectively. Without those dependencies declared, a developer who picks up E7-S5 without completing E3-S4 and E4-S6 will find those service methods do not yet exist.
- Impact: Blocked development or incorrect partial implementation (cascade only works for expenses).
- Recommendation: Update E7-S5 dependencies to: `E7-S1, E2-S6, E3-S4, E4-S6`.

---

**[MAJOR] [ACCEPTANCE CRITERIA] E4-S3 "Mark invoice as paid" accepts any status transition but the requirement does not restrict it to unpaid/overdue only at the UI level**

- Ticket: E4-S3
- Description: The backend correctly validates "only unpaid/overdue -> paid." The acceptance criteria only test that clicking "Mark as Paid" on an unpaid invoice works and that the button is hidden for paid invoices. Missing: what happens when the API receives a request to mark an already-paid invoice as paid (HTTP 422? 409? 400?), and what happens if a user manually POSTs a status transition from paid to unpaid? The button-hiding is purely a frontend guard with no server-side test.
- Impact: A crafted API call can corrupt invoice status. No test enforces the server-side restriction.
- Recommendation: Add acceptance criteria: "Given a paid invoice, when `PATCH /api/invoices/:id/status` is called with `{"status": "paid"}`, then the server returns 422 Unprocessable Entity." Also add: "Given `{"status": "unpaid"}` is sent, then the server returns 400 — only `paid` is a valid target status via this endpoint."

---

**[MAJOR] [UX] E2-S6, E3-S4, E4-S6 delete flows do not specify what happens when the record has attachments (before M2)**

- Ticket(s): E2-S6, E3-S4, E4-S6
- Description: In M1, deleting an expense does not yet cascade to attachments (E7-S5 is M2). The tickets acknowledge this ("cascade is handled in M2") but do not specify: should the backend block deletion of records that have attachments in M1, or silently leave orphaned attachment records? If the backend silently orphans them, those orphaned records in M2 will break the cascade logic (since entity_id no longer exists).
- Impact: Data integrity issue between M1 and M2 — orphaned attachment rows with dangling entity_ids after M1 deletes. The `attachments` table is only created in M2 (E7-S1), so M1 deletes won't create orphans in the attachments table, but this means there is no attachment table FK enforcement either, which is worth calling out.
- Recommendation: Add a note to E2-S6/E3-S4/E4-S6 clarifying: "Since the attachments table does not exist in M1, no cascade is needed. In M2, E7-S5 will retroactively add cascade behavior to these delete methods." This removes ambiguity for the developer.

---

**[MAJOR] [API DESIGN] `POST /api/attachments` uses a generic endpoint rather than entity-scoped routes**

- Description: The attachment upload endpoint is `POST /api/attachments` with `entity_type` and `entity_id` in the form body. This is a polymorphic design that requires the client to know and pass the correct entity_type string. A REST-idiomatic design would scope attachments under their parent resource: `POST /api/expenses/:id/attachments`, `POST /api/invoices/:id/attachments`. The current design also requires the client to trust that entity_id corresponds to an existing entity and type — the server must validate this cross-entity reference, which is not called out.
- Impact: (1) The server must validate that `entity_id` refers to an actual expense/income/invoice, but there is no acceptance criterion for this. A user could upload an attachment referencing a non-existent entity_id. (2) The generic approach is harder for API consumers to discover and understand.
- Recommendation: Either (a) switch to entity-scoped routes (`POST /api/expenses/:id/attachments`) which makes entity validation trivial via the route handler, or (b) add an explicit acceptance criterion to E7-S1: "Given an entity_id that does not exist for the given entity_type, then the server returns 404." If option (a) is chosen, the route table and all frontend API calls must be updated.

---

**[MAJOR] [SIZING] E6-S5 (CSV export) is sized L — appears over-estimated**

- Ticket: E6-S5
- Description: The backend implementation uses Go's standard `encoding/csv` package with a streaming approach over a single filtered query. The frontend adds one button component. The total work is comparable to E2-S3 (filter + URL params, sized S) plus a streaming response handler. L sizing implies 8+ hours of work; this feels more like M (3-5 hours).
- Impact: Over-estimated tickets inflate milestone timeline and can cause under-scoping of other work.
- Recommendation: Resize E6-S5 to M. If the L size accounts for UTF-8 edge-case testing across character sets and large-dataset streaming validation, document that explicitly in the ticket description.

---

**[MAJOR] [COMPLETENESS] `E1-S2` is sized S but the work described is less than XS**

- Ticket: E1-S2
- Description: The backend task is "ensure the JWT `exp` claim is set" — which is part of E1-S1's AuthService.Login() implementation, already specified there. The frontend task is "on 401, clear localStorage and redirect" — also already called out in M0-05's client.ts specification. E1-S2 as written adds no new behavior that is not already covered; it is essentially a documentation ticket confirming the expiry behavior.
- Impact: The ticket inflates M1 scope and may cause duplication — a developer implementing E1-S1 will naturally set the `exp` claim, and then E1-S2 will be trivially "done" with no additional code.
- Recommendation: Either (a) merge E1-S2 into E1-S1 as a sub-task (simply add acceptance criteria for the 25-hour-old token scenario to E1-S1), or (b) keep E1-S2 but resize it to XS and redefine its scope as: "Write a failing test for expired token rejection that becomes the acceptance gate."

---

### Minor

---

**[MINOR] [ACCEPTANCE CRITERIA] E2-S8 (running total) accepts the total as part of the list response but does not handle the case where no filter is applied**

- Ticket: E2-S8
- Description: The acceptance criteria test the filtered case only. When no filter is active, the total should represent all expenses. The criteria do not confirm this behavior or specify whether the summary bar is shown when no filter is active.
- Recommendation: Add: "Given no filter is applied, when the expense list loads, then the total represents all expenses in the database."

---

**[MINOR] [ACCEPTANCE CRITERIA] E5-S3 acceptance criteria do not test that renaming a category is reflected in existing transaction data**

- Ticket: E5-S3
- Description: The criteria check that "the new name is reflected in all forms and transaction lists" but this depends on whether the category name is stored denormalized on the transaction or joined at query time. If the frontend caches category names, renaming may not refresh them. The backend joins category name at query time (per E2-S2), so this should be fine — but there is no acceptance criterion that explicitly tests: "Given an expense previously showing category 'Pets', after renaming to 'Pet Care', when the expense list is refreshed, then 'Pet Care' is shown."
- Recommendation: Add that explicit criterion. Also add: "Given the category rename, then the renamed category is immediately available in the expense/income form category dropdowns."

---

**[MINOR] [DEPENDENCY] E6-S3 and E6-S4 are both sized M and depend on E6-S2, but E6-S2 is sized S — the milestone M3 effectively starts with two M tickets blocked on one S ticket**

- Tickets: E6-S2, E6-S3, E6-S4
- Description: This is not a blocking problem but it means M3 has a linear dependency chain (E6-S1 -> E6-S2 -> E6-S3 and E6-S4 in parallel). If E6-S3 and E6-S4 are parallelizable (they are, since they are separate charts with separate endpoints), this should be called out.
- Recommendation: Add a note to E6-S3 and E6-S4 that they can be worked in parallel once E6-S2 is complete. This helps a two-person team split work.

---

**[MINOR] [ACCEPTANCE CRITERIA] M0-07 (Docker Compose) has no acceptance criterion for the frontend service**

- Ticket: M0-07
- Description: The Docker Compose ticket creates a frontend Dockerfile and optionally a dev service, but the acceptance criteria only mention the backend and MinIO. There is no exit criterion confirming the frontend builds and serves successfully within Docker.
- Recommendation: Add: "Given `docker compose up`, when the frontend service starts, then `http://localhost:3000` (or configured port) serves the React app."

---

**[MINOR] [UX] No ticket specifies a toast/notification system for success feedback**

- Description: E2-S6 mentions "show success toast" on delete, implying a toast component exists, but no ticket creates this component. Several other tickets (form submissions, mark-as-paid, etc.) would also benefit from success feedback. The ConfirmDialog component is explicitly ticketed (E2-S6), but the toast system is assumed to exist.
- Recommendation: Add a sub-task in M0-04 or E2-S6 to create a reusable `Toast` / `Notification` component, or explicitly note which ticket is responsible for it.

---

**[MINOR] [UX] No empty state specified for the Categories page (E5-S2)**

- Ticket: E5-S2
- Description: The categories page will always show default categories (seeded in E5-S1), so a "no categories" empty state may not be needed. However, neither section (expense categories, income categories) has an empty state for the custom categories sub-list. If the user has created no custom categories, the UI should communicate this clearly rather than showing a blank area.
- Recommendation: Add: "Given no custom categories have been created, when the user views the categories page, then a prompt is shown: 'No custom categories yet. Click Add to create one.'"

---

**[MINOR] [COMPLETENESS] `NF-17` (locale-consistent currency formatting) has no explicit acceptance criterion in any ticket**

- Description: E2-S8, E3-S1, E6-S1 all reference NF-17 informally (displaying amounts in VND), but no ticket has an acceptance criterion that verifies the formatting — e.g., that 1,500,000 VND renders as "1,500,000 ₫" or the configured format, and not as "15000.00".
- Recommendation: Add a shared acceptance criterion to E2-S2 and E3-S2: "Given an amount of 1,500,000 minor units (VND), when displayed in the list, then it renders as '1,500,000 ₫' (or the locale-appropriate format for the configured currency)."

---

**[MINOR] [COMPLETENESS] `NF-05` (HTTPS in production) has no ticket and no deployment documentation ticket**

- Description: NF-05 requires HTTPS in production but no ticket mentions TLS termination, Nginx/Traefik reverse proxy configuration, or even a note in `.env.example` about the expected production topology. OQ6 in the requirements asks about deployment environment but this open question has no associated ticket.
- Recommendation: Add a documentation task (M4 or M2) to create a `docs/deployment.md` covering: reverse proxy setup, TLS termination, environment variable security, and the SQLite backup strategy. This also closes OQ6.

---

**[MINOR] [API DESIGN] The income list endpoint `GET /api/incomes` is missing `sort_by` and `sort_order` parameters**

- Description: M2-01 adds sorting to expenses (`GET /api/expenses`), but the equivalent sort parameters are not added to income. The requirements (EX-08) only call out sorting for expenses, but the income list has similar UX, and IN-08 mentions displaying totals in a filterable period. A user who wants to find their highest income in a period cannot sort the income list.
- Recommendation: Either (a) add income sorting as part of M2-01 (rename to "Expense and Income list sorting") or (b) add a separate M2-05 ticket for income sorting. This is a suggestion-level gap per the requirements, but it should be a conscious decision.

---

**[MINOR] [DEPENDENCY] M4-09 (database backup) is listed as priority "Should" in the ticket but the appendix ticket summary table also shows "Should"**

- Description: This is consistent, but the requirements list NF-13 under "Reliability & Data Integrity" without a priority marker. The placement in M4 is inconsistent with the other Should-priority items (E1-S4 change password, M4-09 backup, E4-S7 auto-overdue in M2) — E4-S7 is Should-priority but appears in M2, while M4-09 is also Should-priority but appears in M4. The milestone placement is not driven by priority alone, but the inconsistency should be acknowledged.
- Recommendation: Add a comment to the ticket noting the deferral rationale: "Placed in M4 despite Should priority because it has no blocking dependencies and the risk is acceptable for early milestones." This prevents future reviewers from questioning the milestone assignment.

---

### Suggestions

---

**[SUGGESTION] [UX] Consider adding a loading skeleton specification to all list pages, not just the dashboard**

- Description: E6-S1 specifies a "loading skeleton" for the dashboard. E2-S2 specifies "a loading indicator" but does not specify whether this should be a spinner, skeleton rows, or a blank page with a message. Consistency across all list pages (expenses, income, invoices) should be defined.
- Recommendation: Add a shared UX note or a frontend conventions ticket in M0 specifying the loading state pattern to use across all list components (e.g., skeleton rows that match the table structure). This prevents each ticket being implemented with a different loading style.

---

**[SUGGESTION] [UX] The expense/income filter state persistence (URL query params) is not specified for the invoices page**

- Ticket: E4-S2, M2-03
- Description: E2-S3 explicitly calls out URL query param persistence for the expense filter. M2-03 (invoice date range filter) does not mention URL params. The invoice status filter in E4-S2 also does not mention URL persistence.
- Recommendation: Add URL query param persistence to M2-03 and E4-S2 for consistency.

---

**[SUGGESTION] [UX] The "Mark as Paid" confirmation dialog (E4-S3) may be unnecessary friction for a common action**

- Ticket: E4-S3
- Description: NF-15 requires confirmation for destructive actions. Marking an invoice as paid is not destructive — the status can be reversed via the Edit form. Requiring a confirmation dialog for this common action adds unnecessary clicks.
- Recommendation: Consider whether the confirmation dialog is needed for "Mark as Paid." If the status can be corrected through the edit form, removing the confirmation (or making it optional via a settings preference) is better UX. The ticket can note this trade-off explicitly.

---

**[SUGGESTION] [ACCESSIBILITY] No ticket mentions ARIA roles, keyboard navigation, or screen reader support**

- Description: NF-16 mentions responsiveness but there is no accessibility requirement (WCAG 2.1 AA) and no ticket addresses keyboard navigation, focus management in modals, or ARIA attributes on dynamic components (the ConfirmDialog, status badges, chart tooltips).
- Recommendation: Add a requirement or acceptance criterion to the ConfirmDialog ticket (E2-S6) that the dialog is keyboard accessible (Escape to cancel, Enter to confirm, focus trapped inside modal). Add a note to E6-S3 and E6-S4 that charts should have a text-based alternative or accessible tooltip for screen readers.

---

**[SUGGESTION] [API DESIGN] No pagination on `GET /api/categories` endpoint**

- Ticket: E5-S1
- Description: The category list endpoint returns all categories without pagination. For MVP with a fixed set of default categories (9 + 6 = 15), this is fine. However, once custom categories are added in M2, a power user could create many categories. The response structure uses a flat array (not the paginated wrapper format), which would be a breaking change to paginate later.
- Recommendation: Return the category list inside the standard `{"data": [...]}` wrapper from day one (without pagination metadata for now) so the response shape is consistent and pagination can be added later without breaking clients.

---

**[SUGGESTION] [COMPLETENESS] The `GET /api/incomes/:id` route is listed in the route table attributed to E3-S4, but E3-S4 says "Backend tasks: already covered in E3-S1"**

- Description: E3-S1 defines all five REST methods for incomes including GET by ID, but E3-S4's description says "already covered." This is correct but the route table attribution to E3-S4 is misleading — someone looking at E3-S4's ticket will not find a backend task that implements that route.
- Recommendation: Update the route table to attribute `GET /api/incomes/:id` to E3-S1 (where it is actually implemented), same as `PUT /api/incomes/:id` and `DELETE /api/incomes/:id`.

---

## Pre-Development Checklist

The following items should be resolved before development begins on M1:

### Must Fix Before M1 Starts

- [ ] **Resolve the `/api/invoices/summary` vs `/api/invoices/:id` routing conflict.** Rename to `/api/invoices/stats` or move to `/api/dashboard/invoice-summary`, and update the route table and all ticket references.
- [ ] **Move M4-09 (database backup endpoint) to M1 or early M2.** Add it to the M1 key story list or create a standalone M1 infrastructure ticket.
- [ ] **Consolidate E4-S4 and E4-S7 into a single overdue transition strategy.** Choose one implementation approach, document it in E4-S4, and either delete E4-S7 or reduce it to an XS logging enhancement ticket.
- [ ] **Fix E7-S5 dependencies** to include `E3-S4` and `E4-S6` in addition to the existing `E2-S6`.
- [ ] **Clarify M1 scope for E2-S4.** Either remove it from the M1 key story list in requirements, or update the ticket priority from Should to Must and align both documents.

### Should Fix Before M1 Starts

- [ ] **Add a unit test coverage ticket** (M1 or M2) specifying required service-layer test cases and a minimum coverage target for `internal/services/`.
- [ ] **Add server-side acceptance criteria to E4-S3** for invalid status transitions (paid -> paid, paid -> unpaid via the PATCH endpoint).
- [ ] **Add a performance acceptance criterion** to E2-S2, E3-S2, and E4-S2 (500ms under 10,000 records, per NF-01 / NF-04).
- [ ] **Add a reusable Toast/Notification component task** to M0-04 or create a dedicated XS ticket; it is referenced in E2-S6 and implied in other tickets but never explicitly created.
- [ ] **Clarify orphaned attachment behavior** for M1 deletes in E2-S6, E3-S4, E4-S6 (note that the attachments table does not yet exist in M1, so no cascade is needed).
- [ ] **Add URL query param persistence** for invoice status filter (E4-S2) and invoice date range filter (M2-03) to match expense filter behavior.

### Suggested for Before M2 Starts

- [ ] **Decide on attachment route design** (generic `POST /api/attachments` vs entity-scoped `POST /api/expenses/:id/attachments`) and add server-side entity-existence validation as an explicit acceptance criterion to E7-S1.
- [ ] **Resize E6-S5 (CSV export)** from L to M if the L estimate does not account for specific large-dataset or encoding edge cases. Document the rationale in the ticket.
- [ ] **Add NF-17 (currency format) acceptance criteria** to E2-S2 and E3-S2.
- [ ] **Add a deployment documentation task** to M4 to close OQ6 (deployment topology) and address NF-05 (HTTPS).
- [ ] **Add an explicit empty state criterion** to E5-S2 for the custom category sub-list.
- [ ] **Add accessibility (keyboard navigation, ARIA) note** to E2-S6's ConfirmDialog implementation and to chart tickets E6-S3 and E6-S4.

---

## API Route Summary Review

The API route design is clean and largely RESTful. Specific notes:

| Issue | Severity | Detail |
|---|---|---|
| `/api/invoices/summary` conflicts with `/api/invoices/:id` | Critical | See Critical section above |
| `PATCH /api/invoices/:id/status` is correct | Good | Using PATCH for partial update is appropriate |
| `POST /api/auth/logout` returning 200 with no body is correct | Good | Stateless JWT logout is documented |
| `GET /api/export/transactions` uses GET for an export — fine for streaming | Good | Acceptable; some teams prefer POST for export with body params |
| `GET /api/attachments?entity_type=&entity_id=` rather than scoped routes | Minor concern | See Major section on API design |
| `PUT /api/auth/password` for password change | Minor | PATCH is more semantically accurate for a partial user resource update, but PUT is acceptable |
| `/api/incomes` is missing `sort_by`/`sort_order` params | Minor | See Minor section |
| No versioning prefix (e.g., `/api/v1/`) | Suggestion | Acceptable for a single-user self-hosted app; would matter if external consumers existed |
| `GET /api/categories` returns flat array, not paginated wrapper | Suggestion | See Suggestion section — use standard `{"data":[]}` wrapper for consistency |

---

## Appendix: Requirement Coverage Matrix

| Requirement ID | Priority | Covered By Ticket(s) | Status |
|---|---|---|---|
| AU-01 | M | E1-S1 | Covered |
| AU-02 | M | E1-S1 | Covered |
| AU-03 | M | E1-S1 (AuthMiddleware) | Covered |
| AU-04 | M | E1-S3 | Covered |
| AU-05 | S | E1-S2 | Covered |
| AU-06 | S | E1-S4 | Covered |
| AU-07 | C | M4-02 | Covered |
| EX-01 | M | E2-S1 | Covered |
| EX-02 | M | E2-S2 | Covered |
| EX-03 | M | E2-S5 | Covered |
| EX-04 | M | E2-S6 | Covered |
| EX-05 | M | E2-S3 | Covered |
| EX-06 | S | E2-S4 | Covered |
| EX-07 | S | E2-S7 | Covered |
| EX-08 | S | M2-01 | Covered |
| EX-09 | S | E2-S8 | Covered |
| EX-10 | C | M4-04 | Covered |
| EX-11 | C | M4-05 | Covered |
| IN-01 | M | E3-S1 | Covered |
| IN-02 | M | E3-S2 | Covered |
| IN-03 | M | E3-S4 | Covered |
| IN-04 | M | E3-S4 | Covered |
| IN-05 | M | E3-S3 | Covered |
| IN-06 | S | M2-02 | Covered |
| IN-07 | S | E3-S5 | Covered |
| IN-08 | S | M2-04 | Covered |
| IV-01 | M | E4-S1 | Covered |
| IV-02 | M | E4-S2 | Covered |
| IV-03 | M | E4-S3 | Covered |
| IV-04 | M | E4-S6 | Covered |
| IV-05 | M | E4-S6 | Covered |
| IV-06 | M | E4-S5 | Covered |
| IV-07 | M | E4-S2 | Covered |
| IV-08 | S | E4-S4, E4-S7 | Covered (overlap issue — see Critical) |
| IV-09 | S | M2-03 | Covered |
| IV-10 | S | E4-S5 (PDF preview in InvoiceForm) | Covered |
| IV-11 | S | E4-S8 | Covered |
| IV-12 | C | M4-03 | Covered |
| IV-13 | C | M4-10 | Covered |
| CT-01 | M | E5-S1 | Covered |
| CT-02 | M | E5-S1 | Covered |
| CT-03 | S | E5-S2 | Covered |
| CT-04 | S | E5-S3 | Covered |
| CT-05 | S | E5-S4 | Covered |
| CT-06 | C | M4-01 | Covered |
| DB-01 | M | E6-S1 | Covered |
| DB-02 | M | E6-S6 | Covered |
| DB-03 | S | E6-S2 | Covered |
| DB-04 | S | E6-S3 | Covered |
| DB-05 | S | E6-S4 | Covered |
| DB-06 | S | E6-S5 | Covered |
| DB-07 | C | M4-08 | Covered |
| DB-08 | C | M4-07 | Covered |
| FA-01 | M | E7-S1 | Covered |
| FA-02 | M | E7-S1, E7-S2 | Covered |
| FA-03 | M | E7-S1 | Covered |
| FA-04 | M | E7-S3 | Covered |
| FA-05 | M | E7-S5 | Covered |
| FA-06 | S | E7-S1 (multi-file support in AttachmentList) | Covered |
| FA-07 | S | E7-S4 | Covered |
| FA-08 | S | E4-S5 (PDF inline preview) | Covered |
| FA-09 | C | M4-06 | Covered |
| NF-01 | — | No acceptance criterion | **Gap** |
| NF-04 | — | Indexes in migrations but no test | **Gap** |
| NF-13 | — | M4-09 (deferred) | **Risk** |
| NF-20 | — | No unit test ticket | **Gap** |

All Must-have functional requirements (MoSCoW M) are covered by tickets. The gaps are concentrated in non-functional requirements (performance testing, unit test coverage, backup availability).
