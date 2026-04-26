package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/beohoang98/moneyapp/internal/models"
)

var ErrNotFound = errors.New("not found")

type ExpenseListParams struct {
	Page        int
	PerPage     int
	DateFrom    string
	DateTo      string
	CategoryIDs []int64
}

type ExpenseListResult struct {
	Expenses    []models.Expense
	Total       int64
	TotalAmount int64
}

type ExpenseService struct {
	db              *sql.DB
	categoryService *CategoryService
}

func NewExpenseService(db *sql.DB, cs *CategoryService) *ExpenseService {
	return &ExpenseService{db: db, categoryService: cs}
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

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO expenses (amount, date, category_id, description) VALUES (?, ?, ?, ?)`,
		e.Amount, e.Date, e.CategoryID, e.Description,
	)
	if err != nil {
		return fmt.Errorf("insert expense: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	e.ID = id
	return nil
}

func (s *ExpenseService) GetByID(ctx context.Context, id int64) (*models.Expense, error) {
	var e models.Expense
	err := s.db.QueryRowContext(ctx,
		`SELECT e.id, e.amount, e.date, e.category_id, COALESCE(c.name, ''), e.description, e.created_at, e.updated_at
		 FROM expenses e LEFT JOIN categories c ON e.category_id = c.id WHERE e.id = ?`, id,
	).Scan(&e.ID, &e.Amount, &e.Date, &e.CategoryID, &e.CategoryName, &e.Description, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query expense: %w", err)
	}
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

	result, err := s.db.ExecContext(ctx,
		`UPDATE expenses SET amount = ?, date = ?, category_id = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		e.Amount, e.Date, e.CategoryID, e.Description, id,
	)
	if err != nil {
		return fmt.Errorf("update expense: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *ExpenseService) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM expenses WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
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

	where, args := buildExpenseWhere(params)

	var total int64
	var totalAmount sql.NullInt64
	countQuery := "SELECT COUNT(*), COALESCE(SUM(e.amount), 0) FROM expenses e" + where
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total, &totalAmount); err != nil {
		return nil, fmt.Errorf("count expenses: %w", err)
	}

	offset := (params.Page - 1) * params.PerPage
	query := `SELECT e.id, e.amount, e.date, e.category_id, COALESCE(c.name, ''), e.description, e.created_at, e.updated_at
		FROM expenses e LEFT JOIN categories c ON e.category_id = c.id` + where + ` ORDER BY e.date DESC, e.id DESC LIMIT ? OFFSET ?`
	args = append(args, params.PerPage, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query expenses: %w", err)
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.ID, &e.Amount, &e.Date, &e.CategoryID, &e.CategoryName, &e.Description, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan expense: %w", err)
		}
		expenses = append(expenses, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expenses: %w", err)
	}

	return &ExpenseListResult{
		Expenses:    expenses,
		Total:       total,
		TotalAmount: totalAmount.Int64,
	}, nil
}

func buildExpenseWhere(params ExpenseListParams) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if params.DateFrom != "" {
		conditions = append(conditions, "e.date >= ?")
		args = append(args, params.DateFrom)
	}
	if params.DateTo != "" {
		conditions = append(conditions, "e.date <= ?")
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
		conditions = append(conditions, "e.category_id IN ("+placeholders+")")
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
