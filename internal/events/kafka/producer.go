package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log/slog"
	"pinstack-relation-service/config"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/internal/model"
)

type Producer struct {
	producer *kafka.Producer
	topic    string
	logger   *logger.Logger
}

func NewProducer(kafkaConfig config.Kafka, logger *logger.Logger) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaConfig.Brokers,
		// Настройки надежности доставки
		"acks":                kafkaConfig.Acks,
		"retries":             kafkaConfig.Retries,
		"retry.backoff.ms":    kafkaConfig.RetryBackoffMs,
		"delivery.timeout.ms": kafkaConfig.DeliveryTimeoutMs,
		// Дополнительные настройки производительности
		"queue.buffering.max.messages": kafkaConfig.QueueBufferingMaxMessages,
		"queue.buffering.max.ms":       kafkaConfig.QueueBufferingMaxMs,
		"compression.type":             kafkaConfig.CompressionType,
		"batch.size":                   kafkaConfig.BatchSize,
		"linger.ms":                    kafkaConfig.LingerMs,
	})

	if err != nil {
		logger.Error("Failed to create Kafka producer", slog.String("error", err.Error()))
		return nil, err
	}

	logger.Info("Kafka producer created successfully", slog.String("brokers", kafkaConfig.Brokers), slog.String("topic", kafkaConfig.Topic))

	return &Producer{
		producer: p,
		topic:    kafkaConfig.Topic,
		logger:   logger,
	}, nil
}

func (p *Producer) SendMessage(ctx context.Context, event model.OutboxEvent) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		p.logger.Error("Failed to marshal event payload", slog.String("error", err.Error()), slog.Int64("event_id", event.ID))
		return err
	}

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(event.EventType),
		Value: payload,
		Headers: []kafka.Header{
			{
				Key:   "event_id",
				Value: []byte(fmt.Sprintf("%d", event.ID)),
			},
			{
				Key:   "event_type",
				Value: []byte(event.EventType),
			},
			{
				Key:   "created_at",
				Value: []byte(event.CreatedAt.String()),
			},
		},
	}

	err = p.producer.Produce(message, nil)
	if err != nil {
		p.logger.Error("Failed to produce message", slog.String("error", err.Error()), slog.Int64("event_id", event.ID))
		return err
	}

	go p.handleDeliveryReports()

	p.logger.Info("Message sent to Kafka", slog.Int64("event_id", event.ID), slog.String("event_type", event.EventType))
	return nil
}

func (p *Producer) handleDeliveryReports() {
	for e := range p.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				p.logger.Error("Failed to deliver message",
					slog.String("error", ev.TopicPartition.Error.Error()),
					slog.String("topic", *ev.TopicPartition.Topic),
					slog.Int("partition", int(ev.TopicPartition.Partition)))
			} else {
				p.logger.Debug("Message delivered",
					slog.String("topic", *ev.TopicPartition.Topic),
					slog.Int("partition", int(ev.TopicPartition.Partition)),
					slog.Int("offset", int(ev.TopicPartition.Offset)))
			}
		}
	}
}

func (p *Producer) Close() {
	remainingMessages := p.producer.Flush(10000) // Таймаут в мс
	if remainingMessages > 0 {
		p.logger.Warn("Producer closed with pending messages", slog.Int("count", remainingMessages))
	}

	p.producer.Close()
	p.logger.Info("Kafka producer closed")
}
