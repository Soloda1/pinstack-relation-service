package kafka

import (
	"context"
	model "pinstack-relation-service/internal/domain/models"
)

//go:generate mockery --name=KafkaProducer --output=../../../mocks --outpkg=mocks --case=underscore --with-expecter --filename=mock_producer.go --dir=.
type KafkaProducer interface {
	SendMessage(ctx context.Context, event model.OutboxEvent) <-chan SendResult
	Close()
}

// SendResult represents the delivery result for an outbox event
type SendResult struct {
	EventID int64
	Error   error
}
