.PHONY: dev-backend dev-frontend dev minio

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
