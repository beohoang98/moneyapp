.PHONY: dev-backend dev-frontend dev minio test-e2e

# Run backend
dev-backend:
	cd backend && go run ./cmd/server

# Run frontend
dev-frontend:
	cd frontend && npm run dev

# Run MinIO
minio:
	docker compose up minio -d

# Run all services (requires multiple terminals or use tmux)
dev: minio
	@echo "Run 'make dev-backend' and 'make dev-frontend' in separate terminals"

# Playwright e2e (Chromium); installs deps on first run
test-e2e:
	cd e2e && npm install && npx playwright install chromium && npm test
