package models

import "time"

type Expense struct {
	ID          int64     `json:"id"`
	Amount      int64     `json:"amount"` // stored in minor currency units (cents)
	Currency    string    `json:"currency"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Income struct {
	ID          int64     `json:"id"`
	Amount      int64     `json:"amount"` // stored in minor currency units (cents)
	Currency    string    `json:"currency"`
	Source      string    `json:"source"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Invoice struct {
	ID           int64     `json:"id"`
	Amount       int64     `json:"amount"` // stored in minor currency units (cents)
	Currency     string    `json:"currency"`
	Vendor       string    `json:"vendor"`
	Description  string    `json:"description"`
	DueDate      time.Time `json:"due_date"`
	Status       string    `json:"status"` // pending, paid, overdue
	AttachmentID *string   `json:"attachment_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
