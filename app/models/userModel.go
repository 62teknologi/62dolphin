package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	Id        int32          `json:"id" gorm:"primary_key"`
	Email     string         `json:"email" gorm:"unique"`
	Username  string         `json:"username" gorm:"unique"`
	Password  string         `json:"password"`
	IsActive  bool           `json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}
