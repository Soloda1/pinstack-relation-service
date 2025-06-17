package outbox

import (
	"context"
	"pinstack-relation-service/internal/model"
)

type OutboxRepository interface {
	AddEvent(ctx context.Context, outbox model.OutboxEvent) error
}
