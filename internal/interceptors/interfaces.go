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

// IUnaryInterceptor defines the interface for unary interceptors
type IUnaryInterceptor interface {
	UnaryServerInterceptor() grpc.UnaryServerInterceptor
}

// IStreamInterceptor defines the interface for stream interceptors
type IStreamInterceptor interface {
	StreamServerInterceptor() grpc.StreamServerInterceptor
}

// IInterceptorManager defines the interface for managing interceptors
type IInterceptorManager interface {
	AddUnaryInterceptor(interceptor IUnaryInterceptor)
	AddStreamInterceptor(interceptor IStreamInterceptor)
	GetUnaryInterceptors() []grpc.UnaryServerInterceptor
	GetStreamInterceptors() []grpc.StreamServerInterceptor
}
