package interceptors

import (
	"github.com/Koshsky/subs-service/auth-service/internal/contracts"
	"google.golang.org/grpc"
)

// InterceptorManager manages all interceptors for the gRPC server
type InterceptorManager struct {
	UnaryInterceptors  []contracts.IUnaryInterceptor
	StreamInterceptors []contracts.IStreamInterceptor
}

// NewInterceptorManager creates a new interceptor manager
func NewInterceptorManager() *InterceptorManager {
	return &InterceptorManager{
		UnaryInterceptors:  make([]contracts.IUnaryInterceptor, 0),
		StreamInterceptors: make([]contracts.IStreamInterceptor, 0),
	}
}

// AddUnaryInterceptor adds a unary interceptor to the manager
func (m *InterceptorManager) AddUnaryInterceptor(interceptor contracts.IUnaryInterceptor) {
	m.UnaryInterceptors = append(m.UnaryInterceptors, interceptor)
}

// AddStreamInterceptor adds a stream interceptor to the manager
func (m *InterceptorManager) AddStreamInterceptor(interceptor contracts.IStreamInterceptor) {
	m.StreamInterceptors = append(m.StreamInterceptors, interceptor)
}

// GetUnaryInterceptors returns all unary interceptors in the correct order
func (m *InterceptorManager) GetUnaryInterceptors() []grpc.UnaryServerInterceptor {
	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0, len(m.UnaryInterceptors))

	for _, interceptor := range m.UnaryInterceptors {
		unaryInterceptors = append(unaryInterceptors, interceptor.UnaryServerInterceptor())
	}

	return unaryInterceptors
}

// GetStreamInterceptors returns all stream interceptors in the correct order
func (m *InterceptorManager) GetStreamInterceptors() []grpc.StreamServerInterceptor {
	streamInterceptors := make([]grpc.StreamServerInterceptor, 0, len(m.StreamInterceptors))

	for _, interceptor := range m.StreamInterceptors {
		streamInterceptors = append(streamInterceptors, interceptor.StreamServerInterceptor())
	}

	return streamInterceptors
}
