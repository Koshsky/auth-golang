package contracts

import (
	"context"

	"github.com/Koshsky/subs-service/auth-service/internal/authpb"
)

//go:generate mockery --name=IAuthServer --output=./mocks --outpkg=mocks --filename=IAuthServer.go
type IAuthServer interface {
	ValidateToken(ctx context.Context, req *authpb.TokenRequest) (*authpb.UserResponse, error)
	Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error)
	Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error)
}
