package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/beohoang98/moneyapp/internal/models"
	"github.com/beohoang98/moneyapp/internal/services"
)

type ScanningHandler struct {
	scanningService *services.ScanningService
}

func NewScanningHandler(ss *services.ScanningService) *ScanningHandler {
	return &ScanningHandler{scanningService: ss}
}

func (h *ScanningHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/settings/scanning", h.handleGetSettings)
	mux.HandleFunc("PUT /api/settings/scanning", h.handleUpdateSettings)
	mux.HandleFunc("POST /api/settings/scanning/test", h.handleTestConnection)
	mux.HandleFunc("GET /api/scanning/health", h.handleHealth)
	mux.HandleFunc("POST /api/scanning/invoice", h.handleScanInvoice)
	mux.HandleFunc("DELETE /api/scanning/temp", h.handleDeleteTemp)
}

func (h *ScanningHandler) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.scanningService.GetSettings(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get scanning settings")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":     settings.Enabled,
		"base_url":    settings.BaseURL,
		"model":       settings.Model,
		"api_key_set": settings.APIKey != "",
		"updated_at":  settings.UpdatedAt,
	})
}

func (h *ScanningHandler) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled bool   `json:"enabled"`
		BaseURL string `json:"base_url"`
		Model   string `json:"model"`
		APIKey  string `json:"api_key"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	update := &models.ScanningSettings{
		Enabled: body.Enabled,
		BaseURL: body.BaseURL,
		Model:   body.Model,
		APIKey:  body.APIKey,
	}

	if err := h.scanningService.UpdateSettings(r.Context(), update); err != nil {
		if strings.Contains(err.Error(), "base_url must use") {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update scanning settings")
		return
	}

	h.handleGetSettings(w, r)
}

func (h *ScanningHandler) handleTestConnection(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BaseURL string `json:"base_url"`
		Model   string `json:"model"`
		APIKey  string `json:"api_key"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ok, message := h.scanningService.TestConnection(r.Context(), body.BaseURL, body.Model, body.APIKey)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      ok,
		"message": message,
	})
}

func (h *ScanningHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	ok, message := h.scanningService.CheckHealth(r.Context())
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      ok,
		"message": message,
	})
}

func (h *ScanningHandler) handleScanInvoice(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusRequestEntityTooLarge, "file exceeds the 10 MB limit")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "image file is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType != "" {
		contentType = strings.Split(contentType, ";")[0]
		contentType = strings.TrimSpace(contentType)
	}

	result, storageKey, err := h.scanningService.ScanImage(r.Context(), file, header.Filename, contentType)
	if err != nil {
		h.respondScanError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"scan_result":      result,
		"temp_storage_key": storageKey,
	})
}

func (h *ScanningHandler) handleDeleteTemp(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StorageKey string `json:"storage_key"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.scanningService.DeleteTempScan(r.Context(), body.StorageKey); err != nil {
		if errors.Is(err, services.ErrInvalidStorageKey) {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":  "invalid_storage_key",
				"detail": "storage_key must start with scan-tmp/",
			})
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete temp scan")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ScanningHandler) respondScanError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrScanningDisabled):
		respondJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"error": "scanning_disabled",
		})
	case errors.Is(err, services.ErrInvalidFileType):
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "invalid_file_type",
		})
	case errors.Is(err, services.ErrExtractionFailed):
		respondJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
			"error":  "extraction_failed",
			"detail": "Could not extract data from this image",
		})
	case errors.Is(err, services.ErrScanTimeout):
		respondJSON(w, http.StatusGatewayTimeout, map[string]interface{}{
			"error": "scan_timeout",
		})
	case errors.Is(err, services.ErrTooManyScans):
		respondJSON(w, http.StatusTooManyRequests, map[string]interface{}{
			"error": "too_many_scans",
		})
	default:
		respondError(w, http.StatusInternalServerError, "scan failed")
	}
}
