package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"
	"os/signal"
	"pinstack-relation-service/config"
	follow_grpc "pinstack-relation-service/internal/delivery/grpc"
	"pinstack-relation-service/internal/events/kafka"
	"pinstack-relation-service/internal/events/outbox"
	"pinstack-relation-service/internal/logger"
	repository_postgres "pinstack-relation-service/internal/repository/postgres"
	"pinstack-relation-service/internal/service"
	"pinstack-relation-service/internal/uow"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DbName)
	ctx := context.Background()
	log := logger.New(cfg.Env)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Error("Failed to parse postgres poolConfig", slog.String("error", err.Error()))
		os.Exit(1)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Error("Failed to create postgres pool", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	kafkaProducer, err := kafka.NewProducer(cfg.Kafka, log)
	if err != nil {
		log.Error("Failed to initialize Kafka producer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer kafkaProducer.Close()

	outboxRepo := outbox.NewOutboxRepository(pool, log)

	outboxWorker := outbox.NewOutboxWorker(
		outboxRepo,
		kafkaProducer,
		cfg.Outbox,
		log,
	)

	outboxWorker.Start(ctx)
	defer outboxWorker.Stop()

	unitOfWork := uow.NewPostgresUOW(pool, log)
	followRepo := repository_postgres.NewFollowRepository(pool, log)

	followService := service.NewFollowService(log, followRepo, unitOfWork)
	followGRPCApi := follow_grpc.NewFollowGRPCService(followService, log)
	grpcServer := follow_grpc.NewServer(followGRPCApi, cfg.GRPCServer.Address, cfg.GRPCServer.Port, log)

	done := make(chan bool)
	go func() {
		if err := grpcServer.Run(); err != nil {
			log.Error("gRPC server error", slog.String("error", err.Error()))
		}
		done <- true
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down gRPC server...")
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := grpcServer.Shutdown(); err != nil {
		log.Error("gRPC server shutdown error", slog.String("error", err.Error()))
	}
	<-done
	log.Info("Server exiting")
}
