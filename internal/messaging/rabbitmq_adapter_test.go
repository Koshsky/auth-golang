package messaging

import (
	"fmt"
	"testing"

	"github.com/Koshsky/subs-service/auth-service/internal/config"
	"github.com/Koshsky/subs-service/auth-service/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	messagingMocks "github.com/Koshsky/subs-service/auth-service/internal/messaging/mocks"
)

type RabbitMQAdapterTestSuite struct {
	suite.Suite
	mockPublisher *messagingMocks.IRabbitMQPublisher
	mockConn      *messagingMocks.IRabbitMQConn
	adapter       IMessageBroker
	config        config.RabbitMQConfig
	testUser      *models.User
}

func (suite *RabbitMQAdapterTestSuite) SetupSuite() {
	suite.testUser = &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
}

func (suite *RabbitMQAdapterTestSuite) SetupTest() {
	suite.config = config.RabbitMQConfig{
		Exchange: "test_exchange",
	}
	suite.mockPublisher = messagingMocks.NewIRabbitMQPublisher(suite.T())
	suite.mockConn = messagingMocks.NewIRabbitMQConn(suite.T())
	suite.adapter = &RabbitMQAdapter{
		publisher: suite.mockPublisher,
		conn:      suite.mockConn,
		config:    suite.config,
	}
}

func (suite *RabbitMQAdapterTestSuite) TearDownTest() {
	suite.mockPublisher.AssertExpectations(suite.T())
	suite.mockConn.AssertExpectations(suite.T())
}

// ===== MOCK HELPER FUNCTIONS =====

// mockPublisherPublish mock publisher.Publish(data, routingKeys, options, options)
func (suite *RabbitMQAdapterTestSuite) mockPublisherPublish(data []byte, routingKeys []string, err error) {
	suite.mockPublisher.On("Publish",
		data,
		routingKeys,
		mock.AnythingOfType("func(*rabbitmq.PublishOptions)"),
		mock.AnythingOfType("func(*rabbitmq.PublishOptions)"),
	).Return(err)
}

// mockClose mocks both publisher.Close() and conn.Close()
func (suite *RabbitMQAdapterTestSuite) mockClose(err error) {
	suite.mockPublisher.On("Close").Return()
	suite.mockConn.On("Close").Return(err)
}

// ===== CONSTRUCTOR TESTS =====

func (suite *RabbitMQAdapterTestSuite) TestNewRabbitMQAdapter_InvalidConfig() {
	// Arrange
	cfg := config.RabbitMQConfig{
		URL:      "invalid://url",
		Exchange: "test_exchange",
	}

	// Act
	adapter, err := NewRabbitMQAdapter(cfg)

	// Assert
	suite.Require().Error(err)
	suite.Require().Nil(adapter)
	suite.Contains(err.Error(), "failed to connect to RabbitMQ")
}

// ===== PUBLISH USER CREATED TESTS =====

func (suite *RabbitMQAdapterTestSuite) TestPublishUserCreated_Success() {
	// Arrange
	expectedData := []byte(`{"user_id":"` + suite.testUser.ID.String() + `","email":"test@example.com"}`)
	expectedRoutingKeys := []string{"user.created"}

	suite.mockPublisherPublish(expectedData, expectedRoutingKeys, nil)

	// Act
	err := suite.adapter.PublishUserCreated(suite.testUser)

	// Assert
	suite.Require().NoError(err)
	suite.mockPublisher.AssertExpectations(suite.T())
}

func (suite *RabbitMQAdapterTestSuite) TestPublishUserCreated_NilPublisher() {
	// Arrange
	adapter := &RabbitMQAdapter{
		publisher: nil,
		conn:      suite.mockConn,
		config:    suite.config,
	}

	// Act
	err := adapter.PublishUserCreated(suite.testUser)

	// Assert
	suite.Require().Error(err)
	suite.Contains(err.Error(), "publisher is not initialized")
}

func (suite *RabbitMQAdapterTestSuite) TestPublishUserCreated_PublisherError() {
	// Arrange
	expectedError := fmt.Errorf("publisher error")
	suite.mockPublisherPublish([]byte(`{"user_id":"`+suite.testUser.ID.String()+`","email":"test@example.com"}`), []string{"user.created"}, expectedError)

	// Act
	err := suite.adapter.PublishUserCreated(suite.testUser)

	// Assert
	suite.Require().Error(err)
	suite.Contains(err.Error(), "publisher error")
	suite.mockPublisher.AssertExpectations(suite.T())
}

func (suite *RabbitMQAdapterTestSuite) TestPublishUserCreated_NilUser() {
	// Arrange
	var user *models.User = nil

	// Act
	err := suite.adapter.PublishUserCreated(user)

	// Assert
	suite.Require().Error(err)
	suite.Contains(err.Error(), "user cannot be nil")
}

// ===== PUBLISH USER DELETED TESTS =====

func (suite *RabbitMQAdapterTestSuite) TestPublishUserDeleted_Success() {
	// Arrange
	suite.mockPublisherPublish([]byte(`{"user_id":"`+suite.testUser.ID.String()+`"}`), []string{"user.deleted"}, nil)

	// Act
	err := suite.adapter.PublishUserDeleted(suite.testUser)

	// Assert
	suite.Require().NoError(err)
	suite.mockPublisher.AssertExpectations(suite.T())
}

func (suite *RabbitMQAdapterTestSuite) TestPublishUserDeleted_NilPublisher() {
	// Arrange
	adapter := &RabbitMQAdapter{
		publisher: nil,
		conn:      suite.mockConn,
		config:    suite.config,
	}

	// Act
	err := adapter.PublishUserDeleted(suite.testUser)

	// Assert
	suite.Require().Error(err)
	suite.Contains(err.Error(), "publisher is not initialized")
}

func (suite *RabbitMQAdapterTestSuite) TestPublishUserDeleted_PublisherError() {
	// Arrange
	expectedError := fmt.Errorf("publisher error")
	suite.mockPublisherPublish([]byte(`{"user_id":"`+suite.testUser.ID.String()+`"}`), []string{"user.deleted"}, expectedError)

	// Act
	err := suite.adapter.PublishUserDeleted(suite.testUser)

	// Assert
	suite.Require().Error(err)
	suite.Contains(err.Error(), "publisher error")
	suite.mockPublisher.AssertExpectations(suite.T())
}

func (suite *RabbitMQAdapterTestSuite) TestPublishUserDeleted_NilUser() {
	// Arrange
	var user *models.User = nil

	// Act
	err := suite.adapter.PublishUserDeleted(user)

	// Assert
	suite.Require().Error(err)
	suite.Contains(err.Error(), "user cannot be nil")
}

// ===== CLOSE TESTS =====

func (suite *RabbitMQAdapterTestSuite) TestClose_Success() {
	// Arrange
	suite.mockClose(nil)

	// Act & Assert
	suite.NotPanics(func() {
		suite.adapter.Close()
	})
	suite.mockPublisher.AssertExpectations(suite.T())
	suite.mockConn.AssertExpectations(suite.T())
}

func TestRabbitMQAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(RabbitMQAdapterTestSuite))
}
