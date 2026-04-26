---
name: Backend local startup requires CGO_ENABLED=1
description: go-sqlite3 fails without CGO; must set CGO_ENABLED=1 for local backend runs
type: feedback
---

Always run backend as `CGO_ENABLED=1 STORAGE_TYPE=local go run ./cmd/server` (or with `LOCAL_STORAGE_PATH` set).

**Why:** go-sqlite3 is a CGO library. Without `CGO_ENABLED=1`, the binary is compiled with a stub and panics on first DB open with: "Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub."
**How to apply:** Every time you start or test the backend locally, prefix with `CGO_ENABLED=1`. This applies to `go run`, `go test`, and `go build`.
