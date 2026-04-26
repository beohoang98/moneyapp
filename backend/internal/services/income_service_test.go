package services_test

import (
	"context"
	"testing"

	"github.com/beohoang98/moneyapp/internal/models"
	"github.com/beohoang98/moneyapp/internal/services"
)

func TestIncomeCreate_Valid(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	is := services.NewIncomeService(db, cs)

	inc := &models.Income{Amount: 10000000, Date: "2026-04-26", CategoryID: 4, Description: "April salary"}
	if err := is.Create(context.Background(), inc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inc.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestIncomeCreate_ZeroAmount(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	is := services.NewIncomeService(db, cs)

	inc := &models.Income{Amount: 0, Date: "2026-04-26", CategoryID: 4}
	err := is.Create(context.Background(), inc)
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestIncomeCreate_NegativeAmount(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	is := services.NewIncomeService(db, cs)

	inc := &models.Income{Amount: -100, Date: "2026-04-26", CategoryID: 4}
	err := is.Create(context.Background(), inc)
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestIncomeCreate_CategoryTypeMismatch(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	is := services.NewIncomeService(db, cs)

	// CategoryID 1 = "Food" (expense type)
	inc := &models.Income{Amount: 10000, Date: "2026-04-26", CategoryID: 1}
	err := is.Create(context.Background(), inc)
	if err == nil {
		t.Fatal("expected error when using expense category for income")
	}
}

func TestIncomeCreate_InvalidCategoryID(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	is := services.NewIncomeService(db, cs)

	inc := &models.Income{Amount: 10000, Date: "2026-04-26", CategoryID: 9999}
	err := is.Create(context.Background(), inc)
	if err == nil {
		t.Fatal("expected error for invalid category ID")
	}
}

func TestIncomeGetByID(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	is := services.NewIncomeService(db, cs)

	inc := &models.Income{Amount: 10000000, Date: "2026-04-26", CategoryID: 4}
	is.Create(context.Background(), inc)

	got, err := is.GetByID(context.Background(), inc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Amount != 10000000 {
		t.Fatalf("expected amount 10000000, got %d", got.Amount)
	}
}

func TestIncomeUpdate(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	is := services.NewIncomeService(db, cs)

	inc := &models.Income{Amount: 10000000, Date: "2026-04-26", CategoryID: 4}
	is.Create(context.Background(), inc)

	updated := &models.Income{Amount: 12000000, Date: "2026-04-27", CategoryID: 5, Description: "updated"}
	if err := is.Update(context.Background(), inc.ID, updated); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := is.GetByID(context.Background(), inc.ID)
	if got.Amount != 12000000 {
		t.Fatalf("expected 12000000, got %d", got.Amount)
	}
}

func TestIncomeDelete(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	is := services.NewIncomeService(db, cs)

	inc := &models.Income{Amount: 10000000, Date: "2026-04-26", CategoryID: 4}
	is.Create(context.Background(), inc)

	if err := is.Delete(context.Background(), inc.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := is.GetByID(context.Background(), inc.ID)
	if err != services.ErrNotFound {
		t.Fatal("expected ErrNotFound after delete")
	}
}

func TestIncomeList_DateFilter(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	is := services.NewIncomeService(db, cs)

	is.Create(context.Background(), &models.Income{Amount: 1000, Date: "2026-01-15", CategoryID: 4})
	is.Create(context.Background(), &models.Income{Amount: 2000, Date: "2026-02-15", CategoryID: 4})

	result, _ := is.List(context.Background(), services.IncomeListParams{
		Page: 1, PerPage: 20, DateFrom: "2026-01-01", DateTo: "2026-01-31",
	})
	if result.Total != 1 {
		t.Fatalf("expected 1 income in January, got %d", result.Total)
	}
}
