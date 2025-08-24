package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Koshsky/subs-service/auth-service/internal/config"
	"github.com/Koshsky/subs-service/auth-service/internal/messaging"
	"github.com/Koshsky/subs-service/auth-service/internal/models"
	"github.com/Koshsky/subs-service/auth-service/internal/repositories"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService implements authentication business logic
type AuthService struct {
	userRepo      repositories.IUserRepository
	messageBroker messaging.IMessageBroker
	JWTSecret     []byte
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userRepo repositories.IUserRepository, messageBroker messaging.IMessageBroker, cfg *config.Config) *AuthService {
	if cfg == nil || cfg.JWTSecret == "" {
		return &AuthService{
			userRepo:      userRepo,
			messageBroker: messageBroker,
			JWTSecret:     nil,
		}
	}
	return &AuthService{
		userRepo:      userRepo,
		messageBroker: messageBroker,
		JWTSecret:     []byte(cfg.JWTSecret),
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, email, password string) (*models.User, error) {
	_ = ctx // TODO: use ctx in future
	if s.userRepo == nil {
		return nil, errors.New("user repository is not initialized")
	}

	// Check if user already exists
	exists, err := s.userRepo.UserExists(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, errors.New("user already exists")
	}

	// Hash password in service layer
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create new user with hashed password
	user := &models.User{
		Email:    email,
		Password: string(hashedPassword),
	}

	err = s.userRepo.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Publish user created event
	if s.messageBroker != nil {
		err = s.messageBroker.PublishUserCreated(user)
		if err != nil {
			// Log error but don't fail registration
			fmt.Printf("Failed to publish user created event: %v\n", err)
		}
	}

	return user, nil
}

// Login authenticates a user and returns JWT token
func (s *AuthService) Login(ctx context.Context, email, password string) (string, *models.User, error) {
	_ = ctx // TODO: use ctx in future
	if s.userRepo == nil {
		return "", nil, errors.New("user repository is not initialized")
	}

	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return "", nil, fmt.Errorf("invalid credentials: %v", err)
	}

	// Compare password with hashed password in service layer
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", nil, fmt.Errorf("invalid credentials: %v", err)
	}

	token, err := s.GenerateJWTToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// ValidateToken validates JWT token and returns claims
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (jwt.MapClaims, error) {
	_ = ctx // TODO: use ctx in future
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.JWTSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateJWTToken generates JWT token for user
func (s *AuthService) GenerateJWTToken(user *models.User) (string, error) {
	if user == nil {
		return "", errors.New("user cannot be nil")
	}
	if s.JWTSecret == nil {
		return "", errors.New("JWT secret is not configured")
	}

	claims := jwt.MapClaims{
		"email":   user.Email,
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.JWTSecret)
}
