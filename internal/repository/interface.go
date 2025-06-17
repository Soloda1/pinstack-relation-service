package repository

import (
	"context"
	"pinstack-relation-service/internal/model"
)

//go:generate mockery --name=FollowRepository --output=../../mocks --outpkg=mocks --case=underscore --with-expecter
type FollowRepository interface {
	Create(ctx context.Context, followerID, followeeID int64) (model.Follower, error)
	Delete(ctx context.Context, followerID, followeeID int64) error
	Exists(ctx context.Context, followerID, followeeID int64) (bool, error)
	GetFollowers(ctx context.Context, followeeID int64, limit, offset int32) ([]int64, error)
	GetFollowees(ctx context.Context, followerID int64, limit, offset int32) ([]int64, error)
}
