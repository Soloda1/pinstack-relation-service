package outbox

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"pinstack-relation-service/config"
	"pinstack-relation-service/internal/events/kafka"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/internal/model"
	"pinstack-relation-service/internal/utils"
)

type OutboxWorker struct {
	repo      OutboxRepository
	producer  kafka.KafkaProducer
	log       *logger.Logger
	config    config.OutboxConfig
	wg        *sync.WaitGroup
	stopChan  chan struct{}
	ticker    *time.Ticker
	semaphore *utils.Semaphore
}

func NewOutboxWorker(
	repo OutboxRepository,
	producer kafka.KafkaProducer,
	config config.OutboxConfig,
	log *logger.Logger,
) *OutboxWorker {
	return &OutboxWorker{
		repo:      repo,
		producer:  producer,
		config:    config,
		log:       log,
		wg:        &sync.WaitGroup{},
		stopChan:  make(chan struct{}),
		ticker:    time.NewTicker(config.TickInterval()),
		semaphore: utils.NewSemaphore(config.Concurrency),
	}
}

func (wp *OutboxWorker) Start(ctx context.Context) {
	wp.log.Info("Starting outbox worker pool",
		slog.Int("concurrency", wp.config.Concurrency),
		slog.Int("batch_size", wp.config.BatchSize),
		slog.Int("tick_interval_ms", wp.config.TickIntervalMs))

	go func() {
		for {
			select {
			case <-wp.ticker.C:
				wp.processBatch(ctx)
			case <-wp.stopChan:
				wp.log.Info("Worker pool stopping due to stop signal")
				return
			case <-ctx.Done():
				wp.log.Info("Worker pool stopping due to context cancellation")
				return
			}
		}
	}()
}

func (wp *OutboxWorker) Stop() {
	wp.log.Info("Stopping outbox worker pool")
	wp.ticker.Stop()
	close(wp.stopChan)
	wp.wg.Wait()
	wp.log.Info("Outbox worker pool stopped")
}

func (wp *OutboxWorker) processBatch(ctx context.Context) {
	wp.log.Debug("Processing outbox batch", slog.Int("batch_size", wp.config.BatchSize))

	events, err := wp.repo.GetEventsForProcessing(ctx, wp.config.BatchSize)
	if err != nil {
		wp.log.Error("Failed to get events for processing", slog.String("error", err.Error()))
		return
	}

	if len(events) == 0 {
		wp.log.Debug("No events to process")
		return
	}

	wp.log.Info("Found events to process", slog.Int("count", len(events)))

	for _, event := range events {
		wp.wg.Add(1)
		go wp.worker(ctx, event)
	}
}

func (wp *OutboxWorker) worker(ctx context.Context, event model.OutboxEvent) {
	defer wp.wg.Done()

	wp.semaphore.Acquire()
	defer wp.semaphore.Release()

	select {
	case <-ctx.Done():
		wp.log.Debug("Skipping event processing due to context cancellation",
			slog.Int64("event_id", event.ID))
		return
	default:
		wp.processEvent(ctx, event)
	}
}

func (wp *OutboxWorker) processEvent(ctx context.Context, event model.OutboxEvent) {
	if err := wp.repo.MarkEventAsPending(ctx, event.ID); err != nil {
		wp.log.Error("Failed to mark event as pending",
			slog.Int64("event_id", event.ID),
			slog.String("error", err.Error()))
		return
	}

	resultChan := wp.producer.SendMessage(ctx, event)
	result := <-resultChan

	if result.Error != nil {
		wp.log.Error("Failed to send event to Kafka",
			slog.Int64("event_id", event.ID),
			slog.String("error", result.Error.Error()))

		if err := wp.repo.UpdateEventStatus(ctx, event.ID, model.OutboxStatusError, nil); err != nil {
			wp.log.Error("Failed to update event status to error",
				slog.Int64("event_id", event.ID),
				slog.String("error", err.Error()))
		}
		return
	}

	now := time.Now()
	if err := wp.repo.UpdateEventStatus(ctx, event.ID, model.OutboxStatusSent, &now); err != nil {
		wp.log.Error("Failed to update event status to sent",
			slog.Int64("event_id", event.ID),
			slog.String("error", err.Error()))
		return
	}

	wp.log.Info("Event successfully processed and sent", slog.Int64("event_id", event.ID))
}
