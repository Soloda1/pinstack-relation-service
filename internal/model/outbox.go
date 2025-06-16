package model

import (
	"encoding/json"
	"time"
)

type OutboxEvent struct {
	ID          int64           `json:"id"`
	AggregateID int64           `json:"aggregate_id"`
	EventType   string          `json:"event_type"`
	Payload     json.RawMessage `json:"payload"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	SentAt      *time.Time      `json:"sent_at"`
}
