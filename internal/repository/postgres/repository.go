package repository_postgres

import "pinstack-relation-service/internal/logger"

type Repository struct {
	log *logger.Logger
	db  PgDB
}

func NewFollowRepository(db PgDB, log *logger.Logger) *Repository {
	return &Repository{db: db, log: log}
}

func (r Repository) Create(ctx, followerID, followeeID int64) error {
	//TODO implement me
	panic("implement me")
}

func (r Repository) Delete(ctx, followerID, followeeID int64) error {
	//TODO implement me
	panic("implement me")
}

func (r Repository) GetFollowers(ctx, followeeID int64) ([]int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r Repository) GetFollowees(ctx, followerID int64) ([]int64, error) {
	//TODO implement me
	panic("implement me")
}
