package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/beohoang98/moneyapp/internal/models"
	"gorm.io/gorm"
)

type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{db: db}
}

func (s *CategoryService) List(ctx context.Context, categoryType string) ([]models.Category, error) {
	var categories []models.Category
	q := s.db.WithContext(ctx).Order("id ASC")
	if categoryType != "" {
		q = q.Where("type = ?", categoryType)
	}
	if err := q.Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	return categories, nil
}

func (s *CategoryService) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	var c models.Category
	if err := s.db.WithContext(ctx).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query category: %w", err)
	}
	return &c, nil
}

func (s *CategoryService) Create(ctx context.Context, name string, categoryType string) (*models.Category, error) {
	name = trimAndValidate(name)
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}
	if categoryType != "expense" && categoryType != "income" {
		return nil, fmt.Errorf("type must be 'expense' or 'income'")
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Category{}).
		Where("name = ? AND type = ?", name, categoryType).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if count > 0 {
		return nil, ErrDuplicateCategory
	}

	cat := models.Category{Name: name, Type: categoryType, IsDefault: false}
	if err := s.db.WithContext(ctx).Create(&cat).Error; err != nil {
		return nil, fmt.Errorf("insert category: %w", err)
	}
	return &cat, nil
}

func (s *CategoryService) Update(ctx context.Context, id int64, name string) (*models.Category, error) {
	cat, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cat.IsDefault {
		return nil, ErrDefaultCategory
	}

	name = trimAndValidate(name)
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Category{}).
		Where("name = ? AND type = ? AND id != ?", name, cat.Type, id).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if count > 0 {
		return nil, ErrDuplicateCategory
	}

	if err := s.db.WithContext(ctx).Model(cat).Update("name", name).Error; err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}
	return s.GetByID(ctx, id)
}

func (s *CategoryService) Delete(ctx context.Context, id int64) error {
	cat, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if cat.IsDefault {
		return ErrDefaultCategory
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var uncatID int64
		if err := tx.Model(&models.Category{}).
			Where("name = 'Uncategorized' AND type = ? AND is_default = 1", cat.Type).
			Select("id").Row().Scan(&uncatID); err != nil {
			return fmt.Errorf("find Uncategorized category: %w", err)
		}

		if cat.Type == "expense" {
			if err := tx.Model(&models.Expense{}).
				Where("category_id = ?", id).
				Updates(map[string]interface{}{"category_id": uncatID, "updated_at": gorm.Expr("CURRENT_TIMESTAMP")}).Error; err != nil {
				return fmt.Errorf("reassign expenses: %w", err)
			}
		} else {
			if err := tx.Model(&models.Income{}).
				Where("category_id = ?", id).
				Updates(map[string]interface{}{"category_id": uncatID, "updated_at": gorm.Expr("CURRENT_TIMESTAMP")}).Error; err != nil {
				return fmt.Errorf("reassign incomes: %w", err)
			}
		}

		if err := tx.Delete(&models.Category{}, id).Error; err != nil {
			return fmt.Errorf("delete category: %w", err)
		}
		return nil
	})
}

var (
	ErrDuplicateCategory = fmt.Errorf("category already exists")
	ErrDefaultCategory   = fmt.Errorf("default categories cannot be modified")
)

func trimAndValidate(s string) string {
	return strings.TrimSpace(s)
}

func (s *CategoryService) ValidateCategory(ctx context.Context, categoryID int64, expectedType string) error {
	var cat models.Category
	if err := s.db.WithContext(ctx).Select("type").First(&cat, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("category not found")
		}
		return fmt.Errorf("query category: %w", err)
	}
	if cat.Type != expectedType {
		return fmt.Errorf("category is not of type %q", expectedType)
	}
	return nil
}
