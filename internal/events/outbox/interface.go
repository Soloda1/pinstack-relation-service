package outbox

import (
	"context"
	"pinstack-relation-service/internal/model"
	"time"
)

//go:generate mockery --name=OutboxRepository --output=../../../mocks --outpkg=mocks --case=underscore --with-expecter
type OutboxRepository interface {
	AddEvent(ctx context.Context, outbox model.OutboxEvent) error
	GetEventsForProcessing(ctx context.Context) ([]model.OutboxEvent, error)
	UpdateEventStatus(ctx context.Context, eventID int64, status model.OutboxStatus, sentAt *time.Time) error
	MarkEventAsPending(ctx context.Context, eventID int64) error
}
