package main

import (
	"context"
	"github.com/Filiphasan/rabbitmqgolang/configs"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/consumers/consumers"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/services/implementations"
	"github.com/Filiphasan/rabbitmqgolang/internal/rabbitmq"
	"go.uber.org/zap"
	"time"
)

func main() {
	// Start the server
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configs.LoadConfig()
	appConfig := configs.GetConfig()
	logger := &zap.Logger{}

	myService := implementations.NewMyService(logger)

	connectionManager := rabbitmq.NewConnectionManager(appConfig, logger)

	basicSendConsumer := consumers.NewBasicSendConsumer(logger, connectionManager, myService)

	go func() {
		err := basicSendConsumer.StartConsuming()
		if err != nil {
			logger.Error("Failed to start consuming", zap.Error(err))
		}
	}()
}
