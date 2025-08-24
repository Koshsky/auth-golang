package services

import (
	"context"

	"github.com/Koshsky/subs-service/auth-service/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

//go:generate mockery --name=IAuthService --output=./mocks --outpkg=mocks --filename=IAuthService.go
type IAuthService interface {
	Register(ctx context.Context, email, password string) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, *models.User, error)
	ValidateToken(ctx context.Context, tokenString string) (jwt.MapClaims, error)
	GenerateJWTToken(user *models.User) (string, error)
}

// Interface compliance checks - will fail at compile time if interfaces are not implemented
var _ IAuthService = (*AuthService)(nil)
