package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Koshsky/subs-service/auth-service/internal/authpb"
	"github.com/Koshsky/subs-service/auth-service/internal/models"
	"github.com/Koshsky/subs-service/auth-service/internal/server"
	"github.com/Koshsky/subs-service/auth-service/internal/services/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type AuthServerTestSuite struct {
	suite.Suite
	mockAuthService *mocks.IAuthService
	authServer      *server.AuthServer
	ctx             context.Context
	token           string
	invalidToken    string
	email           string
	password        string
}

func (suite *AuthServerTestSuite) SetupSuite() {
	suite.token = "valid.jwt.token"
	suite.invalidToken = "invalid.jwt.token"
	suite.email = "test@example.com"
	suite.password = "password123"
}

func (suite *AuthServerTestSuite) SetupTest() {
	suite.mockAuthService = new(mocks.IAuthService)
	suite.authServer = server.NewAuthServer(suite.mockAuthService)
	suite.ctx = context.Background()
}

func (suite *AuthServerTestSuite) TearDownTest() {
	suite.mockAuthService.AssertExpectations(suite.T())
}

// ===== VALIDATE TOKEN TESTS =====

func (suite *AuthServerTestSuite) TestValidateToken_Success() {
	// Arrange
	req := &authpb.TokenRequest{Token: suite.token}
	expectedClaims := jwt.MapClaims{
		"user_id": "test-user-id",
		"email":   suite.email,
	}
	suite.mockAuthService.On("ValidateToken", suite.ctx, suite.token).Return(expectedClaims, nil)

	// Act
	response, err := suite.authServer.ValidateToken(suite.ctx, req)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(response)
	suite.True(response.Valid)
	suite.Equal("test-user-id", response.UserId)
	suite.Equal("test@example.com", response.Email)
	suite.Empty(response.Error)
}

func (suite *AuthServerTestSuite) TestValidateToken_InvalidToken() {
	// Arrange
	req := &authpb.TokenRequest{Token: suite.invalidToken}
	expectedError := errors.New("invalid token")
	suite.mockAuthService.On("ValidateToken", suite.ctx, suite.invalidToken).Return(nil, expectedError)

	// Act
	response, err := suite.authServer.ValidateToken(suite.ctx, req)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(response)
	suite.False(response.Valid)
	suite.Empty(response.UserId)
	suite.Empty(response.Email)
	suite.Equal("invalid token", response.Error)
}

func (suite *AuthServerTestSuite) TestValidateToken_InvalidUserID() {
	// Arrange
	req := &authpb.TokenRequest{Token: suite.token}
	expectedClaims := jwt.MapClaims{
		"user_id": 123, // Invalid type
		"email":   suite.email,
	}
	suite.mockAuthService.On("ValidateToken", suite.ctx, suite.token).Return(expectedClaims, nil)

	// Act
	response, err := suite.authServer.ValidateToken(suite.ctx, req)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(response)
	suite.False(response.Valid)
	suite.Empty(response.UserId)
	suite.Empty(response.Email)
	suite.Equal("Invalid user ID in token", response.Error)
}

func (suite *AuthServerTestSuite) TestValidateToken_InvalidEmail() {
	// Arrange
	req := &authpb.TokenRequest{Token: suite.token}
	expectedClaims := jwt.MapClaims{
		"user_id": "test-user-id",
		"email":   123, // Invalid type
	}
	suite.mockAuthService.On("ValidateToken", suite.ctx, suite.token).Return(expectedClaims, nil)

	// Act
	response, err := suite.authServer.ValidateToken(suite.ctx, req)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(response)
	suite.False(response.Valid)
	suite.Empty(response.UserId)
	suite.Empty(response.Email)
	suite.Equal("Invalid email in token", response.Error)
}

// ===== REGISTER TESTS =====

func (suite *AuthServerTestSuite) TestRegister_Success() {
	// Arrange
	req := &authpb.RegisterRequest{
		Email:    suite.email,
		Password: suite.password,
	}
	expectedUser := &models.User{
		ID:    uuid.New(),
		Email: suite.email,
	}

	suite.mockAuthService.On("Register", suite.ctx, suite.email, suite.password).Return(expectedUser, nil)

	// Act
	response, err := suite.authServer.Register(suite.ctx, req)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(response)
	suite.True(response.Success)
	suite.Equal(expectedUser.ID.String(), response.UserId)
	suite.Equal(suite.email, response.Email)
	suite.Equal("User created successfully", response.Message)
	suite.Empty(response.Error)
}

func (suite *AuthServerTestSuite) TestRegister_Error() {
	// Arrange
	req := &authpb.RegisterRequest{
		Email:    suite.email,
		Password: suite.password,
	}
	expectedError := errors.New("user already exists")
	suite.mockAuthService.On("Register", suite.ctx, suite.email, suite.password).Return(nil, expectedError)

	// Act
	response, err := suite.authServer.Register(suite.ctx, req)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(response)
	suite.False(response.Success)
	suite.Empty(response.UserId)
	suite.Empty(response.Email)
	suite.Empty(response.Message)
	suite.Equal("user already exists", response.Error)
}

// ===== LOGIN TESTS =====

func (suite *AuthServerTestSuite) TestLogin_Success() {
	// Arrange
	req := &authpb.LoginRequest{
		Email:    suite.email,
		Password: suite.password,
	}
	expectedUser := &models.User{
		ID:    uuid.New(),
		Email: suite.email,
	}
	expectedToken := "jwt.token.here"

	suite.mockAuthService.On("Login", suite.ctx, suite.email, suite.password).Return(expectedToken, expectedUser, nil)

	// Act
	response, err := suite.authServer.Login(suite.ctx, req)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(response)
	suite.True(response.Success)
	suite.Equal(expectedToken, response.Token)
	suite.Equal(expectedUser.ID.String(), response.UserId)
	suite.Equal(suite.email, response.Email)
	suite.Equal("Successful login", response.Message)
	suite.Empty(response.Error)
}

func (suite *AuthServerTestSuite) TestLogin_Error() {
	// Arrange
	req := &authpb.LoginRequest{
		Email:    suite.email,
		Password: "wrongpassword",
	}
	expectedError := errors.New("invalid credentials")
	suite.mockAuthService.On("Login", suite.ctx, suite.email, "wrongpassword").Return("", nil, expectedError)

	// Act
	response, err := suite.authServer.Login(suite.ctx, req)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(response)
	suite.False(response.Success)
	suite.Empty(response.Token)
	suite.Empty(response.UserId)
	suite.Empty(response.Email)
	suite.Empty(response.Message)
	suite.Equal("invalid credentials", response.Error)
}

// Run tests
func TestAuthServerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServerTestSuite))
}
