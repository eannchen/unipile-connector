package entity

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Account represents a linked account for a user
type Account struct {
	ID uint `json:"id"`

	UserID uint `json:"user_id"`
	User   User `json:"user" gorm:"foreignKey:UserID"`

	CurrentStatus          string                 `json:"current_status"`
	AccountStatusHistories []AccountStatusHistory `json:"account_status_histories" gorm:"foreignKey:AccountID"`

	Provider  string         `json:"provider"`   // e.g., "LINKEDIN"
	AccountID string         `json:"account_id"` // Account ID from Unipile
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

// AccountStatusHistory represents the status history of an account
type AccountStatusHistory struct {
	ID        uint `json:"id"`
	AccountID uint `json:"account_id"`

	Checkpoint          string          `json:"checkpoint"`
	CheckpointMetadata  json.RawMessage `json:"checkpoint_metadata"`
	CheckpointExpiresAt time.Time       `json:"checkpoint_expires_at"`

	// System status: OK, PENDING
	// Unipile status: OK, ERROR/STOPPED, CREDENTIALS, CONNECTING, DELETED, CREATION_SUCCESS, RECONNECTED, SYNC_SUCCESS
	Status string `json:"status"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`
}

// AccountWithStatus represents an account with its current status
type AccountWithStatus struct {
	Account
	CurrentStatus       string    `json:"current_status"`
	Checkpoint          string    `json:"checkpoint"`
	CheckpointExpiresAt time.Time `json:"checkpoint_expires_at"`
}
