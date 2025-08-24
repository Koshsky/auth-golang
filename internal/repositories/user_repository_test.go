package repositories_test

import (
	"errors"
	"testing"

	"github.com/Koshsky/subs-service/auth-service/internal/models"
	"github.com/Koshsky/subs-service/auth-service/internal/repositories"
	"github.com/Koshsky/subs-service/auth-service/internal/repositories/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	mockDB   *mocks.IDatabase
	userRepo *repositories.UserRepository
	testUser *models.User
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	suite.mockDB = new(mocks.IDatabase)
	suite.userRepo = &repositories.UserRepository{DB: suite.mockDB}
	suite.testUser = &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "hashedpassword123",
	}
}

func (suite *UserRepositoryTestSuite) TearDownTest() {
	suite.mockDB.AssertExpectations(suite.T())
}

// ===== MOCK HELPER FUNCTIONS =====

// mockCreateUser mocks DB.Create(user).GetError() with provided error
func (suite *UserRepositoryTestSuite) mockCreateUser(user *models.User, err error) {
	suite.mockDB.On("Create", user).Return(suite.mockDB)
	suite.mockDB.On("GetError").Return(err)
}

// mockWhereEmail mocks DB.Where("email = ?", email)
func (suite *UserRepositoryTestSuite) mockWhereEmail(email string) {
	suite.mockDB.On("Where", "email = ?", email).Return(suite.mockDB)
}

// mockGetUserByEmail mocks DB.First(&user).GetError()
func (suite *UserRepositoryTestSuite) mockGetUserByEmail(email string, u *models.User, err error) {
	suite.mockWhereEmail(email)
	suite.mockDB.On("First", mock.AnythingOfType("*models.User")).Run(func(args mock.Arguments) {
		if u != nil {
			dest := args.Get(0).(*models.User)
			*dest = *u
		}
	}).Return(suite.mockDB)
	suite.mockDB.On("GetError").Return(err)
}

// mockCountByEmail mocks Model(User).Where(email).Count(&count).GetError()
func (suite *UserRepositoryTestSuite) mockCountByEmail(email string, countValue int64, err error) {
	suite.mockDB.On("Model", mock.AnythingOfType("*models.User")).Return(suite.mockDB)
	suite.mockWhereEmail(email)
	suite.mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		cnt := args.Get(0).(*int64)
		*cnt = countValue
	}).Return(suite.mockDB)
	suite.mockDB.On("GetError").Return(err)
}

// ===== CONSTRUCTOR TESTS =====

func (suite *UserRepositoryTestSuite) TestNewUserRepository_Success() {
	// Arrange
	mockDB := new(mocks.IDatabase)

	// Act
	repo := repositories.NewUserRepository(mockDB)

	// Assert
	suite.Require().NotNil(repo)
	suite.Equal(mockDB, repo.DB)
}

// ===== CREATE USER TESTS =====

func (suite *UserRepositoryTestSuite) TestCreateUser_Success() {
	// Arrange
	suite.mockCreateUser(suite.testUser, nil)

	// Act
	err := suite.userRepo.CreateUser(suite.testUser)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotEqual(uuid.Nil, suite.testUser.ID)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestCreateUser_NilDatabase() {
	// Arrange
	repo := &repositories.UserRepository{DB: nil}

	// Act
	err := repo.CreateUser(suite.testUser)

	// Assert
	suite.Require().Error(err)
	suite.Contains(err.Error(), "database connection is not initialized")
}

func (suite *UserRepositoryTestSuite) TestCreateUser_WithExistingUUID() {
	// Arrange
	existingUUID := uuid.New()
	suite.testUser.ID = existingUUID
	suite.mockCreateUser(suite.testUser, nil)

	// Act
	err := suite.userRepo.CreateUser(suite.testUser)

	// Assert
	suite.Require().NoError(err)
	suite.Equal(existingUUID, suite.testUser.ID) // UUID should not change
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestCreateUser_DatabaseError() {
	// Arrange
	expectedError := errors.New("database error")
	suite.mockCreateUser(suite.testUser, expectedError)

	// Act
	err := suite.userRepo.CreateUser(suite.testUser)

	// Assert
	suite.Require().Error(err)
	suite.Contains(err.Error(), "database error")
	suite.mockDB.AssertExpectations(suite.T())
}

// ===== GET USER BY EMAIL TESTS =====

func (suite *UserRepositoryTestSuite) TestGetUserByEmail_Success() {
	// Arrange
	suite.mockGetUserByEmail(suite.testUser.Email, suite.testUser, nil)

	// Act
	user, err := suite.userRepo.GetUserByEmail(suite.testUser.Email)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(user)
	suite.Equal(suite.testUser.Email, user.Email)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestGetUserByEmail_NilDatabase() {
	// Arrange
	repo := &repositories.UserRepository{DB: nil}

	// Act
	user, err := repo.GetUserByEmail(suite.testUser.Email)

	// Assert
	suite.Require().Error(err)
	suite.Require().Nil(user)
	suite.Contains(err.Error(), "database connection is not initialized")
}

func (suite *UserRepositoryTestSuite) TestGetUserByEmail_UserNotFound() {
	// Arrange
	suite.mockGetUserByEmail(suite.testUser.Email, nil, errors.New("record not found"))

	// Act
	user, err := suite.userRepo.GetUserByEmail(suite.testUser.Email)

	// Assert
	suite.Require().Error(err)
	suite.Require().Nil(user)
	suite.Contains(err.Error(), "record not found")
	suite.mockDB.AssertExpectations(suite.T())
}

// ===== USER EXISTS TESTS =====

func (suite *UserRepositoryTestSuite) TestUserExists_Success() {
	// Arrange
	suite.mockCountByEmail(suite.testUser.Email, 1, nil)

	// Act
	exists, err := suite.userRepo.UserExists(suite.testUser.Email)

	// Assert
	suite.Require().NoError(err)
	suite.Require().True(exists)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestUserExists_UserNotFound() {
	// Arrange
	suite.mockCountByEmail(suite.testUser.Email, 0, nil)

	// Act
	exists, err := suite.userRepo.UserExists(suite.testUser.Email)

	// Assert
	suite.Require().NoError(err)
	suite.Require().False(exists)
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestUserExists_DatabaseError() {
	// Arrange
	expectedError := errors.New("database error")
	suite.mockCountByEmail(suite.testUser.Email, 0, expectedError)

	// Act
	exists, err := suite.userRepo.UserExists(suite.testUser.Email)

	// Assert
	suite.Require().Error(err)
	suite.Require().False(exists)
	suite.Contains(err.Error(), "database error")
	suite.mockDB.AssertExpectations(suite.T())
}

func (suite *UserRepositoryTestSuite) TestUserExists_NilDatabase() {
	// Arrange
	repo := &repositories.UserRepository{DB: nil}

	// Act
	exists, err := repo.UserExists(suite.testUser.Email)

	// Assert
	suite.Require().Error(err)
	suite.Require().False(exists)
	suite.Contains(err.Error(), "database connection is not initialized")
}

// Run tests
func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
