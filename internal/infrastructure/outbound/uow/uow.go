package uow

import (
	"context"
	"fmt"
	ports "pinstack-relation-service/internal/domain/ports/output"
	outbox_port "pinstack-relation-service/internal/domain/ports/output/outbox"
	repository_port "pinstack-relation-service/internal/domain/ports/output/repository"
	uow_port "pinstack-relation-service/internal/domain/ports/output/uow"
	outbox_postgres "pinstack-relation-service/internal/infrastructure/outbound/outbox"
	repository_postgres "pinstack-relation-service/internal/infrastructure/outbound/repository/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUnitOfWork struct {
	pool *pgxpool.Pool
	log  ports.Logger
}

func NewPostgresUOW(pool *pgxpool.Pool, log ports.Logger) uow_port.UnitOfWork {
	return &PostgresUnitOfWork{pool: pool, log: log}
}

func (uow *PostgresUnitOfWork) Begin(ctx context.Context) (uow_port.Transaction, error) {
	tx, err := uow.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}
	return &PostgresTransaction{tx: tx, log: uow.log}, nil
}

type PostgresTransaction struct {
	tx  pgx.Tx
	log ports.Logger
}

func (t *PostgresTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *PostgresTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func (t *PostgresTransaction) FollowRepository() repository_port.FollowRepository {
	return repository_postgres.NewFollowRepository(t.tx, t.log)
}

func (t *PostgresTransaction) OutboxRepository() outbox_port.OutboxRepository {
	return outbox_postgres.NewOutboxRepository(t.tx, t.log)
}
