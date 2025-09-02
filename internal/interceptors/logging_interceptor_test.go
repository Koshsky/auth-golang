package interceptors

import (
	"context"
	"log/slog"
	"testing"

	"github.com/Koshsky/subs-service/auth-service/internal/authpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNewLoggingInterceptor(t *testing.T) {
	logger := &slog.Logger{}
	interceptor := NewLoggingInterceptor(logger)

	assert.NotNil(t, interceptor)
	assert.Equal(t, logger, interceptor.logger)
}

func TestLoggingInterceptor_UnaryServerInterceptor_Success(t *testing.T) {
	// Setup
	logger := slog.Default()
	interceptor := NewLoggingInterceptor(logger)

	// Create test context with metadata
	ctx := context.Background()
	md := metadata.New(map[string]string{
		"x-request-id": "test-req-123",
		"x-trace-id":   "test-trace-456",
	})
	ctx = metadata.NewIncomingContext(ctx, md)

	// Create test request
	req := &authpb.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Create mock handler
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return &authpb.LoginResponse{Success: true}, nil
	}

	// Execute interceptor
	unaryInterceptor := interceptor.UnaryServerInterceptor()
	resp, err := unaryInterceptor(ctx, req, &grpc.UnaryServerInfo{
		FullMethod: "/authpb.AuthService/Login",
	}, handler)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestLoggingInterceptor_UnaryServerInterceptor_Error(t *testing.T) {
	// Setup
	logger := slog.Default()
	interceptor := NewLoggingInterceptor(logger)

	ctx := context.Background()
	req := &authpb.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Create mock handler that returns error
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.InvalidArgument, "Invalid credentials")
	}

	// Execute interceptor
	unaryInterceptor := interceptor.UnaryServerInterceptor()
	resp, err := unaryInterceptor(ctx, req, &grpc.UnaryServerInfo{
		FullMethod: "/authpb.AuthService/Login",
	}, handler)

	// Verify results
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestLoggingInterceptor_StreamServerInterceptor(t *testing.T) {
	// Setup
	logger := slog.Default()
	interceptor := NewLoggingInterceptor(logger)

	// Create mock stream
	mockStream := &MockServerStream{}
	mockStream.On("Context").Return(context.Background())

	// Create mock handler
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	// Execute interceptor
	streamInterceptor := interceptor.StreamServerInterceptor()
	err := streamInterceptor(nil, mockStream, &grpc.StreamServerInfo{
		FullMethod:     "/authpb.AuthService/StreamMethod",
		IsClientStream: true,
		IsServerStream: false,
	}, handler)

	// Verify results
	assert.NoError(t, err)
	mockStream.AssertExpectations(t)
}

func TestInterceptorManager(t *testing.T) {
	// Setup
	manager := NewInterceptorManager()

	// Create a real logging interceptor for testing
	logger := slog.Default()
	realInterceptor := NewLoggingInterceptor(logger)

	// Add interceptor to manager
	manager.AddLoggingInterceptor(realInterceptor)

	// Test unary interceptors
	unaryInterceptors := manager.GetUnaryInterceptors()
	assert.Len(t, unaryInterceptors, 1)

	// Test stream interceptors
	streamInterceptors := manager.GetStreamInterceptors()
	assert.Len(t, streamInterceptors, 1)

	// Test that interceptors are actually functions
	assert.NotNil(t, unaryInterceptors[0])
	assert.NotNil(t, streamInterceptors[0])
}

func TestGetLogLevelForStatusCode(t *testing.T) {
	// Test various status codes
	testCases := []struct {
		code     codes.Code
		expected slog.Level
	}{
		{codes.OK, slog.LevelInfo},
		{codes.Canceled, slog.LevelWarn},
		{codes.DeadlineExceeded, slog.LevelWarn},
		{codes.InvalidArgument, slog.LevelInfo},
		{codes.NotFound, slog.LevelInfo},
		{codes.PermissionDenied, slog.LevelInfo},
		{codes.Unauthenticated, slog.LevelInfo},
		{codes.ResourceExhausted, slog.LevelError},
		{codes.Internal, slog.LevelError},
		{codes.Unavailable, slog.LevelError},
	}

	for _, tc := range testCases {
		t.Run(tc.code.String(), func(t *testing.T) {
			level := getLogLevelForStatusCode(tc.code)
			assert.Equal(t, tc.expected, level)
		})
	}
}

func TestGetRequestType(t *testing.T) {
	// Test different request types
	loginReq := &authpb.LoginRequest{Email: "test@example.com"}
	tokenReq := &authpb.TokenRequest{Token: "jwt-token"}

	assert.Equal(t, "auth_request", getRequestType(loginReq))
	assert.Equal(t, "token_request", getRequestType(tokenReq))
	assert.Equal(t, "nil", getRequestType(nil))
	assert.Equal(t, "unknown_request", getRequestType("unknown"))
}

// MockServerStream is a mock implementation of grpc.ServerStream for testing
type MockServerStream struct {
	mock.Mock
}

func (m *MockServerStream) SetHeader(metadata.MD) error {
	args := m.Called(metadata.MD{})
	return args.Error(0)
}

func (m *MockServerStream) SendHeader(metadata.MD) error {
	args := m.Called(metadata.MD{})
	return args.Error(0)
}

func (m *MockServerStream) SetTrailer(metadata.MD) {
	m.Called(metadata.MD{})
}

func (m *MockServerStream) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *MockServerStream) SendMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockServerStream) RecvMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}
