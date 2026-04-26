package handlers

import (
	"net/http"
	"strconv"

	"github.com/beohoang98/moneyapp/internal/models"
	"github.com/beohoang98/moneyapp/internal/services"
)

type InvoiceHandler struct {
	invoiceService *services.InvoiceService
}

func NewInvoiceHandler(is *services.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: is}
}

func (h *InvoiceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/invoices/stats", h.handleStats)
	mux.HandleFunc("POST /api/invoices/check-overdue", h.handleCheckOverdue)
	mux.HandleFunc("POST /api/invoices", h.handleCreate)
	mux.HandleFunc("GET /api/invoices", h.handleList)
	mux.HandleFunc("GET /api/invoices/{id}", h.handleGet)
	mux.HandleFunc("PUT /api/invoices/{id}", h.handleUpdate)
	mux.HandleFunc("DELETE /api/invoices/{id}", h.handleDelete)
	mux.HandleFunc("PATCH /api/invoices/{id}/status", h.handleUpdateStatus)
}

type invoiceRequest struct {
	VendorName  string `json:"vendor_name"`
	Amount      int64  `json:"amount"`
	IssueDate   string `json:"issue_date"`
	DueDate     string `json:"due_date"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

func (h *InvoiceHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req invoiceRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	invoice := &models.Invoice{
		VendorName:  req.VendorName,
		Amount:      req.Amount,
		IssueDate:   req.IssueDate,
		DueDate:     req.DueDate,
		Status:      req.Status,
		Description: req.Description,
	}

	if err := h.invoiceService.Create(r.Context(), invoice); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.invoiceService.GetByID(r.Context(), invoice.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to retrieve created invoice")
		return
	}

	respondSingle(w, http.StatusCreated, created)
}

func (h *InvoiceHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	invoice, err := h.invoiceService.GetByID(r.Context(), id)
	if err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "invoice not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get invoice")
		return
	}

	respondSingle(w, http.StatusOK, invoice)
}

func (h *InvoiceHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	var req invoiceRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	invoice := &models.Invoice{
		VendorName:  req.VendorName,
		Amount:      req.Amount,
		IssueDate:   req.IssueDate,
		DueDate:     req.DueDate,
		Status:      req.Status,
		Description: req.Description,
	}

	if err := h.invoiceService.Update(r.Context(), id, invoice); err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "invoice not found")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.invoiceService.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to retrieve updated invoice")
		return
	}

	respondSingle(w, http.StatusOK, updated)
}

func (h *InvoiceHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	if err := h.invoiceService.Delete(r.Context(), id); err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "invoice not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete invoice")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type statusUpdateRequest struct {
	Status string `json:"status"`
}

func (h *InvoiceHandler) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	var req statusUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Status != "paid" {
		respondError(w, http.StatusBadRequest, "only 'paid' is a valid target status via this endpoint")
		return
	}

	invoice, err := h.invoiceService.MarkAsPaid(r.Context(), id)
	if err != nil {
		if err == services.ErrNotFound {
			respondError(w, http.StatusNotFound, "invoice not found")
			return
		}
		if err.Error() == "invoice is already paid" {
			respondError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update invoice status")
		return
	}

	respondSingle(w, http.StatusOK, invoice)
}

type invoiceListResponse struct {
	Data        []models.Invoice `json:"data"`
	Total       int64            `json:"total"`
	TotalAmount int64            `json:"total_amount"`
	Page        int              `json:"page"`
	PerPage     int              `json:"per_page"`
}

func (h *InvoiceHandler) handleList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	perPage, _ := strconv.Atoi(q.Get("per_page"))

	dateField := q.Get("date_field")
	if dateField != "" && dateField != "issue_date" && dateField != "due_date" {
		respondError(w, http.StatusBadRequest, "date_field must be 'issue_date' or 'due_date'")
		return
	}

	params := services.InvoiceListParams{
		Page:      page,
		PerPage:   perPage,
		Status:    q.Get("status"),
		DateFrom:  q.Get("date_from"),
		DateTo:    q.Get("date_to"),
		DateField: dateField,
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

	result, err := h.invoiceService.List(r.Context(), params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list invoices")
		return
	}

	invoices := result.Invoices
	if invoices == nil {
		invoices = []models.Invoice{}
	}

	respondJSON(w, http.StatusOK, invoiceListResponse{
		Data:        invoices,
		Total:       result.Total,
		TotalAmount: result.TotalAmount,
		Page:        params.Page,
		PerPage:     params.PerPage,
	})
}

func (h *InvoiceHandler) handleCheckOverdue(w http.ResponseWriter, r *http.Request) {
	count, err := h.invoiceService.UpdateOverdueStatuses(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to check overdue invoices")
		return
	}
	respondJSON(w, http.StatusOK, map[string]int64{"updated_count": count})
}

func (h *InvoiceHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.invoiceService.GetStats(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get invoice stats")
		return
	}

	respondSingle(w, http.StatusOK, stats)
}
