package interceptors

import (
	"google.golang.org/grpc"
)

//go:generate mockery --name=ILoggingInterceptor --output=./mocks --outpkg=mocks
//go:generate mockery --name=IInterceptorManager --output=./mocks --outpkg=mocks

// ILoggingInterceptor defines the interface for logging interceptors
type ILoggingInterceptor interface {
	UnaryServerInterceptor() grpc.UnaryServerInterceptor
	StreamServerInterceptor() grpc.StreamServerInterceptor
}

// IInterceptorManager defines the interface for managing interceptors
type IInterceptorManager interface {
	GetUnaryInterceptors() []grpc.UnaryServerInterceptor
	GetStreamInterceptors() []grpc.StreamServerInterceptor
}
