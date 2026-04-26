---
name: Configurable Storage Backend Decision
description: Storage backend changed from MinIO-only to STORAGE_TYPE=local|s3 — impacts requirements, tickets, health endpoint, and CLAUDE.md
type: project
---

Storage is now configurable via `STORAGE_TYPE` env var (`local` | `s3`).

- `STORAGE_TYPE=local` (default): writes files to `LOCAL_STORAGE_PATH` (default `./uploads`); no Docker/MinIO needed.
- `STORAGE_TYPE=s3`: uses MinIO or any S3-compatible endpoint via existing `MINIO_*` env vars.

**Why:** Stakeholder requirement so developers can run `go run ./cmd/server` without Docker during basic development.

**How to apply:**
- Any reference to "MinIO" as the storage layer in requirements or tickets should be phrased as "storage backend" unless it is explicitly in an s3-mode context.
- Health endpoint behavior differs by mode: local = check path writable; s3 = BucketExists ping.
- Docker Compose stack always uses `STORAGE_TYPE=s3` (MinIO).
- The `Storage` interface in `internal/storage/` (with `local.go` and `s3.go` implementations) is the architectural pattern — `AttachmentService` takes `Storage` interface, not a concrete MinIO client.
- All doc edits captured in requirements.md v1.2 changelog and tickets.md (ref: `docs/requirements.md`, `docs/tickets.md`, `CLAUDE.md`, `docs/test-plans/milestone-0.md`).
