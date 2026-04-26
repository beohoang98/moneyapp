package handlers

import (
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
}

func (h *DashboardHandler) handleSummary(w http.ResponseWriter, r *http.Request) {
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	summary, err := h.dashboardService.GetSummary(r.Context(), dateFrom, dateTo)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get dashboard summary")
		return
	}

	respondSingle(w, http.StatusOK, summary)
}
