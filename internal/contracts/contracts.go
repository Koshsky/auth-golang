package contracts

import (
	contracts_interceptor "github.com/Koshsky/subs-service/auth-service/internal/contracts/interceptor"
	contracts_messaging "github.com/Koshsky/subs-service/auth-service/internal/contracts/messaging"
	contracts_repository "github.com/Koshsky/subs-service/auth-service/internal/contracts/repository"
	contracts_service "github.com/Koshsky/subs-service/auth-service/internal/contracts/service"
)

// Repository layer
type IUserRepository = contracts_repository.IUserRepository
type IDatabase = contracts_repository.IDatabase

// Service layer
type IAuthService = contracts_service.IAuthService

// Messaging layer
type IMessageBroker = contracts_messaging.IMessageBroker
type IRabbitMQConn = contracts_messaging.IRabbitMQConn
type IRabbitMQPublisher = contracts_messaging.IRabbitMQPublisher

// Interceptors layer
type ILoggingInterceptor = contracts_interceptor.ILoggingInterceptor
type IUnaryInterceptor = contracts_interceptor.IUnaryInterceptor
type IStreamInterceptor = contracts_interceptor.IStreamInterceptor
type IInterceptorManager = contracts_interceptor.IInterceptorManager
