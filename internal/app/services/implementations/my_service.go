package implementations

import (
	"github.com/Filiphasan/rabbitmqgolang/internal/app/consumers/models"
	"go.uber.org/zap"
	"time"
)

type MyService struct {
	logger *zap.Logger
}

func NewMyService(logger *zap.Logger) *MyService {
	return &MyService{
		logger: logger,
	}
}

func (s *MyService) DoBasicSendConsumerWork(message *models.QueueBasicMessage) error {
	s.logger.Info("Processing message started", zap.String("message", message.Message))
	// Simulate some processing work
	time.Sleep(1 * time.Second)
	s.logger.Info("Processing message completed", zap.String("message", message.Message))
	return nil
}

func (s *MyService) DoBasicPublishFirstConsumerWork(message *models.QueueBasicMessage) error {
	s.logger.Info("Processing message in First Consumer started", zap.String("message", message.Message))
	// Simulate some processing work
	time.Sleep(1 * time.Second)
	s.logger.Info("Processing message in First Consumer completed", zap.String("message", message.Message))
	return nil
}

func (s *MyService) DoBasicPublishSecondConsumerWork(message *models.QueueBasicMessage) error {
	s.logger.Info("Processing message in Second Consumer started", zap.String("message", message.Message))
	// Simulate some processing work
	time.Sleep(1 * time.Second)
	s.logger.Info("Processing message in Second Consumer completed", zap.String("message", message.Message))
	return nil
}
