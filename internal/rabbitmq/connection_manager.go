package rabbitmq

import (
	"github.com/Filiphasan/rabbitmqgolang/configs"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"math"
	"sync"
	"time"
)

const MaxRetries = 4

type ConnectionManager struct {
	appConfig *configs.AppConfig
	mtx       sync.Mutex
	cnn       *amqp091.Connection
	logger    *zap.Logger
}

func NewConnectionManager(appConfig *configs.AppConfig, logger *zap.Logger) *ConnectionManager {
	return &ConnectionManager{
		appConfig: appConfig,
		logger:    logger,
		mtx:       sync.Mutex{},
	}
}

func (c *ConnectionManager) IsConnected() bool {
	return c.cnn != nil && !c.cnn.IsClosed()
}

func (c *ConnectionManager) CreateChannel() (*amqp091.Channel, error) {
	if c.IsConnected() {
		return c.cnn.Channel()
	}

	err := c.Connect()
	if err != nil {
		return nil, err
	}

	return c.cnn.Channel()
}

func (c *ConnectionManager) Connect() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.IsConnected() {
		return nil
	}

	for i := 1; i <= MaxRetries; i++ {
		cnn, err := amqp091.DialConfig(c.appConfig.GetRabbitMqConnectionStr(), amqp091.Config{
			Properties: amqp091.Table{
				"connection_name": "golang-rabbitmq",
			},
		})
		if err != nil {
			c.logger.Error("Failed to connect to RabbitMQ", zap.Error(err), zap.Int("Attempt", i))
			if i == MaxRetries {
				return err
			}

			waitDuration := time.Duration(math.Pow(2, float64(i))) * time.Second
			time.Sleep(waitDuration)
			continue
		}
		c.cnn = cnn
		c.logger.Info("Connected to RabbitMQ")
		return nil
	}

	return nil
}
