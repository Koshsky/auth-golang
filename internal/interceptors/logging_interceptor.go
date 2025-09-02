package interceptors

import (
	"context"
	"time"

	"log/slog"

	"github.com/Koshsky/subs-service/auth-service/internal/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type LoggingInterceptor struct {
	logger *slog.Logger
}

func NewLoggingInterceptor(logger *slog.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{
		logger: logger,
	}
}

func (i *LoggingInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		startTime := time.Now()

		logCtx := i.extractLoggingContext(ctx, info)

		i.logger.InfoContext(logCtx, "gRPC request started",
			"method", info.FullMethod,
			"request_type", getRequestType(req),
		)

		resp, err := handler(logCtx, req)

		duration := time.Since(startTime)

		i.logRequestCompletion(logCtx, info.FullMethod, duration, err)

		return resp, err
	}
}

func (i *LoggingInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		startTime := time.Now()

		logCtx := i.extractLoggingContext(stream.Context(), info)

		i.logger.InfoContext(logCtx, "gRPC stream started",
			"method", info.FullMethod,
			"is_client_stream", info.IsClientStream,
			"is_server_stream", info.IsServerStream,
		)

		err := handler(srv, stream)

		duration := time.Since(startTime)

		i.logRequestCompletion(logCtx, info.FullMethod, duration, err)

		return err
	}
}

func (i *LoggingInterceptor) extractLoggingContext(ctx context.Context, info interface{}) context.Context {
	logCtx := logging.WithOperation(ctx, "grpc_request")

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if requestIDs := md.Get("x-request-id"); len(requestIDs) > 0 {
			logCtx = logging.WithRequestID(logCtx, requestIDs[0])
		}

		if traceIDs := md.Get("x-trace-id"); len(traceIDs) > 0 {
			logCtx = logging.WithTraceID(logCtx, traceIDs[0])
		}

		if userIDs := md.Get("x-user-id"); len(userIDs) > 0 {
			logCtx = logging.WithUserID(logCtx, userIDs[0])
		}
	}

	return logCtx
}

func (i *LoggingInterceptor) logRequestCompletion(
	ctx context.Context,
	method string,
	duration time.Duration,
	err error,
) {
	var level slog.Level
	var message string
	var additionalFields []interface{}

	if err != nil {
		if st, ok := status.FromError(err); ok {
			level = getLogLevelForStatusCode(st.Code())
			message = "gRPC request completed with error"
			additionalFields = []interface{}{
				"grpc_code", st.Code().String(),
				"grpc_message", st.Message(),
				"duration_ms", duration.Milliseconds(),
			}
		} else {
			level = slog.LevelError
			message = "gRPC request completed with unexpected error"
			additionalFields = []interface{}{
				"error", err.Error(),
				"duration_ms", duration.Milliseconds(),
			}
		}
	} else {
		level = slog.LevelInfo
		message = "gRPC request completed successfully"
		additionalFields = []interface{}{
			"duration_ms", duration.Milliseconds(),
		}
	}

	switch level {
	case slog.LevelDebug:
		i.logger.DebugContext(ctx, message, additionalFields...)
	case slog.LevelInfo:
		i.logger.InfoContext(ctx, message, additionalFields...)
	case slog.LevelWarn:
		i.logger.WarnContext(ctx, message, additionalFields...)
	case slog.LevelError:
		i.logger.ErrorContext(ctx, message, additionalFields...)
	}
}

func getLogLevelForStatusCode(code codes.Code) slog.Level {
	switch code {
	case codes.OK:
		return slog.LevelInfo
	case codes.Canceled, codes.DeadlineExceeded:
		return slog.LevelWarn
	case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists,
		codes.PermissionDenied, codes.Unauthenticated:
		return slog.LevelInfo
	case codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted,
		codes.OutOfRange, codes.Unimplemented, codes.Internal,
		codes.Unavailable, codes.DataLoss:
		return slog.LevelError
	default:
		return slog.LevelError
	}
}

// getRequestType extracts the type name from the request interface
func getRequestType(req interface{}) string {
	if req == nil {
		return "nil"
	}

	// Use type assertion to get the concrete type
	switch req.(type) {
	case interface{ GetEmail() string }:
		return "auth_request"
	case interface{ GetToken() string }:
		return "token_request"
	default:
		// Fallback to reflection if needed
		return "unknown_request"
	}
}
