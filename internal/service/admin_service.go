package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/takadao/banking/internal/models"
	"github.com/takadao/banking/internal/repository"
)

type AdminService struct {
	userRepo        *repository.UserRepository
	transactionRepo *repository.TransactionRepository
}

func NewAdminService(userRepo *repository.UserRepository, transactionRepo *repository.TransactionRepository) *AdminService {
	return &AdminService{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *AdminService) ListAllTransactions(page, pageSize int) ([]models.Transaction, int64, error) {
	return s.transactionRepo.GetAll(page, pageSize)
}

func (s *AdminService) GetUserBalanceAtTime(userID uuid.UUID, currency string, atTime time.Time) (float64, error) {
	return s.transactionRepo.GetBalanceAtTime(userID, currency, atTime)
} 