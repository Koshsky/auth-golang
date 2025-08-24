package repositories

import "github.com/Koshsky/subs-service/auth-service/internal/models"

//go:generate mockery --name=IUserRepository --output=./mocks --outpkg=mocks --filename=IUserRepository.go
type IUserRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	UserExists(email string) (bool, error)
}

//go:generate mockery --name=IDatabase --output=./mocks --outpkg=mocks --filename=IDatabase.go
type IDatabase interface {
	Create(value interface{}) IDatabase
	Where(query interface{}, args ...interface{}) IDatabase
	First(dest interface{}, conds ...interface{}) IDatabase
	Model(value interface{}) IDatabase
	Count(value *int64) IDatabase
	GetError() error
}

// Interface compliance checks - will fail at compile time if interfaces are not implemented
var _ IUserRepository = (*UserRepository)(nil)
var _ IDatabase = (*GormAdapter)(nil)
