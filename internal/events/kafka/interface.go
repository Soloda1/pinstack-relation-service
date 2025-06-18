package kafka

import (
	"context"
	"pinstack-relation-service/internal/model"
)

//go:generate mockery --name=KafkaProducer --output=../../../mocks --outpkg=mocks --case=underscore --with-expecter --filename=mock_producer.go --dir=.
type KafkaProducer interface {
	SendMessage(ctx context.Context, event model.OutboxEvent) <-chan SendResult
	Close()
}
