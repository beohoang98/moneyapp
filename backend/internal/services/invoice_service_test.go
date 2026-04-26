package services_test

import (
	"context"
	"testing"

	"github.com/beohoang98/moneyapp/internal/models"
	"github.com/beohoang98/moneyapp/internal/services"
)

func TestInvoiceCreate_Valid(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	inv := &models.Invoice{
		VendorName: "ACME Corp", Amount: 5000000,
		IssueDate: "2026-04-01", DueDate: "2026-04-30",
		Status: "unpaid", Description: "Office supplies",
	}
	if err := is.Create(context.Background(), inv); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestInvoiceCreate_EmptyVendor(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	inv := &models.Invoice{
		VendorName: "", Amount: 5000000,
		IssueDate: "2026-04-01", DueDate: "2026-04-30",
	}
	err := is.Create(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for empty vendor name")
	}
}

func TestInvoiceCreate_ZeroAmount(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	inv := &models.Invoice{
		VendorName: "ACME", Amount: 0,
		IssueDate: "2026-04-01", DueDate: "2026-04-30",
	}
	err := is.Create(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestInvoiceCreate_DueDateBeforeIssueDate(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	inv := &models.Invoice{
		VendorName: "ACME", Amount: 5000000,
		IssueDate: "2026-04-30", DueDate: "2026-04-01",
	}
	err := is.Create(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for due_date before issue_date")
	}
}

func TestInvoiceCreate_DefaultStatus(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	inv := &models.Invoice{
		VendorName: "ACME", Amount: 5000000,
		IssueDate: "2026-04-01", DueDate: "2026-04-30",
	}
	if err := is.Create(context.Background(), inv); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := is.GetByID(context.Background(), inv.ID)
	if got.Status != "unpaid" {
		t.Fatalf("expected default status 'unpaid', got %q", got.Status)
	}
}

func TestInvoiceCreate_InvalidStatus(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	inv := &models.Invoice{
		VendorName: "ACME", Amount: 5000000,
		IssueDate: "2026-04-01", DueDate: "2026-04-30",
		Status: "invalid",
	}
	err := is.Create(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestInvoiceMarkAsPaid(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	inv := &models.Invoice{
		VendorName: "ACME", Amount: 5000000,
		IssueDate: "2026-04-01", DueDate: "2026-04-30",
		Status: "unpaid",
	}
	is.Create(context.Background(), inv)

	paid, err := is.MarkAsPaid(context.Background(), inv.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if paid.Status != "paid" {
		t.Fatalf("expected status 'paid', got %q", paid.Status)
	}
}

func TestInvoiceMarkAsPaid_AlreadyPaid(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	inv := &models.Invoice{
		VendorName: "ACME", Amount: 5000000,
		IssueDate: "2026-04-01", DueDate: "2026-04-30",
		Status: "paid",
	}
	is.Create(context.Background(), inv)

	_, err := is.MarkAsPaid(context.Background(), inv.ID)
	if err == nil {
		t.Fatal("expected error for already paid invoice")
	}
}

func TestInvoiceMarkAsPaid_NotFound(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	_, err := is.MarkAsPaid(context.Background(), 9999)
	if err != services.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestInvoiceUpdateOverdueStatuses(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	// Unpaid, past due date -> should become overdue
	is.Create(context.Background(), &models.Invoice{
		VendorName: "Past Due", Amount: 1000,
		IssueDate: "2020-01-01", DueDate: "2020-01-31",
		Status: "unpaid",
	})
	// Unpaid, future due date -> should stay unpaid
	is.Create(context.Background(), &models.Invoice{
		VendorName: "Future Due", Amount: 2000,
		IssueDate: "2026-04-01", DueDate: "2099-12-31",
		Status: "unpaid",
	})
	// Paid, past due date -> should NOT change
	is.Create(context.Background(), &models.Invoice{
		VendorName: "Paid Past", Amount: 3000,
		IssueDate: "2020-01-01", DueDate: "2020-01-31",
		Status: "paid",
	})

	count, err := is.UpdateOverdueStatuses(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 updated, got %d", count)
	}

	inv1, _ := is.GetByID(context.Background(), 1)
	if inv1.Status != "overdue" {
		t.Fatalf("expected 'overdue', got %q", inv1.Status)
	}

	inv2, _ := is.GetByID(context.Background(), 2)
	if inv2.Status != "unpaid" {
		t.Fatalf("expected 'unpaid', got %q", inv2.Status)
	}

	inv3, _ := is.GetByID(context.Background(), 3)
	if inv3.Status != "paid" {
		t.Fatalf("expected 'paid', got %q", inv3.Status)
	}
}

func TestInvoiceGetStats(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	is.Create(context.Background(), &models.Invoice{
		VendorName: "A", Amount: 1000000,
		IssueDate: "2026-04-01", DueDate: "2099-04-30",
		Status: "unpaid",
	})
	is.Create(context.Background(), &models.Invoice{
		VendorName: "B", Amount: 2000000,
		IssueDate: "2026-04-01", DueDate: "2099-04-30",
		Status: "unpaid",
	})
	is.Create(context.Background(), &models.Invoice{
		VendorName: "C", Amount: 500000,
		IssueDate: "2020-01-01", DueDate: "2020-01-31",
		Status: "unpaid",
	})
	is.Create(context.Background(), &models.Invoice{
		VendorName: "D", Amount: 3000000,
		IssueDate: "2026-04-01", DueDate: "2099-04-30",
		Status: "paid",
	})

	stats, err := is.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After UpdateOverdueStatuses: vendor C becomes overdue
	if stats.UnpaidCount != 2 {
		t.Fatalf("expected 2 unpaid, got %d", stats.UnpaidCount)
	}
	if stats.OverdueCount != 1 {
		t.Fatalf("expected 1 overdue, got %d", stats.OverdueCount)
	}
	expectedOutstanding := int64(1000000 + 2000000 + 500000)
	if stats.TotalOutstanding != expectedOutstanding {
		t.Fatalf("expected outstanding %d, got %d", expectedOutstanding, stats.TotalOutstanding)
	}
}

func TestInvoiceUpdate(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	inv := &models.Invoice{
		VendorName: "ACME", Amount: 5000000,
		IssueDate: "2026-04-01", DueDate: "2026-04-30",
		Status: "unpaid",
	}
	is.Create(context.Background(), inv)

	updated := &models.Invoice{
		VendorName: "ACME Updated", Amount: 6000000,
		IssueDate: "2026-04-01", DueDate: "2026-05-15",
		Status: "unpaid",
	}
	if err := is.Update(context.Background(), inv.ID, updated); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := is.GetByID(context.Background(), inv.ID)
	if got.VendorName != "ACME Updated" {
		t.Fatalf("expected 'ACME Updated', got %q", got.VendorName)
	}
	if got.Amount != 6000000 {
		t.Fatalf("expected 6000000, got %d", got.Amount)
	}
}

func TestInvoiceList_StatusFilter(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	is.Create(context.Background(), &models.Invoice{VendorName: "A", Amount: 1000, IssueDate: "2026-04-01", DueDate: "2099-04-30", Status: "unpaid"})
	is.Create(context.Background(), &models.Invoice{VendorName: "B", Amount: 2000, IssueDate: "2026-04-01", DueDate: "2099-04-30", Status: "paid"})
	is.Create(context.Background(), &models.Invoice{VendorName: "C", Amount: 3000, IssueDate: "2026-04-01", DueDate: "2099-04-30", Status: "unpaid"})

	result, err := is.List(context.Background(), services.InvoiceListParams{Page: 1, PerPage: 20, Status: "unpaid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("expected 2 unpaid, got %d", result.Total)
	}

	result2, _ := is.List(context.Background(), services.InvoiceListParams{Page: 1, PerPage: 20, Status: "paid"})
	if result2.Total != 1 {
		t.Fatalf("expected 1 paid, got %d", result2.Total)
	}

	result3, _ := is.List(context.Background(), services.InvoiceListParams{Page: 1, PerPage: 20})
	if result3.Total != 3 {
		t.Fatalf("expected 3 total, got %d", result3.Total)
	}
}

func TestInvoiceUpdate_NotFound(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	err := is.Update(context.Background(), 9999, &models.Invoice{
		VendorName: "X", Amount: 1000, IssueDate: "2026-04-01", DueDate: "2026-04-30", Status: "unpaid",
	})
	if err != services.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestInvoiceDelete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	err := is.Delete(context.Background(), 9999)
	if err != services.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestInvoiceDelete(t *testing.T) {
	db := setupTestDB(t)
	is := services.NewInvoiceService(db)

	inv := &models.Invoice{
		VendorName: "ACME", Amount: 5000000,
		IssueDate: "2026-04-01", DueDate: "2026-04-30",
		Status: "unpaid",
	}
	is.Create(context.Background(), inv)

	if err := is.Delete(context.Background(), inv.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := is.GetByID(context.Background(), inv.ID)
	if err != services.ErrNotFound {
		t.Fatal("expected ErrNotFound after delete")
	}
}
