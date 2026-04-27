package handlers

import (
	"net/http"
	"strconv"

	"github.com/beohoang98/moneyapp/internal/models"
	"github.com/beohoang98/moneyapp/internal/services"
)

const maxUploadSize = 10*1024*1024 + 1024 // 10 MB + overhead for multipart headers

type AttachmentHandler struct {
	attachmentService *services.AttachmentService
}

func NewAttachmentHandler(as *services.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{attachmentService: as}
}

func (h *AttachmentHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/attachments", h.handleUpload)
	mux.HandleFunc("GET /api/attachments", h.handleList)
	mux.HandleFunc("DELETE /api/attachments/{id}", h.handleDelete)
	mux.HandleFunc("GET /api/attachments/{id}/download", h.handleDownload)
	mux.HandleFunc("GET /api/attachments/{id}/preview", h.handlePreview)
}

func (h *AttachmentHandler) handleUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusRequestEntityTooLarge, "file exceeds the 10 MB limit")
		return
	}

	entityType := r.FormValue("entity_type")
	entityIDStr := r.FormValue("entity_id")
	entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
	if err != nil || entityID <= 0 {
		respondError(w, http.StatusBadRequest, "invalid entity_id")
		return
	}

	sourceStorageKey := r.FormValue("source_storage_key")

	_, _, fileErr := r.FormFile("file")
	hasFile := fileErr == nil

	if sourceStorageKey != "" && hasFile {
		respondError(w, http.StatusBadRequest, "cannot provide both file and source_storage_key")
		return
	}

	if sourceStorageKey != "" {
		att, err := h.attachmentService.PromoteFromTemp(r.Context(), entityType, entityID, sourceStorageKey)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondSingle(w, http.StatusCreated, att)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	att, err := h.attachmentService.Upload(r.Context(), entityType, entityID, file, header)
	if err != nil {
		if err.Error() == "file exceeds the 10 MB limit" {
			respondError(w, http.StatusRequestEntityTooLarge, err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondSingle(w, http.StatusCreated, att)
}

func (h *AttachmentHandler) handleList(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityIDStr := r.URL.Query().Get("entity_id")
	entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
	if err != nil || entityID <= 0 {
		respondError(w, http.StatusBadRequest, "entity_type and entity_id are required")
		return
	}

	attachments, err := h.attachmentService.ListByEntity(r.Context(), entityType, entityID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list attachments")
		return
	}

	if attachments == nil {
		attachments = []models.Attachment{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": attachments})
}

func (h *AttachmentHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid attachment id")
		return
	}

	if err := h.attachmentService.Delete(r.Context(), id); err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "attachment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete attachment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AttachmentHandler) handleDownload(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid attachment id")
		return
	}

	att, reader, err := h.attachmentService.Download(r.Context(), id)
	if err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "attachment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to download attachment")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", att.MimeType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+att.Filename+`"`)
	w.WriteHeader(http.StatusOK)

	if _, err := copyBuffer(w, reader); err != nil {
		return
	}
}

func (h *AttachmentHandler) handlePreview(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid attachment id")
		return
	}

	att, reader, err := h.attachmentService.Download(r.Context(), id)
	if err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "attachment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to preview attachment")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", att.MimeType)
	w.Header().Set("Content-Disposition", `inline; filename="`+att.Filename+`"`)
	w.WriteHeader(http.StatusOK)

	if _, err := copyBuffer(w, reader); err != nil {
		return
	}
}
