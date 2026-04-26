package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/beohoang98/moneyapp/internal/models"
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
	db              *sql.DB
	categoryService *CategoryService
}

func NewIncomeService(db *sql.DB, cs *CategoryService) *IncomeService {
	return &IncomeService{db: db, categoryService: cs}
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

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO incomes (amount, date, category_id, description) VALUES (?, ?, ?, ?)`,
		inc.Amount, inc.Date, inc.CategoryID, inc.Description,
	)
	if err != nil {
		return fmt.Errorf("insert income: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	inc.ID = id
	return nil
}

func (s *IncomeService) GetByID(ctx context.Context, id int64) (*models.Income, error) {
	var inc models.Income
	err := s.db.QueryRowContext(ctx,
		`SELECT i.id, i.amount, i.date, i.category_id, COALESCE(c.name, ''), i.description, i.created_at, i.updated_at
		 FROM incomes i LEFT JOIN categories c ON i.category_id = c.id WHERE i.id = ?`, id,
	).Scan(&inc.ID, &inc.Amount, &inc.Date, &inc.CategoryID, &inc.CategoryName, &inc.Description, &inc.CreatedAt, &inc.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query income: %w", err)
	}
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

	result, err := s.db.ExecContext(ctx,
		`UPDATE incomes SET amount = ?, date = ?, category_id = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		inc.Amount, inc.Date, inc.CategoryID, inc.Description, id,
	)
	if err != nil {
		return fmt.Errorf("update income: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *IncomeService) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM incomes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete income: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
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

	where, args := buildIncomeWhere(params)

	var total int64
	var totalAmount sql.NullInt64
	countQuery := "SELECT COUNT(*), COALESCE(SUM(i.amount), 0) FROM incomes i" + where
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total, &totalAmount); err != nil {
		return nil, fmt.Errorf("count incomes: %w", err)
	}

	offset := (params.Page - 1) * params.PerPage
	query := `SELECT i.id, i.amount, i.date, i.category_id, COALESCE(c.name, ''), i.description, i.created_at, i.updated_at
		FROM incomes i LEFT JOIN categories c ON i.category_id = c.id` + where + ` ORDER BY i.date DESC, i.id DESC LIMIT ? OFFSET ?`
	args = append(args, params.PerPage, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query incomes: %w", err)
	}
	defer rows.Close()

	var incomes []models.Income
	for rows.Next() {
		var inc models.Income
		if err := rows.Scan(&inc.ID, &inc.Amount, &inc.Date, &inc.CategoryID, &inc.CategoryName, &inc.Description, &inc.CreatedAt, &inc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan income: %w", err)
		}
		incomes = append(incomes, inc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate incomes: %w", err)
	}

	return &IncomeListResult{
		Incomes:     incomes,
		Total:       total,
		TotalAmount: totalAmount.Int64,
	}, nil
}

func buildIncomeWhere(params IncomeListParams) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if params.DateFrom != "" {
		conditions = append(conditions, "i.date >= ?")
		args = append(args, params.DateFrom)
	}
	if params.DateTo != "" {
		conditions = append(conditions, "i.date <= ?")
		args = append(args, params.DateTo)
	}
	if len(params.CategoryIDs) > 0 {
		placeholders := ""
		for i, id := range params.CategoryIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, id)
		}
		conditions = append(conditions, "i.category_id IN ("+placeholders+")")
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
