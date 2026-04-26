package handlers

import (
	"net/http"

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
