package handlers

import (
	"errors"
	"net/http"

	"github.com/beohoang98/moneyapp/internal/services"
)

type DashboardHandler struct {
	dashboardService *services.DashboardService
}

func NewDashboardHandler(ds *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: ds}
}

func (h *DashboardHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/dashboard/summary", h.handleSummary)
	mux.HandleFunc("GET /api/dashboard/monthly-trend", h.handleMonthlyTrend)
	mux.HandleFunc("GET /api/dashboard/expense-by-category", h.handleExpenseByCategory)
}

func (h *DashboardHandler) parseDateRange(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")
	if err := validateOptionalISODate("date_from", dateFrom); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return "", "", false
	}
	if err := validateOptionalISODate("date_to", dateTo); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return "", "", false
	}
	if err := h.dashboardService.ValidateDateRange(dateFrom, dateTo); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return "", "", false
	}
	return dateFrom, dateTo, true
}

func (h *DashboardHandler) handleSummary(w http.ResponseWriter, r *http.Request) {
	dateFrom, dateTo, ok := h.parseDateRange(w, r)
	if !ok {
		return
	}

	summary, err := h.dashboardService.GetSummary(r.Context(), dateFrom, dateTo)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get dashboard summary")
		return
	}

	respondSingle(w, http.StatusOK, summary)
}

func (h *DashboardHandler) handleMonthlyTrend(w http.ResponseWriter, r *http.Request) {
	dateFrom, dateTo, ok := h.parseDateRange(w, r)
	if !ok {
		return
	}

	items, err := h.dashboardService.GetMonthlyTrend(r.Context(), dateFrom, dateTo)
	if err != nil {
		if errors.Is(err, services.ErrDateRangeInvalid) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get monthly trend")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (h *DashboardHandler) handleExpenseByCategory(w http.ResponseWriter, r *http.Request) {
	dateFrom, dateTo, ok := h.parseDateRange(w, r)
	if !ok {
		return
	}

	items, err := h.dashboardService.GetExpenseByCategory(r.Context(), dateFrom, dateTo)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get expense by category")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}
