package repository

type FollowRepository interface {
	Create(ctx, followerID, followeeID int64) error
	Delete(ctx, followerID, followeeID int64) error
	GetFollowers(ctx, followeeID int64) ([]int64, error)
	GetFollowees(ctx, followerID int64) ([]int64, error)
}
