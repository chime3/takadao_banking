package models

import (
	"time"

	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypeTransfer TransactionType = "transfer"
)

type Transaction struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID       `gorm:"type:uuid;not null;index" json:"user_id"`
	Type        TransactionType `gorm:"type:varchar(20);not null" json:"type"`
	Amount      float64         `gorm:"type:decimal(20,2);not null" json:"amount"`
	Currency    string          `gorm:"type:varchar(3);not null" json:"currency"`
	RecipientID *uuid.UUID      `gorm:"type:uuid;index" json:"recipient_id,omitempty"`
	Description string          `gorm:"type:text" json:"description"`
	Status      string          `gorm:"type:varchar(20);not null;default:'completed'" json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	User      User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Recipient *User `gorm:"foreignKey:RecipientID" json:"recipient,omitempty"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// Validate checks if the transaction is valid
func (t *Transaction) Validate() error {
	if t.Amount <= 0 {
		return ErrInvalidAmount
	}

	if t.Type == TransactionTypeTransfer && t.RecipientID == nil {
		return ErrMissingRecipient
	}

	return nil
}

// Custom errors
var (
	ErrInvalidAmount    = errors.New("invalid amount")
	ErrMissingRecipient = errors.New("recipient is required for transfer")
)
