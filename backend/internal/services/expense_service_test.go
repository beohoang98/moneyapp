package services_test

import (
	"context"
	"testing"

	"github.com/beohoang98/moneyapp/internal/models"
	"github.com/beohoang98/moneyapp/internal/services"
)

func TestExpenseCreate_Valid(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	exp := &models.Expense{Amount: 50000, Date: "2026-04-26", CategoryID: 1, Description: "Lunch"}
	if err := es.Create(context.Background(), exp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestExpenseCreate_ZeroAmount(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	exp := &models.Expense{Amount: 0, Date: "2026-04-26", CategoryID: 1}
	err := es.Create(context.Background(), exp)
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestExpenseCreate_NegativeAmount(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	exp := &models.Expense{Amount: -100, Date: "2026-04-26", CategoryID: 1}
	err := es.Create(context.Background(), exp)
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestExpenseCreate_InvalidCategoryID(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	exp := &models.Expense{Amount: 50000, Date: "2026-04-26", CategoryID: 9999}
	err := es.Create(context.Background(), exp)
	if err == nil {
		t.Fatal("expected error for invalid category ID")
	}
}

func TestExpenseCreate_IncomeCategoryForExpense(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	// CategoryID 4 = "Salary" (income type) in test setup
	exp := &models.Expense{Amount: 50000, Date: "2026-04-26", CategoryID: 4}
	err := es.Create(context.Background(), exp)
	if err == nil {
		t.Fatal("expected error when using income category for expense")
	}
}

func TestExpenseCreate_DefaultsDateWhenEmpty(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	exp := &models.Expense{Amount: 50000, Date: "", CategoryID: 1}
	if err := es.Create(context.Background(), exp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp.Date == "" {
		t.Fatal("expected default date on expense")
	}
	if len(exp.Date) != 10 {
		t.Fatalf("expected YYYY-MM-DD date, got %q", exp.Date)
	}
}

func TestExpenseGetByID(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	exp := &models.Expense{Amount: 50000, Date: "2026-04-26", CategoryID: 1}
	es.Create(context.Background(), exp)

	got, err := es.GetByID(context.Background(), exp.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Amount != 50000 {
		t.Fatalf("expected amount 50000, got %d", got.Amount)
	}
	if got.CategoryName != "Food" {
		t.Fatalf("expected category_name Food, got %s", got.CategoryName)
	}
}

func TestExpenseGetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	_, err := es.GetByID(context.Background(), 9999)
	if err != services.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestExpenseUpdate(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	exp := &models.Expense{Amount: 50000, Date: "2026-04-26", CategoryID: 1}
	es.Create(context.Background(), exp)

	updated := &models.Expense{Amount: 75000, Date: "2026-04-27", CategoryID: 2, Description: "updated"}
	if err := es.Update(context.Background(), exp.ID, updated); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := es.GetByID(context.Background(), exp.ID)
	if got.Amount != 75000 {
		t.Fatalf("expected 75000, got %d", got.Amount)
	}
}

func TestExpenseDelete(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	exp := &models.Expense{Amount: 50000, Date: "2026-04-26", CategoryID: 1}
	es.Create(context.Background(), exp)

	if err := es.Delete(context.Background(), exp.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := es.GetByID(context.Background(), exp.ID)
	if err != services.ErrNotFound {
		t.Fatal("expected ErrNotFound after delete")
	}
}

func TestExpenseDelete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	err := es.Delete(context.Background(), 9999)
	if err != services.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestExpenseList(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	for i := 0; i < 25; i++ {
		exp := &models.Expense{Amount: int64(1000 * (i + 1)), Date: "2026-04-26", CategoryID: 1}
		es.Create(context.Background(), exp)
	}

	result, err := es.List(context.Background(), services.ExpenseListParams{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 25 {
		t.Fatalf("expected total 25, got %d", result.Total)
	}
	if len(result.Expenses) != 20 {
		t.Fatalf("expected 20 expenses on page 1, got %d", len(result.Expenses))
	}

	result2, _ := es.List(context.Background(), services.ExpenseListParams{Page: 2, PerPage: 20})
	if len(result2.Expenses) != 5 {
		t.Fatalf("expected 5 expenses on page 2, got %d", len(result2.Expenses))
	}
}

func TestExpenseList_DateFilter(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	es.Create(context.Background(), &models.Expense{Amount: 1000, Date: "2026-01-15", CategoryID: 1})
	es.Create(context.Background(), &models.Expense{Amount: 2000, Date: "2026-02-15", CategoryID: 1})
	es.Create(context.Background(), &models.Expense{Amount: 3000, Date: "2026-03-15", CategoryID: 1})

	result, _ := es.List(context.Background(), services.ExpenseListParams{
		Page: 1, PerPage: 20, DateFrom: "2026-01-01", DateTo: "2026-01-31",
	})
	if result.Total != 1 {
		t.Fatalf("expected 1 expense in January, got %d", result.Total)
	}
	if result.TotalAmount != 1000 {
		t.Fatalf("expected total_amount 1000, got %d", result.TotalAmount)
	}
}

func TestExpenseList_CategoryFilter(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)

	es.Create(context.Background(), &models.Expense{Amount: 1000, Date: "2026-04-01", CategoryID: 1})
	es.Create(context.Background(), &models.Expense{Amount: 2000, Date: "2026-04-01", CategoryID: 2})
	es.Create(context.Background(), &models.Expense{Amount: 3000, Date: "2026-04-01", CategoryID: 1})

	result, _ := es.List(context.Background(), services.ExpenseListParams{
		Page: 1, PerPage: 20, CategoryIDs: []int64{1},
	})
	if result.Total != 2 {
		t.Fatalf("expected 2 expenses for category 1, got %d", result.Total)
	}
	if result.TotalAmount != 4000 {
		t.Fatalf("expected total_amount 4000, got %d", result.TotalAmount)
	}
}
