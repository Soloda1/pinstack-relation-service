package repository_postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/internal/repository"
)

type UnitOfWork interface {
	Begin(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	FollowRepository() repository.FollowRepository
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type PostgresUnitOfWork struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

func NewPostgresUOW(pool *pgxpool.Pool, log *logger.Logger) UnitOfWork {
	return &PostgresUnitOfWork{pool: pool, log: log}
}

func (uow *PostgresUnitOfWork) Begin(ctx context.Context) (Transaction, error) {
	tx, err := uow.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}
	return &PostgresTransaction{tx: tx, log: uow.log}, nil
}

type PostgresTransaction struct {
	tx  pgx.Tx
	log *logger.Logger
}

func (t *PostgresTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *PostgresTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func (t *PostgresTransaction) FollowRepository() repository.FollowRepository {
	return NewFollowRepository(t.tx, t.log)
}
