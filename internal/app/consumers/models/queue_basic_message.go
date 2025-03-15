package models

type QueueBasicMessage struct {
	Message string `json:"message"`
}

func NewQueueBasicMessage(message string) *QueueBasicMessage {
	return &QueueBasicMessage{
		Message: message,
	}
}
