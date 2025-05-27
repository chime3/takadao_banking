package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/takadao/banking/internal/models"
	"github.com/takadao/banking/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register creates a new user
func (s *UserService) Register(email, password, role string) (*models.User, error) {
	if role != "user" && role != "admin" {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	user := &models.User{
		Email:    email,
		Password: password,
		Role:     role,
	}
	if err := user.HashPassword(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// Authenticate verifies user credentials
func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if err := user.CheckPassword(password); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}

// GetByID retrieves a user by their ID
func (s *UserService) GetByID(id uuid.UUID) (*models.User, error) {
	return s.repo.GetByID(id)
}

// Update updates a user's information
func (s *UserService) Update(user *models.User) (*models.User, error) {
	// Get existing user to preserve role if not being updated
	existingUser, err := s.repo.GetByID(user.ID)
	if err != nil {
		return nil, err
	}

	// If role is empty, keep the existing role
	if user.Role == "" {
		user.Role = existingUser.Role
	}

	// If password is being updated, hash it
	if user.Password != "" {
		if err := user.HashPassword(); err != nil {
			return nil, err
		}
	} else {
		// Keep existing password if not being updated
		user.Password = existingUser.Password
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

// ListAll retrieves all users
func (s *UserService) ListAll() ([]models.User, error) {
	return s.repo.ListAll()
}

// Delete removes a user
func (s *UserService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}
