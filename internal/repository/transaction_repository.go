package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/takadao/banking/internal/models"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create creates a new transaction
func (r *TransactionRepository) Create(tx *models.Transaction) error {
	return r.db.Transaction(func(db *gorm.DB) error {
		// Create the transaction record
		if err := db.Create(tx).Error; err != nil {
			return err
		}

		// Update sender's balance
		if tx.Type == models.TransactionTypeWithdraw || tx.Type == models.TransactionTypeTransfer {
			var senderBalance models.Balance
			if err := db.Where("user_id = ? AND currency = ?", tx.UserID, tx.Currency).First(&senderBalance).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return models.ErrInsufficientFunds
				}
				return err
			}

			if err := senderBalance.Subtract(tx.Amount); err != nil {
				return err
			}

			if err := db.Save(&senderBalance).Error; err != nil {
				return err
			}
		}

		// Update recipient's balance for transfers
		if tx.Type == models.TransactionTypeTransfer && tx.RecipientID != nil {
			var recipientBalance models.Balance
			err := db.Where("user_id = ? AND currency = ?", tx.RecipientID, tx.Currency).First(&recipientBalance).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					recipientBalance = models.Balance{
						UserID:   *tx.RecipientID,
						Currency: tx.Currency,
						Amount:   0,
					}
				} else {
					return err
				}
			}

			recipientBalance.Add(tx.Amount)
			if err := db.Save(&recipientBalance).Error; err != nil {
				return err
			}
		}

		// Update balance for deposits
		if tx.Type == models.TransactionTypeDeposit {
			var balance models.Balance
			err := db.Where("user_id = ? AND currency = ?", tx.UserID, tx.Currency).First(&balance).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					balance = models.Balance{
						UserID:   tx.UserID,
						Currency: tx.Currency,
						Amount:   0,
					}
				} else {
					return err
				}
			}

			balance.Add(tx.Amount)
			if err := db.Save(&balance).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetByID retrieves a transaction by ID
func (r *TransactionRepository) GetByID(id uuid.UUID) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.Preload("User").Preload("Recipient").First(&tx, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// GetByUserID retrieves transactions for a user with pagination
func (r *TransactionRepository) GetByUserID(userID uuid.UUID, page, pageSize int) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	err := r.db.Model(&models.Transaction{}).Where("user_id = ? OR recipient_id = ?", userID, userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = r.db.Preload("User").Preload("Recipient").
		Where("user_id = ? OR recipient_id = ?", userID, userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetAll retrieves all transactions with pagination
func (r *TransactionRepository) GetAll(page, pageSize int) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	err := r.db.Model(&models.Transaction{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = r.db.Preload("User").Preload("Recipient").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetBalanceAtTime retrieves a user's balance at a specific point in time
func (r *TransactionRepository) GetBalanceAtTime(userID uuid.UUID, currency string, atTime time.Time) (float64, error) {
	var balance float64

	// Calculate balance by summing all transactions up to the specified time
	err := r.db.Model(&models.Transaction{}).
		Select("COALESCE(SUM(CASE WHEN type = 'deposit' OR (type = 'transfer' AND recipient_id = ?) THEN amount ELSE -amount END), 0)", userID).
		Where("(user_id = ? OR recipient_id = ?) AND currency = ? AND created_at <= ?", userID, userID, currency, atTime).
		Scan(&balance).Error

	if err != nil {
		return 0, err
	}

	return balance, nil
}

// GetByIDAndUserID retrieves a transaction by ID and user ID
func (r *TransactionRepository) GetByIDAndUserID(transactionID, userID uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := r.db.Where("id = ? AND user_id = ?", transactionID, userID).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

// ListByUserID retrieves transactions for a user
func (r *TransactionRepository) ListByUserID(userID uuid.UUID) ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.Where("user_id = ?", userID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// ListAll retrieves all transactions
func (r *TransactionRepository) ListAll() ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetDB returns the database instance
func (r *TransactionRepository) GetDB() *gorm.DB {
	return r.db
}
