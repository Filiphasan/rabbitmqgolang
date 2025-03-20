package implementations

import (
	"context"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/services/interfaces"
	"github.com/Filiphasan/rabbitmqgolang/internal/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type MqService struct {
	cm     *rabbitmq.ConnectionManager
	logger *zap.Logger
}

func NewMqService(cm *rabbitmq.ConnectionManager, logger *zap.Logger) *MqService {
	return &MqService{
		cm:     cm,
		logger: logger,
	}
}

func (m *MqService) SendMessage(ctx context.Context, model *interfaces.SendMessageModel) error {
	channel, err := m.cm.CreateChannel()
	if err != nil {
		return err
	}
	defer func(channel *amqp091.Channel) {
		err := channel.Close()
		if err != nil {
			m.logger.Error("Error closing channel", zap.Error(err))
		}
	}(channel)

	msg := amqp091.Publishing{
		ContentType:  "application/json",
		Body:         model.Message,
		MessageId:    model.MessageId,
		DeliveryMode: model.DeliveryMode,
		Priority:     model.Priority,
	}
	err = channel.PublishWithContext(ctx, "", model.QueueName, false, false, msg)
	if err != nil {
		return err
	}

	return nil
}

func (m *MqService) PublishMessage(ctx context.Context, model *interfaces.PublishMessageModel) error {
	channel, err := m.cm.CreateChannel()
	if err != nil {
		return err
	}
	defer func(channel *amqp091.Channel) {
		err := channel.Close()
		if err != nil {
			m.logger.Error("Error closing channel", zap.Error(err))
		}
	}(channel)

	err = channel.ExchangeDeclare(model.ExchangeName, model.ExchangeType, model.Durable, false, false, false, model.Arguments)
	if err != nil {
		return err
	}

	msg := amqp091.Publishing{
		ContentType:  "application/json",
		Body:         model.Message,
		MessageId:    model.MessageId,
		DeliveryMode: model.DeliveryMode,
		Priority:     model.Priority,
	}

	err = channel.PublishWithContext(ctx, model.ExchangeName, model.RoutingKey, false, false, msg)
	if err != nil {
		return err
	}

	return nil
}
