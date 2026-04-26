package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beohoang98/moneyapp/internal/models"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type ExpenseListParams struct {
	Page        int
	PerPage     int
	DateFrom    string
	DateTo      string
	CategoryIDs []int64
	SortBy      string
	SortOrder   string
}

type ExpenseListResult struct {
	Expenses    []models.Expense
	Total       int64
	TotalAmount int64
}

type ExpenseService struct {
	db                *gorm.DB
	categoryService   *CategoryService
	attachmentService *AttachmentService
}

func NewExpenseService(db *gorm.DB, cs *CategoryService) *ExpenseService {
	return &ExpenseService{db: db, categoryService: cs}
}

func (s *ExpenseService) SetAttachmentService(as *AttachmentService) {
	s.attachmentService = as
}

func (s *ExpenseService) Create(ctx context.Context, e *models.Expense) error {
	if e.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if e.Date == "" {
		e.Date = time.Now().Format("2006-01-02")
	}
	if err := s.categoryService.ValidateCategory(ctx, e.CategoryID, "expense"); err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Create(e).Error; err != nil {
		return fmt.Errorf("insert expense: %w", err)
	}
	return nil
}

func (s *ExpenseService) GetByID(ctx context.Context, id int64) (*models.Expense, error) {
	type row struct {
		models.Expense
		CategoryName string `gorm:"column:category_name"`
	}
	var r row
	result := s.db.WithContext(ctx).
		Table("expenses").
		Select("expenses.*, COALESCE(categories.name, '') AS category_name").
		Joins("LEFT JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.id = ?", id).
		Scan(&r)
	if result.Error != nil {
		return nil, fmt.Errorf("query expense: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	e := r.Expense
	e.CategoryName = r.CategoryName
	return &e, nil
}

func (s *ExpenseService) Update(ctx context.Context, id int64, e *models.Expense) error {
	if e.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if e.Date == "" {
		return fmt.Errorf("date is required")
	}
	if err := s.categoryService.ValidateCategory(ctx, e.CategoryID, "expense"); err != nil {
		return err
	}

	result := s.db.WithContext(ctx).Model(&models.Expense{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"amount":      e.Amount,
			"date":        e.Date,
			"category_id": e.CategoryID,
			"description": e.Description,
			"updated_at":  gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		return fmt.Errorf("update expense: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *ExpenseService) Delete(ctx context.Context, id int64) error {
	if s.attachmentService != nil {
		if err := s.attachmentService.DeleteByEntity(ctx, "expense", id); err != nil {
			return fmt.Errorf("delete expense attachments: %w", err)
		}
	}

	result := s.db.WithContext(ctx).Delete(&models.Expense{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete expense: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *ExpenseService) List(ctx context.Context, params ExpenseListParams) (*ExpenseListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}

	base := s.db.WithContext(ctx).Model(&models.Expense{}).Table("expenses AS e")
	base = applyExpenseFilters(base, params)

	var total int64
	var totalAmount *int64
	countRow := base.Select("COUNT(*), COALESCE(SUM(e.amount), 0)").Row()
	if err := countRow.Scan(&total, &totalAmount); err != nil {
		return nil, fmt.Errorf("count expenses: %w", err)
	}

	offset := (params.Page - 1) * params.PerPage
	orderClause := buildExpenseOrderClause(params)

	type expenseRow struct {
		models.Expense
		CategoryName string `gorm:"column:category_name"`
	}
	var rows []expenseRow
	err := s.db.WithContext(ctx).Table("expenses AS e").
		Select("e.*, COALESCE(c.name, '') AS category_name").
		Joins("LEFT JOIN categories c ON e.category_id = c.id").
		Scopes(func(q *gorm.DB) *gorm.DB { return applyExpenseFilters(q, params) }).
		Order(orderClause).
		Limit(params.PerPage).Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("query expenses: %w", err)
	}

	expenses := make([]models.Expense, len(rows))
	for i, r := range rows {
		expenses[i] = r.Expense
		expenses[i].CategoryName = r.CategoryName
	}

	amt := int64(0)
	if totalAmount != nil {
		amt = *totalAmount
	}
	return &ExpenseListResult{
		Expenses:    expenses,
		Total:       total,
		TotalAmount: amt,
	}, nil
}

func applyExpenseFilters(q *gorm.DB, params ExpenseListParams) *gorm.DB {
	if params.DateFrom != "" {
		q = q.Where("e.date >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		q = q.Where("e.date <= ?", params.DateTo)
	}
	if len(params.CategoryIDs) > 0 {
		q = q.Where("e.category_id IN ?", params.CategoryIDs)
	}
	return q
}

func buildExpenseOrderClause(params ExpenseListParams) string {
	col := "e.date"
	switch params.SortBy {
	case "amount":
		col = "e.amount"
	case "date":
		col = "e.date"
	}

	dir := "DESC"
	if params.SortOrder == "asc" {
		dir = "ASC"
	}

	return fmt.Sprintf("%s %s, e.id DESC", col, dir)
}
