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

type OutboxEvent struct {
	ID          int64           `json:"id"`
	AggregateID int64           `json:"aggregate_id"`
	EventType   string          `json:"event_type"`
	Payload     json.RawMessage `json:"payload"`
	Status      OutboxStatus    `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	SentAt      *time.Time      `json:"sent_at"`
}
