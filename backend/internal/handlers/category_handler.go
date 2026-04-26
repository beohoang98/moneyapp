package handlers

import (
	"net/http"
	"strconv"

	"github.com/beohoang98/moneyapp/internal/models"
	"github.com/beohoang98/moneyapp/internal/services"
)

type CategoryHandler struct {
	categoryService *services.CategoryService
}

func NewCategoryHandler(cs *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: cs}
}

func (h *CategoryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/categories", h.handleList)
	mux.HandleFunc("POST /api/categories", h.handleCreate)
	mux.HandleFunc("PUT /api/categories/{id}", h.handleUpdate)
	mux.HandleFunc("DELETE /api/categories/{id}", h.handleDelete)
}

func (h *CategoryHandler) handleList(w http.ResponseWriter, r *http.Request) {
	categoryType := r.URL.Query().Get("type")
	if categoryType != "" && categoryType != "expense" && categoryType != "income" {
		respondError(w, http.StatusBadRequest, "type must be 'expense' or 'income'")
		return
	}

	categories, err := h.categoryService.List(r.Context(), categoryType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}

	if categories == nil {
		categories = []models.Category{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": categories})
}

type categoryCreateRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (h *CategoryHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req categoryCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.categoryService.Create(r.Context(), req.Name, req.Type)
	if err != nil {
		if err == services.ErrDuplicateCategory {
			respondError(w, http.StatusConflict, "Category already exists")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondSingle(w, http.StatusCreated, cat)
}

type categoryUpdateRequest struct {
	Name string `json:"name"`
}

func (h *CategoryHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	var req categoryUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.categoryService.Update(r.Context(), id, req.Name)
	if err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "category not found")
			return
		}
		if err == services.ErrDefaultCategory {
			respondError(w, http.StatusForbidden, "Default categories cannot be modified")
			return
		}
		if err == services.ErrDuplicateCategory {
			respondError(w, http.StatusConflict, "Category already exists")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondSingle(w, http.StatusOK, cat)
}

func (h *CategoryHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	if err := h.categoryService.Delete(r.Context(), id); err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "category not found")
			return
		}
		if err == services.ErrDefaultCategory {
			respondError(w, http.StatusForbidden, "Default categories cannot be deleted")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete category")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
