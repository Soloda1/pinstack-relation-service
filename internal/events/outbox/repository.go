package outbox

import (
	"pinstack-relation-service/internal/logger"
	repository_postgres "pinstack-relation-service/internal/repository/postgres"
)

type Repository struct {
	log *logger.Logger
	db  repository_postgres.PgDB
}

func NewOutboxRepository(db repository_postgres.PgDB, log *logger.Logger) *Repository {
	return &Repository{db: db, log: log}
}
