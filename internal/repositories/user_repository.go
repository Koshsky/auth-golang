package repositories

import (
	"errors"
	"fmt"

	contracts_repository "github.com/Koshsky/subs-service/auth-service/internal/contracts/repository"
	"github.com/Koshsky/subs-service/auth-service/internal/models"
	"github.com/google/uuid"
)

type UserRepository struct {
	DB contracts_repository.IDatabase
}

func NewUserRepository(db contracts_repository.IDatabase) *UserRepository {
	return &UserRepository{DB: db}
}

func (ur *UserRepository) CreateUser(user *models.User) error {
	if ur.DB == nil {
		return errors.New("database connection is not initialized")
	}

	// Generate UUID if not set
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	dbErr := ur.DB.Create(user).GetError()
	if dbErr != nil {
		return fmt.Errorf("cannot create user with email=%s: %w", user.Email, dbErr)
	}

	return nil
}

func (ur *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	if ur.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	var user models.User
	err := ur.DB.Where("email = ?", email).First(&user).GetError()
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) UserExists(email string) (bool, error) {
	if ur.DB == nil {
		return false, errors.New("database connection is not initialized")
	}

	var count int64
	err := ur.DB.Model(&models.User{}).Where("email = ?", email).Count(&count).GetError()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
