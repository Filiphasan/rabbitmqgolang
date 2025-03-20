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

type BasicPublishSecondConsumer struct {
	consumers.BaseConsumer[models.QueueBasicMessage]
	myService interfaces.IMyService
}

func NewBasicPublishSecondConsumer(logger *zap.Logger, cm *rabbitmq.ConnectionManager, myService interfaces.IMyService) *BasicPublishSecondConsumer {
	queue := models.NewConsumerQueueInfo(constants.BasicPublishSecondQueueName)
	exchange := models.NewConsumerExchangeInfo(
		constants.BasicPublishExchangeName,
		constants.BasicPublishRoutingKey,
		amqp091.ExchangeFanout,
	)
	retry := models.NewConsumerRetryInfo(true)

	return &BasicPublishSecondConsumer{
		BaseConsumer: *consumers.NewBaseConsumer[models.QueueBasicMessage](logger, cm, queue, exchange, retry),
		myService:    myService,
	}
}
