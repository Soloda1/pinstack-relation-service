package model

import "time"

type Follower struct {
	ID         int64     `json:"id"`
	FollowerID int64     `json:"follower_id"`
	FolloweeID int64     `json:"followee_id"`
	CreatedAt  time.Time `json:"created_at"`
}
