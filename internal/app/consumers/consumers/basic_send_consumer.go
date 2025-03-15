package consumers

import (
	"github.com/Filiphasan/rabbitmqgolang/internal/app/consumers"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/consumers/models"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/services/interfaces"
	"github.com/Filiphasan/rabbitmqgolang/internal/rabbitmq"
	"github.com/Filiphasan/rabbitmqgolang/pkg/constants"
	"go.uber.org/zap"
)

type BasicSendConsumer struct {
	consumers.BaseConsumer[models.QueueBasicMessage]
	myService interfaces.IMyService
}

func NewBasicSendConsumer(logger *zap.Logger, cm *rabbitmq.ConnectionManager, myService interfaces.IMyService) *BasicSendConsumer {
	queue := models.NewConsumerQueueInfo(constants.BasicSendQueueName)
	exchange := models.NewBlankConsumerExchangeInfoWithName()
	retry := models.NewConsumerRetryInfo(false)

	return &BasicSendConsumer{
		BaseConsumer: *consumers.NewBaseConsumer[models.QueueBasicMessage](logger, cm, queue, exchange, retry),
		myService:    myService,
	}
}

func (c *BasicSendConsumer) ConsumeMessage(message *models.QueueBasicMessage) (*models.ConsumeResult, error) {
	err := c.myService.DoBasicSendConsumerWork(message)
	if err != nil {
		c.Logger.Error("Failed to process message", zap.Error(err))
		return nil, err
	}

	c.Logger.Info("Message processed successfully", zap.String("message", message.Message))
	return models.NewConsumeResultDone(), nil
}
