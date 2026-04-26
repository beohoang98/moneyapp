---
name: M0 Visual Verification — live run findings
description: Results from 2026-04-26 live browser verification of M0 routing and UI shell; includes one FAIL on responsive nav
type: project
---

Live run on 2026-04-26: backend `CGO_ENABLED=1 STORAGE_TYPE=local go run ./cmd/server` (port 8080) + frontend `npm run dev` (port 5173). Backend requires `CGO_ENABLED=1` — without it, go-sqlite3 fails with stub error.

**Why:** CGO_ENABLED must be set explicitly because the default macOS Go toolchain may strip CGO. This is a prerequisite for all future local runs.
**How to apply:** Always set `CGO_ENABLED=1` when running the backend locally.

**Findings:**
- All M0-04 routing scenarios pass visually: `/` → dashboard, `/expenses`, `/login` (no sidebar), `/foo/bar` → 404
- `GET /api/health` returns `{"status":"ok","database":"ok","storage":"ok","storage_type":"local"}` on local-storage mode
- **TS-04-06 FAIL**: Sidebar does NOT collapse at 375px mobile viewport. No hamburger/mobile nav. Full sidebar is visible and content area is squeezed. This is the only substantive gap found in M0 visual pass.

Screenshots live in `docs/milestone-00/` (m0-01-dashboard.png, m0-04-expenses.png, m0-04-login.png, m0-04-not-found.png, m0-04-mobile-nav.png, m0-06-health-check.png).
