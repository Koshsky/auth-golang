package interceptors

import (
	"google.golang.org/grpc"
)

// InterceptorManager manages all interceptors for the gRPC server
type InterceptorManager struct {
	interceptors []ILoggingInterceptor
}

// NewInterceptorManager creates a new interceptor manager
func NewInterceptorManager() *InterceptorManager {
	return &InterceptorManager{
		interceptors: make([]ILoggingInterceptor, 0),
	}
}

// AddLoggingInterceptor adds a logging interceptor to the manager
func (m *InterceptorManager) AddLoggingInterceptor(interceptor ILoggingInterceptor) {
	m.interceptors = append(m.interceptors, interceptor)
}

// GetUnaryInterceptors returns all unary interceptors in the correct order
func (m *InterceptorManager) GetUnaryInterceptors() []grpc.UnaryServerInterceptor {
	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0, len(m.interceptors))

	for _, interceptor := range m.interceptors {
		unaryInterceptors = append(unaryInterceptors, interceptor.UnaryServerInterceptor())
	}

	return unaryInterceptors
}

// GetStreamInterceptors returns all stream interceptors in the correct order
func (m *InterceptorManager) GetStreamInterceptors() []grpc.StreamServerInterceptor {
	streamInterceptors := make([]grpc.StreamServerInterceptor, 0, len(m.interceptors))

	for _, interceptor := range m.interceptors {
		streamInterceptors = append(streamInterceptors, interceptor.StreamServerInterceptor())
	}

	return streamInterceptors
}
