package models

import "github.com/rabbitmq/amqp091-go"

type ConsumerQueueInfo struct {
	Name       string
	Durable    bool
	Exclusive  bool
	AutoDelete bool
	Arguments  amqp091.Table
}

func NewConsumerQueueInfo(name string) *ConsumerQueueInfo {
	return &ConsumerQueueInfo{
		Name:       name,
		Durable:    true,
		Exclusive:  false,
		AutoDelete: false,
		Arguments:  amqp091.Table{},
	}
}

type ConsumerExchangeInfo struct {
	Name       string
	RoutingKey string
	Exchange   string
	Durable    bool
	Arguments  amqp091.Table
}

func NewBlankConsumerExchangeInfoWithName() *ConsumerExchangeInfo {
	return &ConsumerExchangeInfo{
		Name:       "",
		RoutingKey: "",
		Exchange:   "",
		Durable:    true,
		Arguments:  amqp091.Table{},
	}
}

func NewConsumerExchangeInfo(name, routingKey, exchange string) *ConsumerExchangeInfo {
	return &ConsumerExchangeInfo{
		Name:       name,
		RoutingKey: routingKey,
		Exchange:   exchange,
		Durable:    true,
		Arguments:  amqp091.Table{},
	}
}

type ConsumerRetryInfo struct {
	UseRetry bool
	MaxRetry int
	Delay    int
}

func NewConsumerRetryInfo(useRetry bool) *ConsumerRetryInfo {
	return &ConsumerRetryInfo{
		UseRetry: useRetry,
		MaxRetry: 3,
		Delay:    2 * 60 * 1000,
	}
}
