package interceptors

import (
	"google.golang.org/grpc"
)

// InterceptorManager manages all interceptors for the gRPC server
type InterceptorManager struct {
	unaryInterceptors  []IUnaryInterceptor
	streamInterceptors []IStreamInterceptor
}

// NewInterceptorManager creates a new interceptor manager
func NewInterceptorManager() *InterceptorManager {
	return &InterceptorManager{
		unaryInterceptors:  make([]IUnaryInterceptor, 0),
		streamInterceptors: make([]IStreamInterceptor, 0),
	}
}

// AddUnaryInterceptor adds a unary interceptor to the manager
func (m *InterceptorManager) AddUnaryInterceptor(interceptor IUnaryInterceptor) {
	m.unaryInterceptors = append(m.unaryInterceptors, interceptor)
}

// AddStreamInterceptor adds a stream interceptor to the manager
func (m *InterceptorManager) AddStreamInterceptor(interceptor IStreamInterceptor) {
	m.streamInterceptors = append(m.streamInterceptors, interceptor)
}

// GetUnaryInterceptors returns all unary interceptors in the correct order
func (m *InterceptorManager) GetUnaryInterceptors() []grpc.UnaryServerInterceptor {
	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0, len(m.unaryInterceptors))

	for _, interceptor := range m.unaryInterceptors {
		unaryInterceptors = append(unaryInterceptors, interceptor.UnaryServerInterceptor())
	}

	return unaryInterceptors
}

// GetStreamInterceptors returns all stream interceptors in the correct order
func (m *InterceptorManager) GetStreamInterceptors() []grpc.StreamServerInterceptor {
	streamInterceptors := make([]grpc.StreamServerInterceptor, 0, len(m.streamInterceptors))

	for _, interceptor := range m.streamInterceptors {
		streamInterceptors = append(streamInterceptors, interceptor.StreamServerInterceptor())
	}

	return streamInterceptors
}
