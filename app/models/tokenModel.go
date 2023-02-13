package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Token struct {
	Id           uuid.UUID      `json:"id" gorm:"primary_key"`
	UserId       int32          `json:"user_id"`
	RefreshToken string         `json:"refresh_token"`
	PlatformId   int32          `json:"platform_id"`
	IsBlocked    bool           `json:"is_blocked"`
	ExpiresAt    time.Time      `json:"expires_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at"`
}
