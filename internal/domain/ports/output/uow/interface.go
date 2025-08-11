package uow

import (
	"context"
	"pinstack-relation-service/internal/domain/ports/output/outbox"
	"pinstack-relation-service/internal/domain/ports/output/repository"
)

//go:generate mockery --name=UnitOfWork --output=../../mocks --outpkg=mocks --case=underscore --with-expecter
type UnitOfWork interface {
	Begin(ctx context.Context) (Transaction, error)
}

//go:generate mockery --name=Transaction --output=../../mocks --outpkg=mocks --case=underscore --with-expecter
type Transaction interface {
	OutboxRepository() outbox.OutboxRepository
	FollowRepository() repository.FollowRepository
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
