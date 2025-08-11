package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	model "pinstack-relation-service/internal/domain/models"
	ports "pinstack-relation-service/internal/domain/ports/output"
	kafka_port "pinstack-relation-service/internal/domain/ports/output/kafka"
	"pinstack-relation-service/internal/infrastructure/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
)

type Producer struct {
	producer *kafka.Producer
	topic    string
	logger   ports.Logger
}

func NewProducer(kafkaConfig config.Kafka, logger ports.Logger) (*Producer, error) {
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

func (p *Producer) SendMessage(ctx context.Context, event model.OutboxEvent) <-chan kafka_port.SendResult {
	resultChan := make(chan kafka_port.SendResult)

	go func() {
		defer close(resultChan)

		payload, err := json.Marshal(event.Payload)
		if err != nil {
			p.logger.Error("Failed to marshal event payload", slog.String("error", err.Error()), slog.Int64("event_id", event.ID))
			resultChan <- kafka_port.SendResult{EventID: event.ID, Error: err}
			return
		}

		deliveryChan := make(chan kafka.Event)
		defer close(deliveryChan)

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

		err = p.producer.Produce(message, deliveryChan)
		if err != nil {
			p.logger.Error("Failed to produce message", slog.String("error", err.Error()), slog.Int64("event_id", event.ID))
			resultChan <- kafka_port.SendResult{EventID: event.ID, Error: err}
			return
		}

		select {
		case <-ctx.Done():
			resultChan <- kafka_port.SendResult{EventID: event.ID, Error: ctx.Err()}
		case e := <-deliveryChan:
			m, ok := e.(*kafka.Message)
			if !ok {
				p.logger.Error("Unexpected event type received on delivery channel",
					slog.String("event_type", fmt.Sprintf("%T", e)),
					slog.Int64("event_id", event.ID))
				resultChan <- kafka_port.SendResult{EventID: event.ID, Error: custom_errors.ErrUnexpectedEventType}
				return
			}
			if m.TopicPartition.Error != nil {
				p.logger.Error("Message delivery failed",
					slog.String("error", m.TopicPartition.Error.Error()),
					slog.Int64("event_id", event.ID))
				resultChan <- kafka_port.SendResult{EventID: event.ID, Error: m.TopicPartition.Error}
			} else {
				p.logger.Info("Message delivered successfully",
					slog.Int64("event_id", event.ID),
					slog.String("topic", *m.TopicPartition.Topic),
					slog.Int("partition", int(m.TopicPartition.Partition)),
					slog.Int("offset", int(m.TopicPartition.Offset)))
				resultChan <- kafka_port.SendResult{EventID: event.ID, Error: nil}
			}
		}
	}()

	return resultChan
}

func (p *Producer) Close() {
	remainingMessages := p.producer.Flush(10000) // Таймаут в мс
	if remainingMessages > 0 {
		p.logger.Warn("Producer closed with pending messages", slog.Int("count", remainingMessages))
	}

	p.producer.Close()
	p.logger.Info("Kafka producer closed")
}
