package main

import (
	"context"
	"github.com/Filiphasan/rabbitmqgolang/configs"
	baseConsumers "github.com/Filiphasan/rabbitmqgolang/internal/app/consumers"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/consumers/consumers"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/services/implementations"
	"github.com/Filiphasan/rabbitmqgolang/internal/logger"
	"github.com/Filiphasan/rabbitmqgolang/internal/rabbitmq"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Start the server
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	appConfig := configs.LoadConfig()
	logg := logger.UseLogger(appConfig)
	defer func(logg *zap.Logger) {
		_ = logg.Sync()
	}(logg)

	myService := implementations.NewMyService(logg)

	connectionManager := rabbitmq.NewConnectionManager(appConfig, logg)

	var consumerList []baseConsumers.IBaseConsumer[interface{}]
	consumerList = append(consumerList, consumers.NewBasicSendConsumer(logg, connectionManager, myService))

	for _, consumer := range consumerList {
		go func(c baseConsumers.IBaseConsumer[interface{}]) {
			err := c.StartConsuming()
			if err != nil {
				logg.Error("Failed to start consuming", zap.Error(err))
			}
		}(consumer)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs

	logg.Info("Received shutdown signal, shutting down gracefully...")
	for _, consumer := range consumerList {
		consumer.Close()
	}

	connectionManager.Close()
}
