package services_test

import (
	"context"
	"testing"

	"github.com/beohoang98/moneyapp/internal/models"
	"github.com/beohoang98/moneyapp/internal/services"
)

func TestDashboardSummary_Empty(t *testing.T) {
	db := setupTestDB(t)
	invoiceSvc := services.NewInvoiceService(db)
	ds := services.NewDashboardService(db, invoiceSvc)

	summary, err := ds.GetSummary(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalIncome != 0 {
		t.Fatalf("expected 0 income, got %d", summary.TotalIncome)
	}
	if summary.TotalExpenses != 0 {
		t.Fatalf("expected 0 expenses, got %d", summary.TotalExpenses)
	}
	if summary.NetBalance != 0 {
		t.Fatalf("expected 0 balance, got %d", summary.NetBalance)
	}
}

func TestDashboardSummary_WithData(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)
	is := services.NewIncomeService(db, cs)
	invoiceSvc := services.NewInvoiceService(db)
	ds := services.NewDashboardService(db, invoiceSvc)

	es.Create(context.Background(), &models.Expense{Amount: 100000, Date: "2026-04-15", CategoryID: 1})
	es.Create(context.Background(), &models.Expense{Amount: 200000, Date: "2026-04-20", CategoryID: 1})
	is.Create(context.Background(), &models.Income{Amount: 500000, Date: "2026-04-01", CategoryID: 4})

	summary, err := ds.GetSummary(context.Background(), "2026-04-01", "2026-04-30")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalExpenses != 300000 {
		t.Fatalf("expected expenses 300000, got %d", summary.TotalExpenses)
	}
	if summary.TotalIncome != 500000 {
		t.Fatalf("expected income 500000, got %d", summary.TotalIncome)
	}
	if summary.NetBalance != 200000 {
		t.Fatalf("expected balance 200000, got %d", summary.NetBalance)
	}
}

func TestDashboardSummary_DateRange(t *testing.T) {
	db := setupTestDB(t)
	cs := services.NewCategoryService(db)
	es := services.NewExpenseService(db, cs)
	invoiceSvc := services.NewInvoiceService(db)
	ds := services.NewDashboardService(db, invoiceSvc)

	es.Create(context.Background(), &models.Expense{Amount: 100000, Date: "2026-03-15", CategoryID: 1})
	es.Create(context.Background(), &models.Expense{Amount: 200000, Date: "2026-04-15", CategoryID: 1})

	summary, _ := ds.GetSummary(context.Background(), "2026-04-01", "2026-04-30")
	if summary.TotalExpenses != 200000 {
		t.Fatalf("expected 200000 for April, got %d", summary.TotalExpenses)
	}
}
