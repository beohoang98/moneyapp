package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/beohoang98/moneyapp/internal/services"
)

type ExportHandler struct {
	expenseService  *services.ExpenseService
	incomeService   *services.IncomeService
	categoryService *services.CategoryService
}

func NewExportHandler(es *services.ExpenseService, is *services.IncomeService, cs *services.CategoryService) *ExportHandler {
	return &ExportHandler{expenseService: es, incomeService: is, categoryService: cs}
}

func (h *ExportHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/export/transactions", h.handleExport)
}

func (h *ExportHandler) handleExport(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	txnType := q.Get("type")
	if txnType != "expense" && txnType != "income" {
		respondError(w, http.StatusBadRequest, "type must be 'expense' or 'income'")
		return
	}

	dateFrom := q.Get("date_from")
	dateTo := q.Get("date_to")
	if err := validateOptionalISODate("date_from", dateFrom); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateOptionalISODate("date_to", dateTo); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	categoryIDStr := q.Get("category_id")
	var categoryIDs []int64
	if categoryIDStr != "" {
		cid, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid category_id")
			return
		}
		categoryIDs = []int64{cid}
	}

	today := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%ss_%s.csv", txnType, today)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	// UTF-8 BOM for Excel compatibility
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"date", "type", "category", "description", "amount"})

	if txnType == "expense" {
		h.writeExpenses(w, r, writer, dateFrom, dateTo, categoryIDs)
	} else {
		h.writeIncomes(w, r, writer, dateFrom, dateTo, categoryIDs)
	}
}

func (h *ExportHandler) writeExpenses(w http.ResponseWriter, r *http.Request, writer *csv.Writer, dateFrom, dateTo string, categoryIDs []int64) {
	page := 1
	for {
		result, err := h.expenseService.List(r.Context(), services.ExpenseListParams{
			Page:        page,
			PerPage:     100,
			DateFrom:    dateFrom,
			DateTo:      dateTo,
			CategoryIDs: categoryIDs,
			SortBy:      "date",
			SortOrder:   "asc",
		})
		if err != nil {
			return
		}

		for _, e := range result.Expenses {
			writer.Write([]string{
				e.Date,
				"expense",
				e.CategoryName,
				csvSafe(e.Description),
				strconv.FormatInt(e.Amount, 10),
			})
		}
		writer.Flush()

		if int64(page*100) >= result.Total {
			break
		}
		page++
	}
}

func (h *ExportHandler) writeIncomes(w http.ResponseWriter, r *http.Request, writer *csv.Writer, dateFrom, dateTo string, categoryIDs []int64) {
	page := 1
	for {
		result, err := h.incomeService.List(r.Context(), services.IncomeListParams{
			Page:        page,
			PerPage:     100,
			DateFrom:    dateFrom,
			DateTo:      dateTo,
			CategoryIDs: categoryIDs,
		})
		if err != nil {
			return
		}

		for _, inc := range result.Incomes {
			writer.Write([]string{
				inc.Date,
				"income",
				inc.CategoryName,
				csvSafe(inc.Description),
				strconv.FormatInt(inc.Amount, 10),
			})
		}
		writer.Flush()

		if int64(page*100) >= result.Total {
			break
		}
		page++
	}
}

func csvSafe(s string) string {
	if len(s) == 0 {
		return s
	}
	first := s[0]
	if first == '=' || first == '+' || first == '-' || first == '@' {
		return "'" + strings.TrimSpace(s)
	}
	return s
}
