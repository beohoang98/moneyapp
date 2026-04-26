package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/beohoang98/moneyapp/internal/models"
	"gorm.io/gorm"
)

type InvoiceListParams struct {
	Page      int
	PerPage   int
	Status    string
	DateFrom  string
	DateTo    string
	DateField string
}

type InvoiceListResult struct {
	Invoices    []models.Invoice
	Total       int64
	TotalAmount int64
}

type InvoiceService struct {
	db                *gorm.DB
	attachmentService *AttachmentService
}

func NewInvoiceService(db *gorm.DB) *InvoiceService {
	return &InvoiceService{db: db}
}

func (s *InvoiceService) SetAttachmentService(as *AttachmentService) {
	s.attachmentService = as
}

func (s *InvoiceService) Create(ctx context.Context, inv *models.Invoice) error {
	if inv.VendorName == "" {
		return fmt.Errorf("vendor name is required")
	}
	if inv.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if inv.IssueDate == "" {
		return fmt.Errorf("issue date is required")
	}
	if inv.DueDate == "" {
		return fmt.Errorf("due date is required")
	}
	if inv.DueDate < inv.IssueDate {
		return fmt.Errorf("due date must be on or after issue date")
	}
	if inv.Status == "" {
		inv.Status = "unpaid"
	}
	if inv.Status != "unpaid" && inv.Status != "paid" && inv.Status != "overdue" {
		return fmt.Errorf("status must be unpaid, paid, or overdue")
	}

	if err := s.db.WithContext(ctx).Create(inv).Error; err != nil {
		return fmt.Errorf("insert invoice: %w", err)
	}
	return nil
}

func (s *InvoiceService) GetByID(ctx context.Context, id int64) (*models.Invoice, error) {
	var inv models.Invoice
	if err := s.db.WithContext(ctx).First(&inv, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query invoice: %w", err)
	}
	return &inv, nil
}

func (s *InvoiceService) Update(ctx context.Context, id int64, inv *models.Invoice) error {
	if inv.Status == "" {
		existing, err := s.GetByID(ctx, id)
		if err != nil {
			return err
		}
		inv.Status = existing.Status
	}
	if inv.VendorName == "" {
		return fmt.Errorf("vendor name is required")
	}
	if inv.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if inv.IssueDate == "" {
		return fmt.Errorf("issue date is required")
	}
	if inv.DueDate == "" {
		return fmt.Errorf("due date is required")
	}
	if inv.DueDate < inv.IssueDate {
		return fmt.Errorf("due date must be on or after issue date")
	}
	if inv.Status != "unpaid" && inv.Status != "paid" && inv.Status != "overdue" {
		return fmt.Errorf("status must be unpaid, paid, or overdue")
	}

	result := s.db.WithContext(ctx).Model(&models.Invoice{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"vendor_name": inv.VendorName,
			"amount":      inv.Amount,
			"issue_date":  inv.IssueDate,
			"due_date":    inv.DueDate,
			"status":      inv.Status,
			"description": inv.Description,
			"updated_at":  gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		return fmt.Errorf("update invoice: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *InvoiceService) Delete(ctx context.Context, id int64) error {
	if s.attachmentService != nil {
		if err := s.attachmentService.DeleteByEntity(ctx, "invoice", id); err != nil {
			return fmt.Errorf("delete invoice attachments: %w", err)
		}
	}

	result := s.db.WithContext(ctx).Delete(&models.Invoice{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete invoice: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *InvoiceService) UpdateOverdueStatuses(ctx context.Context) (int64, error) {
	result := s.db.WithContext(ctx).Model(&models.Invoice{}).
		Where("status = 'unpaid' AND due_date < date('now')").
		Updates(map[string]interface{}{
			"status":     "overdue",
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		return 0, fmt.Errorf("update overdue statuses: %w", result.Error)
	}
	if result.RowsAffected > 0 {
		log.Printf("Updated %d invoices to overdue status", result.RowsAffected)
	}
	return result.RowsAffected, nil
}

func (s *InvoiceService) MarkAsPaid(ctx context.Context, id int64) (*models.Invoice, error) {
	var inv models.Invoice
	if err := s.db.WithContext(ctx).Select("status").First(&inv, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query invoice status: %w", err)
	}
	if inv.Status == "paid" {
		return nil, fmt.Errorf("invoice is already paid")
	}

	s.db.WithContext(ctx).Model(&models.Invoice{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "paid",
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})

	return s.GetByID(ctx, id)
}

func (s *InvoiceService) List(ctx context.Context, params InvoiceListParams) (*InvoiceListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}

	if _, err := s.UpdateOverdueStatuses(ctx); err != nil {
		log.Printf("Warning: failed to update overdue statuses: %v", err)
	}

	base := s.db.WithContext(ctx).Model(&models.Invoice{})
	base = applyInvoiceFilters(base, params)

	var total int64
	var totalAmount *int64
	row := base.Select("COUNT(*), COALESCE(SUM(amount), 0)").Row()
	if err := row.Scan(&total, &totalAmount); err != nil {
		return nil, fmt.Errorf("count invoices: %w", err)
	}

	offset := (params.Page - 1) * params.PerPage

	var invoices []models.Invoice
	q := s.db.WithContext(ctx).Model(&models.Invoice{})
	q = applyInvoiceFilters(q, params)
	if err := q.Order("due_date ASC, id DESC").
		Limit(params.PerPage).Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("query invoices: %w", err)
	}

	amt := int64(0)
	if totalAmount != nil {
		amt = *totalAmount
	}
	return &InvoiceListResult{
		Invoices:    invoices,
		Total:       total,
		TotalAmount: amt,
	}, nil
}

func (s *InvoiceService) GetStats(ctx context.Context) (*models.InvoiceStats, error) {
	if _, err := s.UpdateOverdueStatuses(ctx); err != nil {
		log.Printf("Warning: failed to update overdue statuses: %v", err)
	}

	type statusRow struct {
		Status string
		Cnt    int
		Amt    int64
	}
	var rows []statusRow
	err := s.db.WithContext(ctx).Model(&models.Invoice{}).
		Select("status, COUNT(*) AS cnt, COALESCE(SUM(amount), 0) AS amt").
		Where("status IN ('unpaid', 'overdue')").
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("query invoice stats: %w", err)
	}

	stats := &models.InvoiceStats{}
	for _, r := range rows {
		switch r.Status {
		case "unpaid":
			stats.UnpaidCount = r.Cnt
			stats.UnpaidAmount = r.Amt
			stats.TotalOutstanding += r.Amt
		case "overdue":
			stats.OverdueCount = r.Cnt
			stats.OverdueAmount = r.Amt
			stats.TotalOutstanding += r.Amt
		}
	}
	return stats, nil
}

func applyInvoiceFilters(q *gorm.DB, params InvoiceListParams) *gorm.DB {
	dateCol := "due_date"
	if params.DateField == "issue_date" {
		dateCol = "issue_date"
	}

	if params.Status != "" {
		q = q.Where("status = ?", params.Status)
	}
	if params.DateFrom != "" {
		q = q.Where(dateCol+" >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		q = q.Where(dateCol+" <= ?", params.DateTo)
	}
	return q
}
