package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/takadao/banking/internal/models"
	"github.com/takadao/banking/internal/repository"
)

type TransactionService struct {
	repo *repository.TransactionRepository
}

func NewTransactionService(repo *repository.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

// Create creates a new transaction
func (s *TransactionService) Create(transaction *models.Transaction) error {
	if err := transaction.Validate(); err != nil {
		return err
	}
	return s.repo.Create(transaction)
}

// ListByUserID retrieves all transactions for a specific user
func (s *TransactionService) ListByUserID(userID uuid.UUID) ([]models.Transaction, error) {
	return s.repo.ListByUserID(userID)
}

// GetByIDAndUserID retrieves a specific transaction for a user
func (s *TransactionService) GetByIDAndUserID(transactionID, userID uuid.UUID) (*models.Transaction, error) {
	return s.repo.GetByIDAndUserID(transactionID, userID)
}

// ListAll retrieves all transactions
func (s *TransactionService) ListAll() ([]models.Transaction, error) {
	return s.repo.ListAll()
}

// GetByID retrieves a transaction by ID
func (s *TransactionService) GetByID(id uuid.UUID) (*models.Transaction, error) {
	return s.repo.GetByID(id)
}

// Deposit creates a deposit transaction
func (s *TransactionService) Deposit(userID uuid.UUID, amount float64, currency, description string) error {
	transaction := &models.Transaction{
		UserID:      userID,
		Type:        models.TransactionTypeDeposit,
		Amount:      amount,
		Currency:    currency,
		Description: description,
	}
	return s.Create(transaction)
}

// Withdraw creates a withdrawal transaction
func (s *TransactionService) Withdraw(userID uuid.UUID, amount float64, currency, description string) error {
	transaction := &models.Transaction{
		UserID:      userID,
		Type:        models.TransactionTypeWithdraw,
		Amount:      amount,
		Currency:    currency,
		Description: description,
	}
	return s.Create(transaction)
}

// Transfer creates a transfer transaction
func (s *TransactionService) Transfer(fromUserID, toUserID uuid.UUID, amount float64, currency, description string) error {
	if fromUserID == toUserID {
		return errors.New("cannot transfer to the same account")
	}

	transaction := &models.Transaction{
		UserID:      fromUserID,
		Type:        models.TransactionTypeTransfer,
		Amount:      amount,
		Currency:    currency,
		RecipientID: &toUserID,
		Description: description,
	}
	return s.Create(transaction)
}

func (s *TransactionService) GetByUserID(userID uuid.UUID, page, pageSize int) ([]models.Transaction, int64, error) {
	return s.repo.GetByUserID(userID, page, pageSize)
}

func (s *TransactionService) GetAll(page, pageSize int) ([]models.Transaction, int64, error) {
	return s.repo.GetAll(page, pageSize)
}

func (s *TransactionService) GetBalanceAtTime(userID uuid.UUID, currency string, atTime time.Time) (float64, error) {
	return s.repo.GetBalanceAtTime(userID, currency, atTime)
}
