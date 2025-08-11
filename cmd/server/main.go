package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"pinstack-relation-service/internal/application/service"
	"pinstack-relation-service/internal/infrastructure/config"
	follow_grpc "pinstack-relation-service/internal/infrastructure/inbound/grpc"
	infra_logger "pinstack-relation-service/internal/infrastructure/logger"
	user_adapter "pinstack-relation-service/internal/infrastructure/outbound/client/user"
	kafka_adapter "pinstack-relation-service/internal/infrastructure/outbound/events/kafka"
	outbox_adapter "pinstack-relation-service/internal/infrastructure/outbound/outbox"
	repository_postgres "pinstack-relation-service/internal/infrastructure/outbound/repository/postgres"
	uow_adapter "pinstack-relation-service/internal/infrastructure/outbound/uow"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	log := infra_logger.New(cfg.Env)

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

	kafkaProducer, err := kafka_adapter.NewProducer(cfg.Kafka, log)
	if err != nil {
		log.Error("Failed to initialize Kafka producer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer kafkaProducer.Close()

	outboxRepo := outbox_adapter.NewOutboxRepository(pool, log)

	outboxWorker := outbox_adapter.NewOutboxWorker(
		outboxRepo,
		kafkaProducer,
		cfg.Outbox,
		log,
	)

	outboxWorker.Start(ctx)
	defer outboxWorker.Stop()

	unitOfWork := uow_adapter.NewPostgresUOW(pool, log)
	followRepo := repository_postgres.NewFollowRepository(pool, log)

	userServiceConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.UserService.Address, cfg.UserService.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("Failed to connect to user service", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func(userServiceConn *grpc.ClientConn) {
		err := userServiceConn.Close()
		if err != nil {
			log.Error("Failed to close user service connection", slog.String("error", err.Error()))
		}
	}(userServiceConn)

	userClient := user_adapter.NewUserClient(userServiceConn, log)

	followService := service.NewFollowService(log, followRepo, unitOfWork, userClient)
	followGRPCApi := follow_grpc.NewFollowGRPCService(followService, log)
	grpcServer := follow_grpc.NewServer(followGRPCApi, cfg.GRPCServer.Address, cfg.GRPCServer.Port, log)

	metricsAddr := fmt.Sprintf("%s:%d", cfg.Prometheus.Address, cfg.Prometheus.Port)
	metricsServer := &http.Server{
		Addr:    metricsAddr,
		Handler: nil,
	}

	done := make(chan bool, 1)
	metricsDone := make(chan bool, 1)

	go func() {
		if err := grpcServer.Run(); err != nil {
			log.Error("gRPC server error", slog.String("error", err.Error()))
		}
		done <- true
	}()

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Info("Starting Prometheus metrics server", slog.String("address", metricsAddr))
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Prometheus metrics server error", slog.String("error", err.Error()))
		}
		metricsDone <- true
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Info("Shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := grpcServer.Shutdown(); err != nil {
		log.Error("gRPC server shutdown error", slog.String("error", err.Error()))
	}

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Error("Metrics server shutdown error", slog.String("error", err.Error()))
	}

	<-done
	<-metricsDone

	log.Info("Server exited")
}
