package entity

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Account represents a linked account for a user
type Account struct {
	ID uint `json:"id" gorm:"primaryKey"`

	UserID uint `json:"user_id" gorm:"not null"`
	User   User `json:"user" gorm:"foreignKey:UserID"`

	CurrentStatus          string                 `json:"current_status" gorm:"not null"`
	AccountStatusHistories []AccountStatusHistory `json:"account_status_histories" gorm:"foreignKey:AccountID"`

	Provider  string         `json:"provider" gorm:"not null"` // e.g., "linkedin"
	AccountID string         `json:"account_id" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// AccountStatusHistory represents the status history of an account
type AccountStatusHistory struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	AccountID uint `json:"account_id" gorm:"not null"`

	Checkpoint          string          `json:"checkpoint"`
	CheckpointMetadata  json.RawMessage `json:"checkpoint_metadata"`
	CheckpointExpiresAt time.Time       `json:"checkpoint_expires_at"`

	// System status: OK, PENDING
	// Unipile status: OK, ERROR/STOPPED, CREDENTIALS, CONNECTING, DELETED, CREATION_SUCCESS, RECONNECTED, SYNC_SUCCESS
	Status string `json:"status" gorm:"not null"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
