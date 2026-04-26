package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/beohoang98/moneyapp/internal/models"
	"github.com/beohoang98/moneyapp/internal/services"
)

type IncomeHandler struct {
	incomeService *services.IncomeService
}

func NewIncomeHandler(is *services.IncomeService) *IncomeHandler {
	return &IncomeHandler{incomeService: is}
}

func (h *IncomeHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/incomes", h.handleCreate)
	mux.HandleFunc("GET /api/incomes", h.handleList)
	mux.HandleFunc("GET /api/incomes/{id}", h.handleGet)
	mux.HandleFunc("PUT /api/incomes/{id}", h.handleUpdate)
	mux.HandleFunc("DELETE /api/incomes/{id}", h.handleDelete)
}

type incomeRequest struct {
	Amount      int64  `json:"amount"`
	Date        string `json:"date"`
	CategoryID  int64  `json:"category_id"`
	Description string `json:"description"`
}

func (h *IncomeHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req incomeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	income := &models.Income{
		Amount:      req.Amount,
		Date:        req.Date,
		CategoryID:  req.CategoryID,
		Description: req.Description,
	}

	if err := h.incomeService.Create(r.Context(), income); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.incomeService.GetByID(r.Context(), income.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to retrieve created income")
		return
	}

	respondSingle(w, http.StatusCreated, created)
}

func (h *IncomeHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid income id")
		return
	}

	income, err := h.incomeService.GetByID(r.Context(), id)
	if err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "income not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get income")
		return
	}

	respondSingle(w, http.StatusOK, income)
}

func (h *IncomeHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid income id")
		return
	}

	var req incomeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	income := &models.Income{
		Amount:      req.Amount,
		Date:        req.Date,
		CategoryID:  req.CategoryID,
		Description: req.Description,
	}

	if err := h.incomeService.Update(r.Context(), id, income); err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "income not found")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.incomeService.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to retrieve updated income")
		return
	}

	respondSingle(w, http.StatusOK, updated)
}

func (h *IncomeHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid income id")
		return
	}

	if err := h.incomeService.Delete(r.Context(), id); err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "income not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete income")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type incomeListResponse struct {
	Data        []models.Income `json:"data"`
	Total       int64           `json:"total"`
	TotalAmount int64           `json:"total_amount"`
	Page        int             `json:"page"`
	PerPage     int             `json:"per_page"`
}

func (h *IncomeHandler) handleList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	perPage, _ := strconv.Atoi(q.Get("per_page"))

	params := services.IncomeListParams{
		Page:     page,
		PerPage:  perPage,
		DateFrom: q.Get("date_from"),
		DateTo:   q.Get("date_to"),
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

	result, err := h.incomeService.List(r.Context(), params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list incomes")
		return
	}

	incomes := result.Incomes
	if incomes == nil {
		incomes = []models.Income{}
	}

	respondJSON(w, http.StatusOK, incomeListResponse{
		Data:        incomes,
		Total:       result.Total,
		TotalAmount: result.TotalAmount,
		Page:        params.Page,
		PerPage:     params.PerPage,
	})
}
