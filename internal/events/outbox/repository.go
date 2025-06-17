package outbox

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/internal/model"
	repository_postgres "pinstack-relation-service/internal/repository/postgres"
)

type Repository struct {
	log *logger.Logger
	db  repository_postgres.PgDB
}

func NewOutboxRepository(db repository_postgres.PgDB, log *logger.Logger) *Repository {
	return &Repository{db: db, log: log}
}

func (r *Repository) AddEvent(ctx context.Context, outbox model.OutboxEvent) error {
	args := pgx.NamedArgs{
		"aggregate_id": outbox.AggregateID,
		"event_type":   outbox.EventType,
		"payload":      outbox.Payload,
	}

	query := `INSERT INTO outbox (aggregate_id, event_type, payload) VALUES (@aggregate_id, @event_type, @payload)`

	_, err := r.db.Exec(ctx, query, args)
	if err != nil {
		r.log.Error("Failed to add event to outbox", slog.String("error", err.Error()), slog.Int64("aggregate_id", outbox.AggregateID), slog.String("event_type", outbox.EventType))
		return err
	}

	r.log.Info("Event added to outbox successfully", slog.Int64("aggregate_id", outbox.AggregateID), slog.String("event_type", outbox.EventType))
	return nil
}
