package models

import "time"

type User struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type Category struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"not null"`
	Type      string    `json:"type" gorm:"not null"`
	IsDefault bool      `json:"is_default" gorm:"column:is_default;not null;default:0"`
	Color     *string   `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type Expense struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Amount       int64     `json:"amount" gorm:"not null"`
	Date         string    `json:"date" gorm:"not null"`
	CategoryID   int64     `json:"category_id" gorm:"not null"`
	CategoryName string    `json:"category_name,omitempty" gorm:"-"`
	Description  string    `json:"description" gorm:"default:''"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Category *Category `json:"-" gorm:"foreignKey:CategoryID"`
}

type Income struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Amount       int64     `json:"amount" gorm:"not null"`
	Date         string    `json:"date" gorm:"not null"`
	CategoryID   int64     `json:"category_id" gorm:"not null"`
	CategoryName string    `json:"category_name,omitempty" gorm:"-"`
	Description  string    `json:"description" gorm:"default:''"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Category *Category `json:"-" gorm:"foreignKey:CategoryID"`
}

type Invoice struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	VendorName  string    `json:"vendor_name" gorm:"column:vendor_name;not null"`
	Amount      int64     `json:"amount" gorm:"not null"`
	IssueDate   string    `json:"issue_date" gorm:"column:issue_date;not null"`
	DueDate     string    `json:"due_date" gorm:"column:due_date;not null"`
	Status      string    `json:"status" gorm:"not null;default:'unpaid'"`
	Description string    `json:"description" gorm:"default:''"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type DashboardSummary struct {
	TotalIncome      int64  `json:"total_income"`
	TotalExpenses    int64  `json:"total_expenses"`
	NetBalance       int64  `json:"net_balance"`
	DateFrom         string `json:"date_from"`
	DateTo           string `json:"date_to"`
	UnpaidCount      int    `json:"unpaid_count"`
	UnpaidAmount     int64  `json:"unpaid_amount"`
	OverdueCount     int    `json:"overdue_count"`
	OverdueAmount    int64  `json:"overdue_amount"`
	TotalOutstanding int64  `json:"total_outstanding"`
}

type InvoiceStats struct {
	TotalOutstanding int64 `json:"total_outstanding"`
	UnpaidCount      int   `json:"unpaid_count"`
	UnpaidAmount     int64 `json:"unpaid_amount"`
	OverdueCount     int   `json:"overdue_count"`
	OverdueAmount    int64 `json:"overdue_amount"`
}

type MonthlyTrendItem struct {
	Month         string `json:"month"`
	TotalIncome   int64  `json:"total_income"`
	TotalExpenses int64  `json:"total_expenses"`
}

type CategoryBreakdownItem struct {
	CategoryName string `json:"category_name"`
	Total        int64  `json:"total"`
}
