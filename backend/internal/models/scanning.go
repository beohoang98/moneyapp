package models

import "time"

type ScanningSettings struct {
	ID        int64     `json:"-" gorm:"primaryKey"`
	Enabled   bool      `json:"enabled" gorm:"not null;default:0"`
	BaseURL   string    `json:"base_url" gorm:"column:base_url;not null;default:'http://localhost:11434/v1'"`
	Model     string    `json:"model" gorm:"not null;default:'qwen3-vl:4b'"`
	APIKey    string    `json:"-" gorm:"column:api_key"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ScanningSettings) TableName() string { return "scanning_settings" }

type ScanResult struct {
	Vendor      string            `json:"vendor"`
	Date        string            `json:"date"`
	Currency    string            `json:"currency"`
	TotalAmount int64             `json:"total_amount"`
	LineItems   []LineItem        `json:"line_items"`
	Confidence  map[string]string `json:"confidence"`
}

type LineItem struct {
	Description string `json:"description"`
	Amount      int64  `json:"amount"`
}
