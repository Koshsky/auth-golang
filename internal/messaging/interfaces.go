package messaging

import (
	"context"

	"github.com/Koshsky/subs-service/auth-service/internal/models"
	"github.com/wagslane/go-rabbitmq"
)

//go:generate mockery --name=IMessageBroker --output=./mocks --outpkg=mocks --filename=IMessageBroker.go
type IMessageBroker interface {
	PublishUserCreated(user *models.User) error
	PublishUserDeleted(user *models.User) error
	Close()
}

//go:generate mockery --name=IRabbitMQConn --output=./mocks --outpkg=mocks --filename=IRabbitMQConn.go
type IRabbitMQConn interface {
	Close() error
}

//go:generate mockery --name=IRabbitMQPublisher --output=./mocks --outpkg=mocks --filename=IRabbitMQPublisher.go
type IRabbitMQPublisher interface {
	Publish(data []byte, routingKeys []string, optionFuncs ...func(*rabbitmq.PublishOptions)) error
	PublishWithContext(ctx context.Context, data []byte, routingKeys []string, optionFuncs ...func(*rabbitmq.PublishOptions)) error
	Close()
	NotifyPublish(handler func(p rabbitmq.Confirmation))
	NotifyReturn(handler func(r rabbitmq.Return))
}

// Interface compliance checks - will fail at compile time if interfaces are not implemented
var _ IMessageBroker = (*RabbitMQAdapter)(nil)
var _ IRabbitMQConn = (*rabbitmq.Conn)(nil)
var _ IRabbitMQPublisher = (*rabbitmq.Publisher)(nil)
