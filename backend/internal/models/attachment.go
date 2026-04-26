package models

import "time"

type Attachment struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	EntityType string    `json:"entity_type" gorm:"column:entity_type;not null"`
	EntityID   int64     `json:"entity_id" gorm:"column:entity_id;not null"`
	Filename   string    `json:"filename" gorm:"not null"`
	MimeType   string    `json:"mime_type" gorm:"column:mime_type;not null"`
	SizeBytes  int64     `json:"size_bytes" gorm:"column:size_bytes;not null"`
	StorageKey string    `json:"storage_key" gorm:"column:storage_key;uniqueIndex;not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}
