package repository

import "context"

type FollowRepository interface {
	Create(ctx context.Context, followerID, followeeID int64) error
	Delete(ctx context.Context, followerID, followeeID int64) error
	GetFollowers(ctx context.Context, followeeID int64) ([]int64, error)
	GetFollowees(ctx context.Context, followerID int64) ([]int64, error)
}
