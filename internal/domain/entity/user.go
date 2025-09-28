package entity

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"` // Hidden from JSON
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// LinkedInCredentials represents LinkedIn login credentials
type LinkedInCredentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LinkedInCookie represents LinkedIn cookie authentication
type LinkedInCookie struct {
	Cookie string `json:"cookie" binding:"required"`
}

// AuthRequest represents authentication request
type AuthRequest struct {
	Type     string `json:"type" binding:"required"` // "credentials" or "cookie"
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Cookie   string `json:"cookie,omitempty"`
}
