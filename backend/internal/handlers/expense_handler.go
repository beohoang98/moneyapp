package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/beohoang98/moneyapp/internal/models"
	"github.com/beohoang98/moneyapp/internal/services"
)

type ExpenseHandler struct {
	expenseService *services.ExpenseService
}

func NewExpenseHandler(es *services.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: es}
}

func (h *ExpenseHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/expenses", h.handleCreate)
	mux.HandleFunc("GET /api/expenses", h.handleList)
	mux.HandleFunc("GET /api/expenses/{id}", h.handleGet)
	mux.HandleFunc("PUT /api/expenses/{id}", h.handleUpdate)
	mux.HandleFunc("DELETE /api/expenses/{id}", h.handleDelete)
}

type expenseRequest struct {
	Amount      int64  `json:"amount"`
	Date        string `json:"date"`
	CategoryID  int64  `json:"category_id"`
	Description string `json:"description"`
}

func (h *ExpenseHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req expenseRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	expense := &models.Expense{
		Amount:      req.Amount,
		Date:        req.Date,
		CategoryID:  req.CategoryID,
		Description: req.Description,
	}

	if err := h.expenseService.Create(r.Context(), expense); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.expenseService.GetByID(r.Context(), expense.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to retrieve created expense")
		return
	}

	respondSingle(w, http.StatusCreated, created)
}

func (h *ExpenseHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	expense, err := h.expenseService.GetByID(r.Context(), id)
	if err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "expense not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get expense")
		return
	}

	respondSingle(w, http.StatusOK, expense)
}

func (h *ExpenseHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	var req expenseRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	expense := &models.Expense{
		Amount:      req.Amount,
		Date:        req.Date,
		CategoryID:  req.CategoryID,
		Description: req.Description,
	}

	if err := h.expenseService.Update(r.Context(), id, expense); err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "expense not found")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.expenseService.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to retrieve updated expense")
		return
	}

	respondSingle(w, http.StatusOK, updated)
}

func (h *ExpenseHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	if err := h.expenseService.Delete(r.Context(), id); err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "expense not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete expense")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type expenseListResponse struct {
	Data        []models.Expense `json:"data"`
	Total       int64            `json:"total"`
	TotalAmount int64            `json:"total_amount"`
	Page        int              `json:"page"`
	PerPage     int              `json:"per_page"`
}

func (h *ExpenseHandler) handleList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	perPage, _ := strconv.Atoi(q.Get("per_page"))

	sortBy := q.Get("sort_by")
	if sortBy != "" && sortBy != "date" && sortBy != "amount" {
		respondError(w, http.StatusBadRequest, "sort_by must be 'date' or 'amount'")
		return
	}
	sortOrder := q.Get("sort_order")
	if sortOrder != "" && sortOrder != "asc" && sortOrder != "desc" {
		respondError(w, http.StatusBadRequest, "sort_order must be 'asc' or 'desc'")
		return
	}

	params := services.ExpenseListParams{
		Page:      page,
		PerPage:   perPage,
		DateFrom:  q.Get("date_from"),
		DateTo:    q.Get("date_to"),
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	if e := validateOptionalISODate("date_from", params.DateFrom); e != nil {
		respondError(w, http.StatusBadRequest, e.Error())
		return
	}
	if e := validateOptionalISODate("date_to", params.DateTo); e != nil {
		respondError(w, http.StatusBadRequest, e.Error())
		return
	}

	if df := params.DateFrom; df != "" {
		if dt := params.DateTo; dt != "" && df > dt {
			respondError(w, http.StatusBadRequest, "date_from must be before date_to")
			return
		}
	}

	params.Page, params.PerPage = clampPagePerPage(params.Page, params.PerPage)

	if catIDStr := q.Get("category_id"); catIDStr != "" {
		id, err := strconv.ParseInt(catIDStr, 10, 64)
		if err == nil {
			params.CategoryIDs = []int64{id}
		}
	}
	if catIDsStr := q.Get("category_ids"); catIDsStr != "" {
		for _, s := range strings.Split(catIDsStr, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
			if err == nil {
				params.CategoryIDs = append(params.CategoryIDs, id)
			}
		}
	}

	result, err := h.expenseService.List(r.Context(), params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list expenses")
		return
	}

	expenses := result.Expenses
	if expenses == nil {
		expenses = []models.Expense{}
	}

	respondJSON(w, http.StatusOK, expenseListResponse{
		Data:        expenses,
		Total:       result.Total,
		TotalAmount: result.TotalAmount,
		Page:        params.Page,
		PerPage:     params.PerPage,
	})
}
