package interfaces

import "github.com/Filiphasan/rabbitmqgolang/internal/app/consumers/models"

type IMyService interface {
	DoBasicSendConsumerWork(message *models.QueueBasicMessage) error
	DoBasicPublishFirstConsumerWork(message *models.QueueBasicMessage) error
	DoBasicPublishSecondConsumerWork(message *models.QueueBasicMessage) error
}
