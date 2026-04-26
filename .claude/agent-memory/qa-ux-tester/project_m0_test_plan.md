---
name: M0 Test Plan — Coverage & Gaps
description: Test plan written for Milestone 0 (Project Foundation), covering TS-01 through TS-08 across all 8 M0 tickets; notes known coverage gaps and ambiguities to resolve before sign-off.
type: project
---

M0 test plan created at `docs/test-plans/milestone-0.md` (2026-04-26).

**Why:** M0 is pure infrastructure — no user-facing features. Test focus is startup correctness, migration runner, env config, middleware, health endpoint, frontend routing, API client, Docker Compose, and CI.

**Coverage**: 48 test cases across 8 suites (TS-01–TS-08), plus cross-cutting smoke tests.

**How to apply:** When reviewing M0 PRs or running M0 smoke tests, reference this file. The "Coverage Gaps" table at the bottom of the plan captures open questions that require ticket clarification before sign-off.

## Open ambiguities from ticket spec (block sign-off if unresolved):
- `JWT_SECRET` policy (warn vs. hard-fail) — M0-02 AC2
- No test endpoint for panic recovery smoke — M0-03 AC2
- `Toast` component has no dev-trigger; TS-04-08 is conditional
- `useAuth` hook re-hydration on page refresh not specified — M0-05

## Automation backlog recorded in plan:
- Go integration tests: migration runner (TS-01), health handler (TS-06), config loader (TS-02)
- Playwright: route navigation (TS-04), 401 redirect with route interception (TS-05-03)
