package service

import "context"

type FollowService interface {
	Follow(ctx context.Context, followerID, followeeID int64) error
	Unfollow(ctx context.Context, followerID, followeeID int64) error
	GetFollowers(ctx context.Context, followeeID int64) ([]int64, error)
	GetFollowees(ctx context.Context, followerID int64) ([]int64, error)
}
