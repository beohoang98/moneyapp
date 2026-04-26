package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	IsDefault bool      `json:"is_default"`
	Color     *string   `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Expense struct {
	ID           int64     `json:"id"`
	Amount       int64     `json:"amount"`
	Date         string    `json:"date"`
	CategoryID   int64     `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Income struct {
	ID           int64     `json:"id"`
	Amount       int64     `json:"amount"`
	Date         string    `json:"date"`
	CategoryID   int64     `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Invoice struct {
	ID          int64     `json:"id"`
	VendorName  string    `json:"vendor_name"`
	Amount      int64     `json:"amount"`
	IssueDate   string    `json:"issue_date"`
	DueDate     string    `json:"due_date"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DashboardSummary struct {
	TotalIncome    int64  `json:"total_income"`
	TotalExpenses  int64  `json:"total_expenses"`
	NetBalance     int64  `json:"net_balance"`
	DateFrom       string `json:"date_from"`
	DateTo         string `json:"date_to"`
	UnpaidCount    int    `json:"unpaid_count"`
	UnpaidAmount   int64  `json:"unpaid_amount"`
	OverdueCount   int    `json:"overdue_count"`
	OverdueAmount  int64  `json:"overdue_amount"`
	TotalOutstanding int64 `json:"total_outstanding"`
}

type InvoiceStats struct {
	TotalOutstanding int64 `json:"total_outstanding"`
	UnpaidCount      int   `json:"unpaid_count"`
	UnpaidAmount     int64 `json:"unpaid_amount"`
	OverdueCount     int   `json:"overdue_count"`
	OverdueAmount    int64 `json:"overdue_amount"`
}
