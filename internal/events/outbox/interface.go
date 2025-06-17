package outbox

import (
	"context"
	"pinstack-relation-service/internal/model"
)

//go:generate mockery --name=OutboxRepository --output=../../../mocks --outpkg=mocks --case=underscore --with-expecter
type OutboxRepository interface {
	AddEvent(ctx context.Context, outbox model.OutboxEvent) error
}
