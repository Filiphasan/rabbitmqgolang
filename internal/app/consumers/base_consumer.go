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

func (c *BaseConsumer[T]) ConsumeMessage(message *T) (*models.ConsumeResult, error) {
	return nil, nil
}

func (c *BaseConsumer[T]) initializeChannel() error {
	channel, err := c.cm.CreateChannel()
	if err != nil {
		c.Logger.Error("Failed to create channel", zap.Error(err))
		return err
	}

	closeChannel := make(chan *amqp091.Error)
	channel.NotifyClose(closeChannel)
	go func() {
		for err := range closeChannel {
			c.Logger.Error("Channel closed", zap.Error(err), zap.String("queue", c.Queue.Name))
		}
	}()

	c.channel = channel
	return nil
}

func (c *BaseConsumer[T]) declareQueueAmdExchange() (amqp091.Queue, error) {
	queue, err := c.channel.QueueDeclare(
		c.Queue.Name,
		c.Queue.Durable,
		c.Queue.AutoDelete,
		c.Queue.Exclusive,
		false,
		c.Queue.Arguments)
	if err != nil {
		c.Logger.Error("Failed to declare queue", zap.Error(err), zap.String("queue", c.Queue.Name))
		return amqp091.Queue{}, err
	}

	if c.Exchange.Name != "" {
		err = c.channel.ExchangeDeclare(
			c.Exchange.Name,
			c.Exchange.Exchange,
			c.Exchange.Durable,
			false,
			false,
			false,
			c.Exchange.Arguments,
		)
		if err != nil {
			c.Logger.Error("Failed to declare exchange", zap.Error(err))
			return amqp091.Queue{}, err
		}

		err = c.channel.QueueBind(c.Queue.Name, c.Exchange.RoutingKey, c.Exchange.Name, false, nil)
		if err != nil {
			c.Logger.Error("Failed to bind queue to exchange", zap.Error(err), zap.String("queue", c.Queue.Name), zap.String("exchange", c.Exchange.Name))
			return amqp091.Queue{}, err
		}
	}

	if c.Retry.UseRetry {
		delayedExName := c.Queue.Name + "_retry"
		err = c.channel.ExchangeDeclare(delayedExName, "x-delayed-message", true, false, false, false, amqp091.Table{
			"x-delayed-type": "direct",
		})
		if err != nil {
			c.Logger.Error("Failed to declare delayed exchange", zap.Error(err))
			return amqp091.Queue{}, err
		}
		err = c.channel.QueueBind(c.Queue.Name, "", delayedExName, false, nil)
		if err != nil {
			c.Logger.Error("Failed to bind queue to delayed exchange", zap.Error(err), zap.String("queue", c.Queue.Name), zap.String("exchange", delayedExName))
			return amqp091.Queue{}, err
		}
	}

	return queue, nil
}

func (c *BaseConsumer[T]) StartConsuming() error {
	err := c.initializeChannel()
	if err != nil {
		c.Logger.Error("Failed to initialize channel", zap.Error(err))
		return err
	}

	_, err = c.declareQueueAmdExchange()
	if err != nil {
		c.Logger.Error("Failed to declare queue and exchange", zap.Error(err), zap.String("queue", c.Queue.Name))
		return err
	}

	err = c.channel.Qos(0, 1, false)
	if err != nil {
		c.Logger.Error("Failed to set QoS", zap.Error(err), zap.String("queue", c.Queue.Name))
		return err
	}

	messages, err := c.channel.Consume(c.Queue.Name, "", false, c.Queue.Exclusive, false, false, nil)
	if err != nil {
		c.Logger.Error("Failed to consume messages", zap.Error(err), zap.String("queue", c.Queue.Name))
		return err
	}

	go func() {
		for d := range messages {
			c.handleMessage(d)
		}
	}()

	c.Logger.Info("Consumer started", zap.String("queue", c.Queue.Name))
	return nil
}

func (c *BaseConsumer[T]) handleMessage(d amqp091.Delivery) {
	var msg T
	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		c.retryOrRejectMessage(d, 0)
		return
	}

	cResult, err := c.ConsumeMessage(&msg)
	if err != nil {
		c.retryOrRejectMessage(d, 0)
	} else if cResult != nil && cResult.Result == models.Retry {
		c.retryOrRejectMessage(d, cResult.DelayMs)
	} else if cResult != nil && cResult.Result == models.Delay && c.Retry.UseRetry {
		c.delayMessage(d, cResult.DelayMs)
	} else {
		_ = d.Ack(false)
	}

	c.Logger.Info("Received a message", zap.String("body", string(d.Body)), zap.String("queue", c.Queue.Name))
}

func (c *BaseConsumer[T]) delayMessage(d amqp091.Delivery, delayMs int) {
	if !c.Retry.UseRetry {
		err := d.Reject(false)
		if err != nil {
			c.Logger.Error("Failed to reject message", zap.Error(err), zap.String("queue", c.Queue.Name), zap.String("message", string(d.Body)), zap.String("message_id", d.MessageId))
		}
		return
	}

	if delayMs <= 0 {
		delayMs = c.Retry.Delay
	}
	delayedExName := c.Queue.Name + "_retry"
	d.Headers["x-delay"] = delayMs
	err := c.channel.Publish(delayedExName, "", false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        d.Body,
	})
	if err != nil {
		c.Logger.Error("Failed to publish message to retry exchange", zap.Error(err), zap.String("queue", c.Queue.Name), zap.String("message", string(d.Body)), zap.String("message_id", d.MessageId))
		return
	}

	err = d.Ack(false)
	if err != nil {
		c.Logger.Error("Failed to acknowledge message", zap.Error(err), zap.String("queue", c.Queue.Name), zap.String("message", string(d.Body)), zap.String("message_id", d.MessageId))
		return
	}
}

func (c *BaseConsumer[T]) retryOrRejectMessage(d amqp091.Delivery, delayMs int) {
	if !c.Retry.UseRetry {
		err := d.Reject(false)
		if err != nil {
			c.Logger.Error("Failed to reject message", zap.Error(err), zap.String("queue", c.Queue.Name), zap.String("message", string(d.Body)), zap.String("message_id", d.MessageId))
		}
		return
	}

	currentRetryCount := 0
	if count, ok := d.Headers["x-retry-count"].(int); ok {
		currentRetryCount = count
	}

	if currentRetryCount > c.Retry.MaxRetry {
		err := d.Reject(false)
		if err != nil {
			c.Logger.Error("Failed to reject message", zap.Error(err), zap.String("queue", c.Queue.Name), zap.String("message", string(d.Body)), zap.String("message_id", d.MessageId))
		}
		return
	}

	d.Headers["x-retry-count"] = currentRetryCount + 1
	c.delayMessage(d, delayMs)
}

func (c *BaseConsumer[T]) Close() {
	if c.channel == nil || c.channel.IsClosed() {
		return
	}

	err := c.channel.Close()
	if err != nil {
		c.Logger.Error("Failed to close channel", zap.Error(err), zap.String("queue", c.Queue.Name))
	}
}
