package main

import (
	"github.com/Filiphasan/rabbitmqgolang/configs"
	baseConsumers "github.com/Filiphasan/rabbitmqgolang/internal/app/consumers"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/consumers/consumers"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/consumers/models"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/services/implementations"
	"github.com/Filiphasan/rabbitmqgolang/internal/app/services/interfaces"
	"github.com/Filiphasan/rabbitmqgolang/internal/logger"
	"github.com/Filiphasan/rabbitmqgolang/internal/rabbitmq"
	"github.com/Filiphasan/rabbitmqgolang/pkg/constants"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// Start the server

	appConfig := configs.LoadConfig()
	logg := logger.UseLogger(appConfig)
	defer logg.Sync()

	router := gin.Default()

	myService := implementations.NewMyService(logg)

	connectionManager := rabbitmq.NewConnectionManager(appConfig, logg)

	var consumerList []baseConsumers.IBaseConsumer[interface{}]
	consumerList = append(consumerList, consumers.NewBasicSendConsumer(logg, connectionManager, myService))
	consumerList = append(consumerList, consumers.NewBasicPublishFirstConsumer(logg, connectionManager, myService))
	consumerList = append(consumerList, consumers.NewBasicPublishSecondConsumer(logg, connectionManager, myService))

	for _, consumer := range consumerList {
		go func(c baseConsumers.IBaseConsumer[interface{}]) {
			err := c.StartConsuming()
			if err != nil {
				logg.Error("Failed to start consuming", zap.Error(err))
			}
		}(consumer)
	}

	mqService := implementations.NewMqService(connectionManager, logg)

	RouteBasicEndpoint(router, mqService)

	err := router.Run(":8080")
	if err != nil {
		logg.Error("Failed to start server", zap.Error(err))
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs

	logg.Info("Received shutdown signal, shutting down gracefully...")

	var wg sync.WaitGroup
	wg.Add(len(consumerList))
	for _, consumer := range consumerList {
		go func(c baseConsumers.IBaseConsumer[interface{}]) {
			defer wg.Done()
			c.Close()
		}(consumer)
	}
	wg.Wait()

	connectionManager.Close()
}

type BasicRequest struct {
	Message string `json:"message"`
}

func RouteBasicEndpoint(router *gin.Engine, mqService *implementations.MqService) {
	router.POST("/api/queue/send", func(c *gin.Context) {
		request := &BasicRequest{}
		err := c.ShouldBindJSON(request)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		queueModel := models.NewQueueBasicMessage(request.Message)
		sendModel, err := interfaces.NewSendMessageModel[models.QueueBasicMessage](
			constants.BasicSendQueueName,
			queueModel,
			"",
		)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		err = mqService.SendMessage(c, sendModel)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
		}

		c.JSON(200, gin.H{"message": "Message sent successfully"})
	})

	router.POST("/api/queue/publish", func(c *gin.Context) {
		request := &BasicRequest{}
		err := c.ShouldBindJSON(request)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		queueModel := models.NewQueueBasicMessage(request.Message)
		publishModel, err := interfaces.NewPublishMessageModel[models.QueueBasicMessage](
			constants.BasicPublishExchangeName,
			constants.BasicPublishRoutingKey,
			queueModel,
			"",
		)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		err = mqService.PublishMessage(c, publishModel)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
		}

		c.JSON(200, gin.H{"message": "Message published successfully"})
	})
}
