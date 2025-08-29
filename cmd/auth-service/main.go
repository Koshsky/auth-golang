package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	"github.com/Koshsky/subs-service/auth-service/internal/authpb"
	"github.com/Koshsky/subs-service/auth-service/internal/config"
	"github.com/Koshsky/subs-service/auth-service/internal/logging"
	"github.com/Koshsky/subs-service/auth-service/internal/messaging"
	"github.com/Koshsky/subs-service/auth-service/internal/repositories"
	"github.com/Koshsky/subs-service/auth-service/internal/server"
	"github.com/Koshsky/subs-service/auth-service/internal/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// setupServices initializes all services and returns them
func setupServices(ctx context.Context, cfg *config.Config) (*services.AuthService, *server.AuthServer, error) {
	ctx = logging.WithOperation(ctx, "setup_services")

	// Initialize RabbitMQ service
	rabbitmqService, err := messaging.NewRabbitMQAdapter(cfg.RabbitMQ)
	if err != nil {
		slog.WarnContext(ctx, "Failed to initialize RabbitMQ service", "error", err)
		slog.InfoContext(ctx, "Auth service will continue without event publishing")
		rabbitmqService = nil
	}

	// Initialize database and repositories
	gormAdapter, err := repositories.NewGormAdapter(&cfg.Database)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to initialize database", "error", err)
		return nil, nil, err
	}

	userRepo := repositories.NewUserRepository(gormAdapter)
	authService := services.NewAuthService(userRepo, rabbitmqService, cfg)
	authServer := server.NewAuthServer(authService)

	slog.InfoContext(ctx, "Services initialized successfully")
	return authService, authServer, nil
}

// createGRPCServer creates and configures the gRPC server
func createGRPCServer(ctx context.Context, cfg *config.Config) (*grpc.Server, error) {
	ctx = logging.WithOperation(ctx, "create_grpc_server")

	var grpcServer *grpc.Server

	if cfg.EnableTLS {
		slog.InfoContext(ctx, "Initializing gRPC server with TLS",
			"cert_file", cfg.TLSCertFile, "key_file", cfg.TLSKeyFile)
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to load TLS credentials", "error", err)
			return nil, err
		}
		grpcServer = grpc.NewServer(grpc.Creds(creds))
		slog.InfoContext(ctx, "gRPC server created with TLS")
	} else {
		slog.InfoContext(ctx, "Initializing gRPC server without TLS")
		grpcServer = grpc.NewServer()
		slog.InfoContext(ctx, "gRPC server created without TLS")
	}

	return grpcServer, nil
}

// startServer starts the gRPC server
func startServer(ctx context.Context, grpcServer *grpc.Server, authServer *server.AuthServer, port string) error {
	ctx = logging.WithOperation(ctx, "start_server")

	authpb.RegisterAuthServiceServer(grpcServer, authServer)
	slog.InfoContext(ctx, "Auth service registered")

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to listen on port", "port", port, "error", err)
		return err
	}

	slog.InfoContext(ctx, "Auth service starting", "port", port)
	return grpcServer.Serve(lis)
}

func main() {
	cfg := config.LoadConfig()

	// Initialize logging from config
	if err := logging.InitLogging(cfg.Logging); err != nil {
		slog.Error("Failed to initialize logging", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	ctx = logging.WithOperation(ctx, "main")

	slog.InfoContext(ctx, "Starting auth service",
		"version", cfg.Logging.Version,
		"environment", cfg.Environment)

	// Setup services
	_, authServer, err := setupServices(ctx, cfg)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to setup services", "error", err)
		os.Exit(1)
	}

	// Create gRPC server
	grpcServer, err := createGRPCServer(ctx, cfg)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create gRPC server", "error", err)
		os.Exit(1)
	}

	// Start server
	if err := startServer(ctx, grpcServer, authServer, cfg.Port); err != nil {
		slog.ErrorContext(ctx, "gRPC server stopped", "error", err)
		os.Exit(1)
	}
}
