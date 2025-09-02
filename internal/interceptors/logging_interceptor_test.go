package interceptors_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/Koshsky/subs-service/auth-service/api/authpb"
	"github.com/Koshsky/subs-service/auth-service/internal/interceptors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type InterceptorsTestSuite struct {
	suite.Suite
	logger      *slog.Logger
	interceptor *interceptors.LoggingInterceptor
	manager     *interceptors.InterceptorManager
	ctx         context.Context
}

func (s *InterceptorsTestSuite) SetupTest() {
	s.logger = slog.Default()
	s.interceptor = interceptors.NewLoggingInterceptor(s.logger)
	s.manager = interceptors.NewInterceptorManager()
	s.ctx = context.Background()
}

func (s *InterceptorsTestSuite) createTestContext() context.Context {
	md := metadata.New(map[string]string{
		"x-request-id": "test-req-123",
		"x-trace-id":   "test-trace-456",
	})
	return metadata.NewIncomingContext(s.ctx, md)
}

func (s *InterceptorsTestSuite) createTestRequest() *authpb.LoginRequest {
	return &authpb.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
}

func (s *InterceptorsTestSuite) createSuccessHandler() func(context.Context, interface{}) (interface{}, error) {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return &authpb.LoginResponse{Success: true}, nil
	}
}

func (s *InterceptorsTestSuite) createErrorHandler() func(context.Context, interface{}) (interface{}, error) {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.InvalidArgument, "Invalid credentials")
	}
}

func (s *InterceptorsTestSuite) TestNewLoggingInterceptor() {
	interceptor := interceptors.NewLoggingInterceptor(s.logger)

	s.NotNil(interceptor)
	s.Equal(s.logger, interceptor.Logger)
}

func (s *InterceptorsTestSuite) TestLoggingInterceptor_UnaryServerInterceptor_Success() {
	ctx := s.createTestContext()
	req := s.createTestRequest()
	handler := s.createSuccessHandler()

	unaryInterceptor := s.interceptor.UnaryServerInterceptor()
	resp, err := unaryInterceptor(ctx, req, &grpc.UnaryServerInfo{
		FullMethod: "/authpb.AuthService/Login",
	}, handler)

	s.NoError(err)
	s.NotNil(resp)
}

func (s *InterceptorsTestSuite) TestLoggingInterceptor_UnaryServerInterceptor_Error() {
	ctx := s.createTestContext()
	req := s.createTestRequest()
	handler := s.createErrorHandler()

	unaryInterceptor := s.interceptor.UnaryServerInterceptor()
	resp, err := unaryInterceptor(ctx, req, &grpc.UnaryServerInfo{
		FullMethod: "/authpb.AuthService/Login",
	}, handler)

	s.Error(err)
	s.Nil(resp)
	s.Equal(codes.InvalidArgument, status.Code(err))
}

func (s *InterceptorsTestSuite) TestLoggingInterceptor_StreamServerInterceptor() {
	mockStream := &MockServerStream{}
	mockStream.On("Context").Return(context.Background())

	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	streamInterceptor := s.interceptor.StreamServerInterceptor()
	err := streamInterceptor(nil, mockStream, &grpc.StreamServerInfo{
		FullMethod:     "/authpb.AuthService/StreamMethod",
		IsClientStream: true,
		IsServerStream: false,
	}, handler)

	s.NoError(err)
	mockStream.AssertExpectations(s.T())
}

func (s *InterceptorsTestSuite) TestInterceptorManager() {
	s.manager.AddUnaryInterceptor(s.interceptor)
	s.manager.AddStreamInterceptor(s.interceptor)

	unaryInterceptors := s.manager.GetUnaryInterceptors()
	streamInterceptors := s.manager.GetStreamInterceptors()

	s.Len(unaryInterceptors, 1)
	s.Len(streamInterceptors, 1)
	s.NotNil(unaryInterceptors[0])
	s.NotNil(streamInterceptors[0])
}

func (s *InterceptorsTestSuite) TestInterceptorManagerWithDifferentTypes() {
	s.manager.AddUnaryInterceptor(s.interceptor)
	s.manager.AddStreamInterceptor(s.interceptor)

	unaryInterceptors := s.manager.GetUnaryInterceptors()
	streamInterceptors := s.manager.GetStreamInterceptors()

	s.Len(unaryInterceptors, 1)
	s.Len(streamInterceptors, 1)

	for _, interceptor := range unaryInterceptors {
		s.NotNil(interceptor)
	}
	for _, interceptor := range streamInterceptors {
		s.NotNil(interceptor)
	}
}

func (s *InterceptorsTestSuite) TestGetLogLevelForStatusCode() {
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
		s.Run(tc.code.String(), func() {
			level := interceptors.GetLogLevelForStatusCode(tc.code)
			s.Equal(tc.expected, level)
		})
	}
}

func (s *InterceptorsTestSuite) TestGetRequestType() {
	loginReq := &authpb.LoginRequest{Email: "test@example.com"}
	tokenReq := &authpb.TokenRequest{Token: "jwt-token"}
	registerReq := &authpb.RegisterRequest{Email: "test@example.com"}

	s.Equal("login_request", interceptors.GetRequestType(loginReq))
	s.Equal("token_request", interceptors.GetRequestType(tokenReq))
	s.Equal("register_request", interceptors.GetRequestType(registerReq))
	s.Equal("nil", interceptors.GetRequestType(nil))
	s.Equal("string_request", interceptors.GetRequestType("unknown"))
}

func (s *InterceptorsTestSuite) TestToSnakeCase() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"LoginRequest", "login_request"},
		{"TokenRequest", "token_request"},
		{"RegisterRequest", "register_request"},
		{"UserProfileRequest", "user_profile_request"},
		{"APIRequest", "a_p_i_request"},
		{"", ""},
		{"A", "a"},
		{"AB", "a_b"},
	}

	for _, tc := range testCases {
		s.Run(tc.input, func() {
			result := interceptors.ToSnakeCase(tc.input)
			s.Equal(tc.expected, result)
		})
	}
}

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

func TestInterceptorsSuite(t *testing.T) {
	suite.Run(t, new(InterceptorsTestSuite))
}
