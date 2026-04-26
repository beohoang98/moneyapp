---
name: Browser MCP Interaction Patterns for MoneyApp
description: Key findings from Browser MCP live run on MoneyApp React frontend — what works, what doesn't, and 3 new browser-only defects found
type: project
---

## Browser MCP findings — MoneyApp live run 2026-04-26

### React input interaction limitations (critical for future test runs)

**Problem**: All `<input type="number">` and `<textarea>` fields in Add/Edit modals fail with "stale element reference" errors when using `browser_fill`, `browser_type`, or `browser_fill_form`. Root cause: React re-renders between `browser_snapshot` (which assigns refs) and the fill call, invalidating refs.

**Problem 2**: `browser_press_key` sends keyboard events but does NOT trigger React's synthetic `onChange` — the DOM state updates but React state does not, so controlled inputs appear unchanged.

**Workaround used**: Verify form submission via direct API call (`curl`), reload page in browser to confirm list updates.

**What DOES work**:
- `browser_select_option` for native `<select>` (category dropdowns, status filters) — reliable
- `browser_mouse_click_xy` for buttons, tabs, modal confirmations — reliable
- `browser_fill` for login form `<input type="text">` and `<input type="password">` — reliable (non-controlled or simpler React state)
- `browser_click` for buttons that don't require refs from a stale snapshot

### New defects found (browser-only, not in API run)

| ID | Severity | Description |
|----|----------|-------------|
| B-01 | Major | No error message displayed on invalid login credentials — form silently resets |
| B-02 | Major | Date column in all lists shows raw ISO string `2026-04-26T00:00:00Z` not formatted date |
| B-03 | Major | Edit form date field blank — passes ISO datetime to `<input type="date">` which expects `YYYY-MM-DD` |

B-02 and B-03 are the same root cause: backend returns full ISO timestamps, frontend doesn't format them before displaying or binding to date inputs.

### Overall browser run result (2026-04-26)

16 PASS / 2 FAIL / 1 Blocked (date range filter inputs)

**Why:** B-02 and B-03 are the same root cause.
**How to apply:** When investigating date display issues, look at the frontend date formatting logic in expense/income/invoice list and form components.
