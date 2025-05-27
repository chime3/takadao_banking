package models

import (
	"time"

	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Balance struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_user_currency" json:"user_id"`
	Currency  string         `gorm:"type:varchar(3);not null;uniqueIndex:idx_user_currency" json:"currency"`
	Amount    float64        `gorm:"type:decimal(20,2);not null;default:0" json:"amount"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (b *Balance) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// Add adds amount to the balance
func (b *Balance) Add(amount float64) {
	b.Amount += amount
}

// Subtract subtracts amount from the balance
func (b *Balance) Subtract(amount float64) error {
	if b.Amount < amount {
		return ErrInsufficientFunds
	}
	b.Amount -= amount
	return nil
}

// GetBalanceAtTime returns the balance at a specific point in time
// This is used for historical balance queries
type BalanceSnapshot struct {
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Currency  string    `gorm:"type:varchar(3);not null" json:"currency"`
	Amount    float64   `gorm:"type:decimal(20,2);not null" json:"amount"`
	Timestamp time.Time `gorm:"not null" json:"timestamp"`
}

// Custom errors
var (
	ErrInsufficientFunds = errors.New("insufficient funds")
)
