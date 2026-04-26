package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/beohoang98/moneyapp/internal/models"
)

type CategoryService struct {
	db *sql.DB
}

func NewCategoryService(db *sql.DB) *CategoryService {
	return &CategoryService{db: db}
}

func (s *CategoryService) List(ctx context.Context, categoryType string) ([]models.Category, error) {
	query := "SELECT id, name, type, is_default, color, created_at FROM categories"
	var args []interface{}

	if categoryType != "" {
		query += " WHERE type = ?"
		args = append(args, categoryType)
	}
	query += " ORDER BY id ASC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.IsDefault, &c.Color, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (s *CategoryService) ValidateCategory(ctx context.Context, categoryID int64, expectedType string) error {
	var catType string
	err := s.db.QueryRowContext(ctx,
		"SELECT type FROM categories WHERE id = ?", categoryID,
	).Scan(&catType)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("category not found")
		}
		return fmt.Errorf("query category: %w", err)
	}
	if catType != expectedType {
		return fmt.Errorf("category is not of type %q", expectedType)
	}
	return nil
}
