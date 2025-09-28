package entity

import (
	"time"

	"gorm.io/gorm"
)

// Account represents a linked account for a user
type Account struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Provider  string         `json:"provider" gorm:"not null"` // e.g., "linkedin"
	AccountID string         `json:"account_id" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Foreign key relationship
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// UnipileConnectRequest represents request to connect LinkedIn account via Unipile
type UnipileConnectRequest struct {
	Type     string `json:"type" binding:"required"` // "credentials" or "cookie"
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Cookie   string `json:"cookie,omitempty"`
}

// UnipileConnectResponse represents response from Unipile connection
type UnipileConnectResponse struct {
	AccountID string `json:"account_id"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
}

