package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Koshsky/subs-service/auth-service/internal/config"
	messagingMocks "github.com/Koshsky/subs-service/auth-service/internal/messaging/mocks"
	"github.com/Koshsky/subs-service/auth-service/internal/models"
	repositoryMocks "github.com/Koshsky/subs-service/auth-service/internal/repositories/mocks"
	"github.com/Koshsky/subs-service/auth-service/internal/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceTestSuite struct {
	suite.Suite
	mockUserRepo      *repositoryMocks.IUserRepository
	mockMessageBroker *messagingMocks.IMessageBroker
	authService       *services.AuthService
	ctx               context.Context
	config            *config.Config
	email             string
	password          string
	wrongPassword     string
	hashedPassword    []byte
	wrongSecret       []byte
	testUser          *models.User // пользователь для тестов с хешированным паролем
}

func (suite *AuthServiceTestSuite) SetupSuite() {
	suite.config = &config.Config{
		JWTSecret: "test-secret",
	}
	suite.email = "test@example.com"
	suite.password = "password123"
	suite.wrongPassword = "wrongpassword"
	suite.wrongSecret = []byte("wrong-secret-key")
	suite.hashedPassword, _ = bcrypt.GenerateFromPassword([]byte(suite.password), bcrypt.DefaultCost)
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.mockUserRepo = repositoryMocks.NewIUserRepository(suite.T())
	suite.mockMessageBroker = messagingMocks.NewIMessageBroker(suite.T())

	suite.authService = services.NewAuthService(suite.mockUserRepo, suite.mockMessageBroker, suite.config)
	suite.ctx = context.Background()

	// testUser с хешированным паролем (как в БД)
	suite.testUser = &models.User{
		ID:       uuid.New(),
		Email:    suite.email,
		Password: string(suite.hashedPassword),
	}
}

// ===== HELPER FUNCTIONS =====

// mockUserExists mock userRepo.UserExists(email)
func (suite *AuthServiceTestSuite) mockUserExists(email string, exists bool, err error) {
	suite.mockUserRepo.On("UserExists", email).Return(exists, err)
}

// mockCreateUser mock userRepo.CreateUser(&user)
func (suite *AuthServiceTestSuite) mockCreateUser(err error) {
	suite.mockUserRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Run(func(args mock.Arguments) {
		user := args.Get(0).(*models.User)
		if user.ID == uuid.Nil {
			user.ID = uuid.New()
		}
	}).Return(err)
}

// mockGetUserByEmail mock userRepo.GetUserByEmail(email)
func (suite *AuthServiceTestSuite) mockGetUserByEmail(email string, user *models.User, err error) {
	suite.mockUserRepo.On("GetUserByEmail", email).Return(user, err)
}

// mockPublishUserCreated mock messageBroker.PublishUserCreated(&user)
func (suite *AuthServiceTestSuite) mockPublishUserCreated(err error) {
	suite.mockMessageBroker.On("PublishUserCreated", mock.AnythingOfType("*models.User")).Return(err)
}

// ===== REGISTER TESTS =====

func (suite *AuthServiceTestSuite) TestRegister_Success() {
	// Arrange
	suite.mockUserExists(suite.email, false, nil)
	suite.mockCreateUser(nil)
	suite.mockPublishUserCreated(nil)

	// Act
	returnedUser, err := suite.authService.Register(suite.ctx, suite.email, suite.password)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(returnedUser)
	suite.Equal(suite.email, returnedUser.Email)
	suite.NotEqual(uuid.Nil, returnedUser.ID)
	// Verify password is hashed
	suite.NotEqual(suite.password, returnedUser.Password)
	suite.Require().NoError(bcrypt.CompareHashAndPassword([]byte(returnedUser.Password), []byte(suite.password)))
}

func (suite *AuthServiceTestSuite) TestRegister_NilUserRepository() {
	// Arrange
	suite.authService = services.NewAuthService(nil, suite.mockMessageBroker, suite.config)

	// Act
	user, err := suite.authService.Register(suite.ctx, suite.email, suite.password)

	// Assert
	suite.Require().Error(err)
	suite.Require().Nil(user)
	suite.Contains(err.Error(), "user repository is not initialized")
}

func (suite *AuthServiceTestSuite) TestRegister_UserAlreadyExists() {
	// Arrange
	suite.mockUserExists(suite.email, true, nil)

	// Act
	user, err := suite.authService.Register(suite.ctx, suite.email, suite.password)

	// Assert
	suite.Require().Error(err)
	suite.Require().Nil(user)
	suite.Contains(err.Error(), "user already exists")
}

func (suite *AuthServiceTestSuite) TestRegister_UserExistsError() {
	// Arrange
	suite.mockUserExists(suite.email, false, errors.New("database error"))

	// Act
	user, err := suite.authService.Register(suite.ctx, suite.email, suite.password)

	// Assert
	suite.Require().Error(err)
	suite.Require().Nil(user)
	suite.Contains(err.Error(), "failed to check user existence")
}

func (suite *AuthServiceTestSuite) TestRegister_CreateUserError() {
	// Arrange
	expectedError := errors.New("database error")

	suite.mockUserExists(suite.email, false, nil)
	suite.mockCreateUser(expectedError)

	// Act
	user, err := suite.authService.Register(suite.ctx, suite.email, suite.password)

	// Assert
	suite.Require().Error(err)
	suite.Require().Nil(user)
	suite.Contains(err.Error(), "database error")
}

func (suite *AuthServiceTestSuite) TestRegister_PublishError() {
	// Arrange
	expectedError := errors.New("publish error")

	suite.mockUserExists(suite.email, false, nil)
	suite.mockCreateUser(nil)
	suite.mockPublishUserCreated(expectedError)

	// Act
	user, err := suite.authService.Register(suite.ctx, suite.email, suite.password)

	// Assert
	suite.Require().NoError(err) // Register should succeed even if RabbitMQ fails
	suite.Require().NotNil(user)
	suite.Equal(suite.email, user.Email)
}

func (suite *AuthServiceTestSuite) TestRegister_PasswordHashingError() {
	// Arrange
	password := strings.Repeat("a", 100) // This should cause bcrypt to fail
	suite.mockUserExists(suite.email, false, nil)

	// Act
	user, err := suite.authService.Register(suite.ctx, suite.email, password)

	// Assert
	suite.Require().Error(err)
	suite.Require().Nil(user)
	suite.Contains(err.Error(), "failed to hash password")
}

// ===== LOGIN TESTS =====

func (suite *AuthServiceTestSuite) TestLogin_Success() {
	// Arrange
	suite.mockGetUserByEmail(suite.email, suite.testUser, nil)

	// Act
	token, returnedUser, err := suite.authService.Login(suite.ctx, suite.email, suite.password)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotEmpty(token)
	suite.Require().NotNil(returnedUser)

	// Validate JWT token structure
	claims, err := suite.authService.ValidateToken(suite.ctx, token)
	suite.Require().NoError(err)
	suite.Require().NotNil(claims)
	suite.Equal(returnedUser.ID.String(), claims["user_id"])
	suite.Equal(returnedUser.Email, claims["email"])
}

func (suite *AuthServiceTestSuite) TestLogin_NilUserRepository() {
	// Arrange
	suite.authService = services.NewAuthService(nil, suite.mockMessageBroker, suite.config)

	// Act
	token, user, err := suite.authService.Login(suite.ctx, suite.email, suite.password)

	// Assert
	suite.Require().Error(err)
	suite.Require().Empty(token)
	suite.Require().Nil(user)
	suite.Contains(err.Error(), "user repository is not initialized")
}

func (suite *AuthServiceTestSuite) TestLogin_ValidationError() {
	// Arrange
	expectedError := errors.New("invalid credentials")
	suite.mockGetUserByEmail(suite.email, nil, expectedError)

	// Act
	token, user, err := suite.authService.Login(suite.ctx, suite.email, suite.password)

	// Assert
	suite.Require().Error(err)
	suite.Require().Empty(token)
	suite.Require().Nil(user)
	suite.Contains(err.Error(), "invalid credentials")
}

func (suite *AuthServiceTestSuite) TestLogin_InvalidPassword() {
	// Arrange
	suite.mockGetUserByEmail(suite.email, suite.testUser, nil)

	// Act
	token, returnedUser, err := suite.authService.Login(suite.ctx, suite.email, suite.wrongPassword)

	// Assert
	suite.Require().Error(err)
	suite.Require().Empty(token)
	suite.Require().Nil(returnedUser)
	suite.Contains(err.Error(), "invalid credentials")
}

func (suite *AuthServiceTestSuite) TestLogin_TokenGenerationError() {
	// Arrange
	suite.mockGetUserByEmail(suite.email, suite.testUser, nil)

	// Create AuthService with empty JWTSecret to cause token generation error
	cfg := &config.Config{JWTSecret: ""}
	authService := services.NewAuthService(suite.mockUserRepo, suite.mockMessageBroker, cfg)

	// Act
	token, returnedUser, err := authService.Login(suite.ctx, suite.email, suite.password)

	// Assert
	suite.Require().Error(err)
	suite.Require().Empty(token)
	suite.Require().Nil(returnedUser)
	suite.Contains(err.Error(), "JWT secret is not configured")
}

// ===== JWT TOKEN TESTS =====

func (suite *AuthServiceTestSuite) TestGenerateJWTToken_Success() {
	// Arrange

	// Act
	token, err := suite.authService.GenerateJWTToken(suite.testUser)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotEmpty(token)

	// Validate JWT token structure
	claims, err := suite.authService.ValidateToken(suite.ctx, token)
	suite.Require().NoError(err)
	suite.Require().NotNil(claims)
	suite.Equal(suite.testUser.ID.String(), claims["user_id"])
	suite.Equal(suite.testUser.Email, claims["email"])
}

func (suite *AuthServiceTestSuite) TestGenerateJWTToken_NilUser() {
	// Act
	token, err := suite.authService.GenerateJWTToken(nil)

	// Assert
	suite.Require().Error(err)
	suite.Require().Empty(token)
	suite.Contains(err.Error(), "user cannot be nil")
}

func (suite *AuthServiceTestSuite) TestGenerateJWTToken_NilSecret() {
	// Arrange
	// Manually set JWTSecret to nil after creation for test
	suite.authService.JWTSecret = nil

	// Act
	token, err := suite.authService.GenerateJWTToken(suite.testUser)

	// Assert
	suite.Require().Error(err)
	suite.Require().Empty(token)
	suite.Contains(err.Error(), "JWT secret is not configured")
}

// ===== VALIDATE TOKEN TESTS =====

func (suite *AuthServiceTestSuite) TestValidateToken_Success() {
	// Arrange
	token, _ := suite.authService.GenerateJWTToken(suite.testUser)

	// Act
	claims, err := suite.authService.ValidateToken(suite.ctx, token)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(claims)
	suite.Equal(suite.testUser.ID.String(), claims["user_id"])
	suite.Equal(suite.testUser.Email, claims["email"])
}

func (suite *AuthServiceTestSuite) TestValidateToken_InvalidClaims() {
	// Arrange
	token, _ := suite.authService.GenerateJWTToken(suite.testUser)

	parts := strings.Split(token, ".")
	if len(parts) >= 2 {
		parts[1] = "invalid-payload"
	}
	corruptedToken := strings.Join(parts, ".")

	// Act
	claims, err := suite.authService.ValidateToken(suite.ctx, corruptedToken)

	// Assert
	suite.Require().Error(err)
	suite.Require().Nil(claims)
	suite.Require().NotEmpty(err.Error())
}

func (suite *AuthServiceTestSuite) TestValidateToken_ForgedToken() {
	// Arrange - create a forged token with correct claims but wrong signing key
	// This simulates a token created by a client without knowing the server's secret
	user := suite.testUser

	// Create claims that look legitimate
	claims := jwt.MapClaims{
		"email":   user.Email,
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	// Create token with correct signing method (HS256) but wrong secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign with a different secret than what the server uses
	wrongSecret := []byte("wrong-secret-key")
	forgedToken, _ := token.SignedString(wrongSecret)

	// Act
	claims, err := suite.authService.ValidateToken(suite.ctx, forgedToken)

	// Assert
	suite.Require().Error(err)
	suite.Require().Nil(claims)
	suite.Contains(err.Error(), "signature is invalid")
}

func (suite *AuthServiceTestSuite) TestValidateToken_ExpiredToken() {
	// Arrange
	claims := jwt.MapClaims{
		"email":   suite.testUser.Email,
		"user_id": suite.testUser.ID.String(),
		"exp":     time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, _ := token.SignedString(suite.authService.JWTSecret)

	// Act
	claims, err := suite.authService.ValidateToken(suite.ctx, expiredToken)

	// Assert
	suite.Require().Error(err)
	suite.Require().Nil(claims)
	suite.Contains(err.Error(), "token is expired")
}

// Run tests
func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
