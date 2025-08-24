package main

import (
	"log"
	"net"

	"github.com/Koshsky/subs-service/auth-service/internal/authpb"
	"github.com/Koshsky/subs-service/auth-service/internal/config"
	"github.com/Koshsky/subs-service/auth-service/internal/messaging"
	"github.com/Koshsky/subs-service/auth-service/internal/repositories"
	"github.com/Koshsky/subs-service/auth-service/internal/server"
	"github.com/Koshsky/subs-service/auth-service/internal/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// setupServices initializes all services and returns them
func setupServices(cfg *config.Config) (*services.AuthService, *server.AuthServer, error) {
	// Initialize RabbitMQ service
	rabbitmqService, err := messaging.NewRabbitMQAdapter(cfg.RabbitMQ)
	if err != nil {
		log.Printf("Warning: Failed to initialize RabbitMQ service: %v", err)
		log.Printf("Auth service will continue without event publishing")
		rabbitmqService = nil
	}

	// Initialize database and repositories
	gormAdapter, err := repositories.NewGormAdapter(cfg.Database)
	if err != nil {
		return nil, nil, err
	}
	userRepo := repositories.NewUserRepository(gormAdapter)
	authService := services.NewAuthService(userRepo, rabbitmqService, cfg)
	authServer := server.NewAuthServer(authService)

	return authService, authServer, nil
}

// createGRPCServer creates and configures the gRPC server
func createGRPCServer(cfg *config.Config) (*grpc.Server, error) {
	var grpcServer *grpc.Server

	if cfg.EnableTLS {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return nil, err
		}
		grpcServer = grpc.NewServer(grpc.Creds(creds))
	} else {
		grpcServer = grpc.NewServer()
	}

	return grpcServer, nil
}

// startServer starts the gRPC server
func startServer(grpcServer *grpc.Server, authServer *server.AuthServer, port string) error {
	authpb.RegisterAuthServiceServer(grpcServer, authServer)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	log.Printf("Auth service starting on port %s", port)
	return grpcServer.Serve(lis)
}

func main() {
	cfg := config.LoadConfig()

	// Setup services
	_, authServer, err := setupServices(cfg)
	if err != nil {
		log.Fatalf("Failed to setup services: %v", err)
	}

	// Create gRPC server
	grpcServer, err := createGRPCServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	// Start server
	if err := startServer(grpcServer, authServer, cfg.Port); err != nil {
		log.Printf("gRPC server stopped: %v", err)
	}
}
