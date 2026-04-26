package services_test

import (
	"context"
	"testing"

	"github.com/beohoang98/moneyapp/internal/services"
)

func TestCategoryList_All(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)

	cats, err := cs.List(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cats) != 6 {
		t.Fatalf("expected 6 categories (3 expense + 3 income), got %d", len(cats))
	}
}

func TestCategoryList_ByType(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)

	expenseCats, _ := cs.List(context.Background(), "expense")
	if len(expenseCats) != 3 {
		t.Fatalf("expected 3 expense categories, got %d", len(expenseCats))
	}

	incomeCats, _ := cs.List(context.Background(), "income")
	if len(incomeCats) != 3 {
		t.Fatalf("expected 3 income categories, got %d", len(incomeCats))
	}
}

func TestCategoryValidate_Valid(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)

	if err := cs.ValidateCategory(context.Background(), 1, "expense"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCategoryValidate_TypeMismatch(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)

	err := cs.ValidateCategory(context.Background(), 1, "income")
	if err == nil {
		t.Fatal("expected error for type mismatch")
	}
}

func TestCategoryValidate_NotFound(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)

	err := cs.ValidateCategory(context.Background(), 9999, "expense")
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
}
