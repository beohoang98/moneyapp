package services

import (
	"context"
	"fmt"
	"time"

	"github.com/beohoang98/moneyapp/internal/models"
	"gorm.io/gorm"
)

type IncomeListParams struct {
	Page        int
	PerPage     int
	DateFrom    string
	DateTo      string
	CategoryIDs []int64
}

type IncomeListResult struct {
	Incomes     []models.Income
	Total       int64
	TotalAmount int64
}

type IncomeService struct {
	db                *gorm.DB
	categoryService   *CategoryService
	attachmentService *AttachmentService
}

func NewIncomeService(db *gorm.DB, cs *CategoryService) *IncomeService {
	return &IncomeService{db: db, categoryService: cs}
}

func (s *IncomeService) SetAttachmentService(as *AttachmentService) {
	s.attachmentService = as
}

func (s *IncomeService) Create(ctx context.Context, inc *models.Income) error {
	if inc.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if inc.Date == "" {
		inc.Date = time.Now().Format("2006-01-02")
	}
	if err := s.categoryService.ValidateCategory(ctx, inc.CategoryID, "income"); err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Create(inc).Error; err != nil {
		return fmt.Errorf("insert income: %w", err)
	}
	return nil
}

func (s *IncomeService) GetByID(ctx context.Context, id int64) (*models.Income, error) {
	type row struct {
		models.Income
		CategoryName string `gorm:"column:category_name"`
	}
	var r row
	result := s.db.WithContext(ctx).
		Table("incomes").
		Select("incomes.*, COALESCE(categories.name, '') AS category_name").
		Joins("LEFT JOIN categories ON incomes.category_id = categories.id").
		Where("incomes.id = ?", id).
		Scan(&r)
	if result.Error != nil {
		return nil, fmt.Errorf("query income: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	inc := r.Income
	inc.CategoryName = r.CategoryName
	return &inc, nil
}

func (s *IncomeService) Update(ctx context.Context, id int64, inc *models.Income) error {
	if inc.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if inc.Date == "" {
		return fmt.Errorf("date is required")
	}
	if err := s.categoryService.ValidateCategory(ctx, inc.CategoryID, "income"); err != nil {
		return err
	}

	result := s.db.WithContext(ctx).Model(&models.Income{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"amount":      inc.Amount,
			"date":        inc.Date,
			"category_id": inc.CategoryID,
			"description": inc.Description,
			"updated_at":  gorm.Expr("CURRENT_TIMESTAMP"),
		})
	if result.Error != nil {
		return fmt.Errorf("update income: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *IncomeService) Delete(ctx context.Context, id int64) error {
	if s.attachmentService != nil {
		if err := s.attachmentService.DeleteByEntity(ctx, "income", id); err != nil {
			return fmt.Errorf("delete income attachments: %w", err)
		}
	}

	result := s.db.WithContext(ctx).Delete(&models.Income{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete income: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *IncomeService) List(ctx context.Context, params IncomeListParams) (*IncomeListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}

	base := s.db.WithContext(ctx).Model(&models.Income{}).Table("incomes AS i")
	base = applyIncomeFilters(base, params)

	var total int64
	var totalAmount *int64
	countRow := base.Select("COUNT(*), COALESCE(SUM(i.amount), 0)").Row()
	if err := countRow.Scan(&total, &totalAmount); err != nil {
		return nil, fmt.Errorf("count incomes: %w", err)
	}

	offset := (params.Page - 1) * params.PerPage

	type incomeRow struct {
		models.Income
		CategoryName string `gorm:"column:category_name"`
	}
	var rows []incomeRow
	err := s.db.WithContext(ctx).Table("incomes AS i").
		Select("i.*, COALESCE(c.name, '') AS category_name").
		Joins("LEFT JOIN categories c ON i.category_id = c.id").
		Scopes(func(q *gorm.DB) *gorm.DB { return applyIncomeFilters(q, params) }).
		Order("i.date DESC, i.id DESC").
		Limit(params.PerPage).Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("query incomes: %w", err)
	}

	incomes := make([]models.Income, len(rows))
	for i, r := range rows {
		incomes[i] = r.Income
		incomes[i].CategoryName = r.CategoryName
	}

	amt := int64(0)
	if totalAmount != nil {
		amt = *totalAmount
	}
	return &IncomeListResult{
		Incomes:     incomes,
		Total:       total,
		TotalAmount: amt,
	}, nil
}

func applyIncomeFilters(q *gorm.DB, params IncomeListParams) *gorm.DB {
	if params.DateFrom != "" {
		q = q.Where("i.date >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		q = q.Where("i.date <= ?", params.DateTo)
	}
	if len(params.CategoryIDs) > 0 {
		q = q.Where("i.category_id IN ?", params.CategoryIDs)
	}
	return q
}
