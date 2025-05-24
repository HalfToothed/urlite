package models

import (
	"time"

	"gorm.io/gorm"
)

type Url struct {
	ID        uint   `gorm:"primaryKey"`
	ShortCode string `gorm:"uniqueIndex;not null"`
	LongURL   string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt *time.Time     `gorm:"column:expires_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"` // Optional soft delete
}

type ShortenRequest struct {
	URL string `json:"url" binding:"required"`
	TTL string `json:"ttl"`
}
