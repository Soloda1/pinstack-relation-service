package model

import (
	"encoding/json"
	"time"
)

type OutboxStatus string

const (
	OutboxStatusNew     OutboxStatus = "new"
	OutboxStatusPending OutboxStatus = "pending"
	OutboxStatusSent    OutboxStatus = "sent"
	OutboxStatusError   OutboxStatus = "error"
)

const (
	EventTypeFollowCreated = "follow_created"
	EventTypeFollowDeleted = "follow_deleted"
)

type FollowCreatedPayload struct {
	FollowerID  int64     `json:"follower_id"`
	FolloweeID  int64     `json:"followee_id"`
	Timestamptz time.Time `json:"timestamptz"`
}

type OutboxEvent struct {
	ID          int64           `json:"id"`
	AggregateID int64           `json:"aggregate_id"`
	EventType   string          `json:"event_type"`
	Payload     json.RawMessage `json:"payload"`
	Status      OutboxStatus    `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	SentAt      *time.Time      `json:"sent_at"`
}
