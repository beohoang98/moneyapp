package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/beohoang98/moneyapp/internal/storage"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db          *gorm.DB
	storage     storage.ObjectStore
	storageType string // "local" | "s3" — for health JSON only
}

func NewHealthHandler(db *gorm.DB, s storage.ObjectStore, storageType string) *HealthHandler {
	return &HealthHandler{db: db, storage: s, storageType: storageType}
}

func (h *HealthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", h.handleHealth)
}

func (h *HealthHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	dbStatus := "ok"
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.PingContext(ctx) != nil {
		dbStatus = "error"
	}

	storageStatus := "ok"
	if h.storage != nil {
		if err := h.storage.HealthCheck(ctx); err != nil {
			storageStatus = "error"
		}
	} else {
		storageStatus = "error"
	}

	overallStatus := "ok"
	statusCode := http.StatusOK

	if dbStatus == "error" {
		overallStatus = "error"
		statusCode = http.StatusServiceUnavailable
	} else if storageStatus == "error" {
		overallStatus = "degraded"
	}

	respondJSON(w, statusCode, map[string]string{
		"status":       overallStatus,
		"database":     dbStatus,
		"storage":      storageStatus,
		"storage_type": h.storageType,
	})
}
