package interfaces

import (
	"context"
	"encoding/json"
	"github.com/rabbitmq/amqp091-go"
)

type IMqService interface {
	SendMessage(ctx context.Context, model SendMessageModel) error
	PublishMessage(ctx context.Context, model PublishMessageModel) error
}

type SendMessageModel struct {
	QueueName    string
	Message      []byte
	MessageId    string
	DeliveryMode uint8 // 0 = transient, 1 = persistent
	Priority     byte
}

func NewSendMessageModel[T any](queueName string, message *T, messageId string) (*SendMessageModel, error) {
	msg, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	return &SendMessageModel{
		QueueName:    queueName,
		Message:      msg,
		MessageId:    messageId,
		DeliveryMode: amqp091.Transient,
		Priority:     0,
	}, nil
}

type PublishMessageModel struct {
	Message      []byte
	ExchangeName string
	RoutingKey   string
	ExchangeType string
	Durable      bool
	Arguments    amqp091.Table
	MessageId    string
	DeliveryMode uint8
	Priority     byte
}

func NewPublishMessageModel[T any](exchangeName string, routingKey string, message *T, messageId string) (*PublishMessageModel, error) {
	msg, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	return &PublishMessageModel{
		Message:      msg,
		ExchangeName: exchangeName,
		RoutingKey:   routingKey,
		ExchangeType: amqp091.ExchangeFanout,
		MessageId:    messageId,
		Durable:      true,
		DeliveryMode: amqp091.Transient,
		Priority:     0,
		Arguments:    amqp091.Table{},
	}, nil
}
