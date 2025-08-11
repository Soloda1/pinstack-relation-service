package outbox

import (
	"context"
	"log/slog"
	model "pinstack-relation-service/internal/domain/models"
	ports "pinstack-relation-service/internal/domain/ports/output"
	repository_postgres "pinstack-relation-service/internal/infrastructure/outbound/repository/postgres"
	"time"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	log ports.Logger
	db  repository_postgres.PgDB
}

func NewOutboxRepository(db repository_postgres.PgDB, log ports.Logger) *Repository {
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
		r.log.Error("Failed to add event to outbox", slog.String("error", err.Error()), slog.Int64("aggregate_id", outbox.AggregateID), slog.String("event_type", string(outbox.EventType)))
		return err
	}

	r.log.Info("Event added to outbox successfully", slog.Int64("aggregate_id", outbox.AggregateID), slog.String("event_type", string(outbox.EventType)))
	return nil
}

func (r *Repository) GetEventsForProcessing(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	query := `
		SELECT id, aggregate_id, event_type, payload, status, created_at, sent_at
		FROM outbox
		WHERE status = 'new'
		ORDER BY created_at
		LIMIT @limit
	`
	args := pgx.NamedArgs{
		"limit": limit,
	}

	rows, err := r.db.Query(ctx, query, args)
	if err != nil {
		r.log.Error("Failed to get events for processing", slog.String("error", err.Error()))
		return nil, err
	}
	defer rows.Close()

	var events []model.OutboxEvent
	for rows.Next() {
		var event model.OutboxEvent
		if err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.EventType,
			&event.Payload,
			&event.Status,
			&event.CreatedAt,
			&event.SentAt,
		); err != nil {
			r.log.Error("Failed to scan event row", slog.String("error", err.Error()))
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Error iterating over event rows", slog.String("error", err.Error()))
		return nil, err
	}

	return events, nil
}

func (r *Repository) UpdateEventStatus(ctx context.Context, eventID int64, status model.OutboxStatus, sentAt *time.Time) error {
	var query string
	var args pgx.NamedArgs

	if sentAt != nil {
		query = `
			UPDATE outbox
			SET status = @status, sent_at = @sent_at
			WHERE id = @id
		`
		args = pgx.NamedArgs{
			"status":  status,
			"sent_at": *sentAt,
			"id":      eventID,
		}
	} else {
		query = `
			UPDATE outbox
			SET status = @status
			WHERE id = @id
		`
		args = pgx.NamedArgs{
			"status": status,
			"id":     eventID,
		}
	}

	_, err := r.db.Exec(ctx, query, args)
	if err != nil {
		r.log.Error("Failed to update event status",
			slog.String("error", err.Error()),
			slog.Int64("event_id", eventID),
			slog.String("status", string(status)))
		return err
	}

	r.log.Info("Event status updated",
		slog.Int64("event_id", eventID),
		slog.String("status", string(status)))
	return nil
}

func (r *Repository) MarkEventAsPending(ctx context.Context, eventID int64) error {
	query := `
		UPDATE outbox
		SET status = @status
		WHERE id = @id
	`

	args := pgx.NamedArgs{
		"status": model.OutboxStatusPending,
		"id":     eventID,
	}

	_, err := r.db.Exec(ctx, query, args)
	if err != nil {
		r.log.Error("Failed to mark event as pending",
			slog.String("error", err.Error()),
			slog.Int64("event_id", eventID))
		return err
	}

	r.log.Debug("Event marked as pending", slog.Int64("event_id", eventID))
	return nil
}
