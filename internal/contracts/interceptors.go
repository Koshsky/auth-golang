package contracts

import (
	"google.golang.org/grpc"
)

//go:generate mockery --name=ILoggingInterceptor --output=./mocks --outpkg=mocks --filename=ILoggingInterceptor.go
type ILoggingInterceptor interface {
	UnaryServerInterceptor() grpc.UnaryServerInterceptor
	StreamServerInterceptor() grpc.StreamServerInterceptor
}

//go:generate mockery --name=IUnaryInterceptor --output=./mocks --outpkg=mocks --filename=IUnaryInterceptor.go
type IUnaryInterceptor interface {
	UnaryServerInterceptor() grpc.UnaryServerInterceptor
}

//go:generate mockery --name=IStreamInterceptor --output=./mocks --outpkg=mocks --filename=IStreamInterceptor.go
type IStreamInterceptor interface {
	StreamServerInterceptor() grpc.StreamServerInterceptor
}

//go:generate mockery --name=IInterceptorManager --output=./mocks --outpkg=mocks --filename=IInterceptorManager.go
type IInterceptorManager interface {
	AddUnaryInterceptor(interceptor IUnaryInterceptor)
	AddStreamInterceptor(interceptor IStreamInterceptor)
	GetUnaryInterceptors() []grpc.UnaryServerInterceptor
	GetStreamInterceptors() []grpc.StreamServerInterceptor
}
