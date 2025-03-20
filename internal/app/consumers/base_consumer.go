package consumers

import (
	"encoding/json"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/consumers/models"
	"github.com/Filiphasan/rabbitmqgolang/internal/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type IBaseConsumer[T interface{}] interface {
	StartConsuming() error
	Close()
}

type BaseConsumer[T interface{}] struct {
	channel  *amqp091.Channel
	Logger   *zap.Logger
	cm       *rabbitmq.ConnectionManager
	Queue    *models.ConsumerQueueInfo
	Exchange *models.ConsumerExchangeInfo
	Retry    *models.ConsumerRetryInfo
}

func NewBaseConsumer[T interface{}](logger *zap.Logger, cm *rabbitmq.ConnectionManager, queue *models.ConsumerQueueInfo, exchange *models.ConsumerExchangeInfo, retry *models.ConsumerRetryInfo) *BaseConsumer[T] {
	return &BaseConsumer[T]{
		Logger:   logger,
		cm:       cm,
		Queue:    queue,
		Exchange: exchange,
		Retry:    retry,
	}
}

func (bc *BaseConsumer[T]) ConsumeMessage(message *T) (*models.ConsumeResult, error) {
	return nil, nil
}

func (bc *BaseConsumer[T]) initializeChannel() error {
	channel, err := bc.cm.CreateChannel()
	if err != nil {
		bc.Logger.Error("Failed to create channel", zap.Error(err))
		return err
	}

	closeChannel := make(chan *amqp091.Error)
	channel.NotifyClose(closeChannel)
	go func() {
		for err := range closeChannel {
			bc.Logger.Error("Channel closed", zap.Error(err), zap.String("queue", bc.Queue.Name))
		}
	}()

	bc.channel = channel
	return nil
}

func (bc *BaseConsumer[T]) declareQueueAmdExchange() (amqp091.Queue, error) {
	queue, err := bc.channel.QueueDeclare(
		bc.Queue.Name,
		bc.Queue.Durable,
		bc.Queue.AutoDelete,
		bc.Queue.Exclusive,
		false,
		bc.Queue.Arguments)
	if err != nil {
		bc.Logger.Error("Failed to declare queue", zap.Error(err), zap.String("queue", bc.Queue.Name))
		return amqp091.Queue{}, err
	}

	if bc.Exchange.Name != "" {
		err = bc.channel.ExchangeDeclare(
			bc.Exchange.Name,
			bc.Exchange.Exchange,
			bc.Exchange.Durable,
			false,
			false,
			false,
			bc.Exchange.Arguments,
		)
		if err != nil {
			bc.Logger.Error("Failed to declare exchange", zap.Error(err))
			return amqp091.Queue{}, err
		}

		err = bc.channel.QueueBind(bc.Queue.Name, bc.Exchange.RoutingKey, bc.Exchange.Name, false, nil)
		if err != nil {
			bc.Logger.Error("Failed to bind queue to exchange", zap.Error(err), zap.String("queue", bc.Queue.Name), zap.String("exchange", bc.Exchange.Name))
			return amqp091.Queue{}, err
		}
	}

	if bc.Retry.UseRetry {
		delayedExName := bc.Queue.Name + "_retry"
		err = bc.channel.ExchangeDeclare(delayedExName, "x-delayed-message", true, false, false, false, amqp091.Table{
			"x-delayed-type": "direct",
		})
		if err != nil {
			bc.Logger.Error("Failed to declare delayed exchange", zap.Error(err))
			return amqp091.Queue{}, err
		}
		err = bc.channel.QueueBind(bc.Queue.Name, "", delayedExName, false, nil)
		if err != nil {
			bc.Logger.Error("Failed to bind queue to delayed exchange", zap.Error(err), zap.String("queue", bc.Queue.Name), zap.String("exchange", delayedExName))
			return amqp091.Queue{}, err
		}
	}

	return queue, nil
}

func (bc *BaseConsumer[T]) StartConsuming() error {
	err := bc.initializeChannel()
	if err != nil {
		bc.Logger.Error("Failed to initialize channel", zap.Error(err))
		return err
	}

	_, err = bc.declareQueueAmdExchange()
	if err != nil {
		bc.Logger.Error("Failed to declare queue and exchange", zap.Error(err), zap.String("queue", bc.Queue.Name))
		return err
	}

	err = bc.channel.Qos(0, 1, false)
	if err != nil {
		bc.Logger.Error("Failed to set QoS", zap.Error(err), zap.String("queue", bc.Queue.Name))
		return err
	}

	messages, err := bc.channel.Consume(bc.Queue.Name, "", false, bc.Queue.Exclusive, false, false, nil)
	if err != nil {
		bc.Logger.Error("Failed to consume messages", zap.Error(err), zap.String("queue", bc.Queue.Name))
		return err
	}

	go func() {
		for d := range messages {
			bc.handleMessage(d)
		}
	}()

	bc.Logger.Info("Consumer started", zap.String("queue", bc.Queue.Name))
	return nil
}

func (bc *BaseConsumer[T]) handleMessage(d amqp091.Delivery) {
	var msg T
	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		bc.retryOrRejectMessage(d, 0)
		return
	}

	cResult, err := bc.ConsumeMessage(&msg)
	if err != nil {
		bc.retryOrRejectMessage(d, 0)
	} else if cResult != nil && cResult.Result == models.Retry {
		bc.retryOrRejectMessage(d, cResult.DelayMs)
	} else if cResult != nil && cResult.Result == models.Delay && bc.Retry.UseRetry {
		bc.delayMessage(d, cResult.DelayMs)
	} else {
		_ = d.Ack(false)
	}

	bc.Logger.Info("Received a message", zap.String("body", string(d.Body)), zap.String("queue", bc.Queue.Name))
}

func (bc *BaseConsumer[T]) delayMessage(d amqp091.Delivery, delayMs int) {
	if !bc.Retry.UseRetry {
		err := d.Reject(false)
		if err != nil {
			bc.Logger.Error("Failed to reject message", zap.Error(err), zap.String("queue", bc.Queue.Name), zap.String("message", string(d.Body)), zap.String("message_id", d.MessageId))
		}
		return
	}

	if delayMs <= 0 {
		delayMs = bc.Retry.Delay
	}
	delayedExName := bc.Queue.Name + "_retry"
	d.Headers["x-delay"] = delayMs
	err := bc.channel.Publish(delayedExName, "", false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        d.Body,
	})
	if err != nil {
		bc.Logger.Error("Failed to publish message to retry exchange", zap.Error(err), zap.String("queue", bc.Queue.Name), zap.String("message", string(d.Body)), zap.String("message_id", d.MessageId))
		return
	}

	err = d.Ack(false)
	if err != nil {
		bc.Logger.Error("Failed to acknowledge message", zap.Error(err), zap.String("queue", bc.Queue.Name), zap.String("message", string(d.Body)), zap.String("message_id", d.MessageId))
		return
	}
}

func (bc *BaseConsumer[T]) retryOrRejectMessage(d amqp091.Delivery, delayMs int) {
	if !bc.Retry.UseRetry {
		err := d.Reject(false)
		if err != nil {
			bc.Logger.Error("Failed to reject message", zap.Error(err), zap.String("queue", bc.Queue.Name), zap.String("message", string(d.Body)), zap.String("message_id", d.MessageId))
		}
		return
	}

	currentRetryCount := 0
	if count, ok := d.Headers["x-retry-count"].(int); ok {
		currentRetryCount = count
	}

	if currentRetryCount > bc.Retry.MaxRetry {
		err := d.Reject(false)
		if err != nil {
			bc.Logger.Error("Failed to reject message", zap.Error(err), zap.String("queue", bc.Queue.Name), zap.String("message", string(d.Body)), zap.String("message_id", d.MessageId))
		}
		return
	}

	d.Headers["x-retry-count"] = currentRetryCount + 1
	bc.delayMessage(d, delayMs)
}

func (bc *BaseConsumer[T]) Close() {
	if bc.channel == nil || bc.channel.IsClosed() {
		return
	}

	err := bc.channel.Close()
	if err != nil {
		bc.Logger.Error("Failed to close channel", zap.Error(err), zap.String("queue", bc.Queue.Name))
	}
}
