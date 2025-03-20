package consumers

import (
	"github.com/Filiphasan/rabbitmqgolang/internal/app/consumers"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/consumers/models"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/services/interfaces"
	"github.com/Filiphasan/rabbitmqgolang/internal/rabbitmq"
	"github.com/Filiphasan/rabbitmqgolang/pkg/constants"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type BasicPublishFirstConsumer struct {
	consumers.BaseConsumer[models.QueueBasicMessage]
	myService interfaces.IMyService
}

func NewBasicPublishFirstConsumer(logger *zap.Logger, cm *rabbitmq.ConnectionManager, myService interfaces.IMyService) *BasicPublishFirstConsumer {
	queue := models.NewConsumerQueueInfo(constants.BasicPublishFirstQueueName)
	exchange := models.NewConsumerExchangeInfo(constants.BasicPublishExchangeName, constants.BasicPublishRoutingKey, amqp091.ExchangeFanout)
	retry := models.NewConsumerRetryInfo(true)
	return &BasicPublishFirstConsumer{
		BaseConsumer: *consumers.NewBaseConsumer[models.QueueBasicMessage](logger, cm, queue, exchange, retry),
		myService:    myService,
	}
}
