package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/beohoang98/moneyapp/internal/models"
)

type InvoiceListParams struct {
	Page     int
	PerPage  int
	Status   string
	DateFrom string
	DateTo   string
}

type InvoiceListResult struct {
	Invoices    []models.Invoice
	Total       int64
	TotalAmount int64
}

type InvoiceService struct {
	db *sql.DB
}

func NewInvoiceService(db *sql.DB) *InvoiceService {
	return &InvoiceService{db: db}
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

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO invoices (vendor_name, amount, issue_date, due_date, status, description) VALUES (?, ?, ?, ?, ?, ?)`,
		inv.VendorName, inv.Amount, inv.IssueDate, inv.DueDate, inv.Status, inv.Description,
	)
	if err != nil {
		return fmt.Errorf("insert invoice: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	inv.ID = id
	return nil
}

func (s *InvoiceService) GetByID(ctx context.Context, id int64) (*models.Invoice, error) {
	var inv models.Invoice
	err := s.db.QueryRowContext(ctx,
		`SELECT id, vendor_name, amount, issue_date, due_date, status, description, created_at, updated_at
		 FROM invoices WHERE id = ?`, id,
	).Scan(&inv.ID, &inv.VendorName, &inv.Amount, &inv.IssueDate, &inv.DueDate, &inv.Status, &inv.Description, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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

	result, err := s.db.ExecContext(ctx,
		`UPDATE invoices SET vendor_name = ?, amount = ?, issue_date = ?, due_date = ?, status = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		inv.VendorName, inv.Amount, inv.IssueDate, inv.DueDate, inv.Status, inv.Description, id,
	)
	if err != nil {
		return fmt.Errorf("update invoice: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *InvoiceService) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM invoices WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete invoice: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *InvoiceService) UpdateOverdueStatuses(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE invoices SET status = 'overdue', updated_at = CURRENT_TIMESTAMP WHERE status = 'unpaid' AND due_date < date('now')`,
	)
	if err != nil {
		return 0, fmt.Errorf("update overdue statuses: %w", err)
	}
	count, _ := result.RowsAffected()
	if count > 0 {
		log.Printf("Updated %d invoices to overdue status", count)
	}
	return count, nil
}

func (s *InvoiceService) MarkAsPaid(ctx context.Context, id int64) (*models.Invoice, error) {
	var currentStatus string
	err := s.db.QueryRowContext(ctx, "SELECT status FROM invoices WHERE id = ?", id).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query invoice status: %w", err)
	}
	if currentStatus == "paid" {
		return nil, fmt.Errorf("invoice is already paid")
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE invoices SET status = 'paid', updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id,
	)
	if err != nil {
		return nil, fmt.Errorf("mark invoice as paid: %w", err)
	}
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

	where, args := buildInvoiceWhere(params)

	var total int64
	var totalAmount sql.NullInt64
	countQuery := "SELECT COUNT(*), COALESCE(SUM(amount), 0) FROM invoices" + where
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total, &totalAmount); err != nil {
		return nil, fmt.Errorf("count invoices: %w", err)
	}

	offset := (params.Page - 1) * params.PerPage
	query := `SELECT id, vendor_name, amount, issue_date, due_date, status, description, created_at, updated_at
		FROM invoices` + where + ` ORDER BY due_date ASC, id DESC LIMIT ? OFFSET ?`
	args = append(args, params.PerPage, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query invoices: %w", err)
	}
	defer rows.Close()

	var invoices []models.Invoice
	for rows.Next() {
		var inv models.Invoice
		if err := rows.Scan(&inv.ID, &inv.VendorName, &inv.Amount, &inv.IssueDate, &inv.DueDate, &inv.Status, &inv.Description, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan invoice: %w", err)
		}
		invoices = append(invoices, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate invoices: %w", err)
	}

	return &InvoiceListResult{
		Invoices:    invoices,
		Total:       total,
		TotalAmount: totalAmount.Int64,
	}, nil
}

func (s *InvoiceService) GetStats(ctx context.Context) (*models.InvoiceStats, error) {
	if _, err := s.UpdateOverdueStatuses(ctx); err != nil {
		log.Printf("Warning: failed to update overdue statuses: %v", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT status, COUNT(*), COALESCE(SUM(amount), 0) FROM invoices WHERE status IN ('unpaid', 'overdue') GROUP BY status`,
	)
	if err != nil {
		return nil, fmt.Errorf("query invoice stats: %w", err)
	}
	defer rows.Close()

	stats := &models.InvoiceStats{}
	for rows.Next() {
		var status string
		var count int
		var amount int64
		if err := rows.Scan(&status, &count, &amount); err != nil {
			return nil, fmt.Errorf("scan invoice stats: %w", err)
		}
		switch status {
		case "unpaid":
			stats.UnpaidCount = count
			stats.UnpaidAmount = amount
			stats.TotalOutstanding += amount
		case "overdue":
			stats.OverdueCount = count
			stats.OverdueAmount = amount
			stats.TotalOutstanding += amount
		}
	}
	return stats, rows.Err()
}

func buildInvoiceWhere(params InvoiceListParams) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if params.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, params.Status)
	}
	if params.DateFrom != "" {
		conditions = append(conditions, "due_date >= ?")
		args = append(args, params.DateFrom)
	}
	if params.DateTo != "" {
		conditions = append(conditions, "due_date <= ?")
		args = append(args, params.DateTo)
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE "
		for i, c := range conditions {
			if i > 0 {
				where += " AND "
			}
			where += c
		}
	}
	return where, args
}
