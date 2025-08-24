package messaging

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Koshsky/subs-service/auth-service/internal/config"
	"github.com/Koshsky/subs-service/auth-service/internal/models"
	"github.com/google/uuid"
	"github.com/wagslane/go-rabbitmq"
)

// RabbitMQAdapter implements IMessageBroker for RabbitMQ
type RabbitMQAdapter struct {
	publisher IRabbitMQPublisher
	conn      IRabbitMQConn
	config    config.RabbitMQConfig
}

type UserCreatedEvent struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

type UserDeletedEvent struct {
	UserID uuid.UUID `json:"user_id"`
}

// NewRabbitMQAdapter creates a new RabbitMQ adapter
func NewRabbitMQAdapter(rabbitmqConfig config.RabbitMQConfig) (IMessageBroker, error) {
	// Create connection with automatic reconnection
	conn, err := rabbitmq.NewConn(
		rabbitmqConfig.URL,
		rabbitmq.WithConnectionOptionsLogging,
		rabbitmq.WithConnectionOptionsReconnectInterval(5), // 5 seconds between reconnection attempts
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	// Create publisher with automatic reconnection
	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName(rabbitmqConfig.Exchange),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsExchangeKind("topic"),
		rabbitmq.WithPublisherOptionsExchangeDurable,
	)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create publisher: %v", err)
	}

	return &RabbitMQAdapter{
		publisher: publisher,
		conn:      conn,
		config:    rabbitmqConfig,
	}, nil
}

// PublishUserCreated publishes user created event to RabbitMQ
func (r *RabbitMQAdapter) PublishUserCreated(user *models.User) error {
	if r.publisher == nil {
		return errors.New("publisher is not initialized")
	}

	if user == nil {
		return errors.New("user cannot be nil")
	}

	event := UserCreatedEvent{
		UserID: user.ID,
		Email:  user.Email,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal user created event: %v", err)
	}

	err = r.publisher.Publish(
		body,
		[]string{"user.created"},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange(r.config.Exchange),
	)
	if err != nil {
		return fmt.Errorf("failed to publish user created event: %v", err)
	}

	return nil
}

func (r *RabbitMQAdapter) PublishUserDeleted(user *models.User) error {
	if r.publisher == nil {
		return errors.New("publisher is not initialized")
	}

	if user == nil {
		return errors.New("user cannot be nil")
	}

	event := UserDeletedEvent{
		UserID: user.ID,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal user deleted event: %v", err)
	}

	err = r.publisher.Publish(
		body,
		[]string{"user.deleted"},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange(r.config.Exchange),
	)
	if err != nil {
		return fmt.Errorf("failed to publish user deleted event: %v", err)
	}

	return nil
}

// Close closes the RabbitMQ connection
func (r *RabbitMQAdapter) Close() {
	if r.publisher != nil {
		r.publisher.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
