package service

import (
	"context"
	"pinstack-relation-service/internal/model"
)

//go:generate mockery --name=FollowService --output=../../mocks --outpkg=mocks --case=underscore --with-expecter
type FollowService interface {
	Follow(ctx context.Context, followerID, followeeID int64) error
	Unfollow(ctx context.Context, followerID, followeeID int64) error
	GetFollowers(ctx context.Context, followeeID int64, limit, page int32) ([]*model.User, int64, error)
	GetFollowees(ctx context.Context, followerID int64, limit, page int32) ([]*model.User, int64, error)
}
