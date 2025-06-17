package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
)

type Config struct {
	Env         string
	GRPCServer  GRPCServer
	Database    Database
	UserService UserService
	EventTypes  EventTypes
	Kafka       Kafka
}

type GRPCServer struct {
	Address string
	Port    int
}

type EventTypes struct {
	FollowCreated string
	FollowDeleted string
}

type Database struct {
	Username       string
	Password       string
	Host           string
	Port           string
	DbName         string
	MigrationsPath string
}

type UserService struct {
	Address string
	Port    int
}

type Kafka struct {
	Brokers                   string
	Topic                     string
	Acks                      string
	Retries                   int
	RetryBackoffMs            int
	DeliveryTimeoutMs         int
	QueueBufferingMaxMessages int
	QueueBufferingMaxMs       int
	CompressionType           string
	BatchSize                 int
	LingerMs                  int
}

func MustLoad() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	viper.SetDefault("env", "dev")

	viper.SetDefault("grpc_server.address", "0.0.0.0")
	viper.SetDefault("grpc_server.port", 50054)

	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "admin")
	viper.SetDefault("database.host", "relation-db")
	viper.SetDefault("database.port", "5435")
	viper.SetDefault("database.db_name", "relationservice")
	viper.SetDefault("database.migrations_path", "migrations")

	viper.SetDefault("event_types.follow_created", "follow_created")
	viper.SetDefault("event_types.follow_deleted", "follow_deleted")

	viper.SetDefault("user_service.address", "user-service")
	viper.SetDefault("user_service.port", 50051)

	viper.SetDefault("kafka.brokers", "kafka1:9092,kafka2:9092,kafka3:9092")
	viper.SetDefault("kafka.topic", "relation_events")
	viper.SetDefault("kafka.acks", "all")
	viper.SetDefault("kafka.retries", 3)
	viper.SetDefault("kafka.retry_backoff_ms", 500)
	viper.SetDefault("kafka.delivery_timeout_ms", 5000)
	viper.SetDefault("kafka.queue_buffering_max_messages", 100000)
	viper.SetDefault("kafka.queue_buffering_max_ms", 5)
	viper.SetDefault("kafka.compression_type", "snappy")
	viper.SetDefault("kafka.batch_size", 16384)
	viper.SetDefault("kafka.linger_ms", 5)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %s", err)
		os.Exit(1)
	}

	config := &Config{
		Env: viper.GetString("env"),
		GRPCServer: GRPCServer{
			Address: viper.GetString("grpc_server.address"),
			Port:    viper.GetInt("grpc_server.port"),
		},
		Database: Database{
			Username:       viper.GetString("database.username"),
			Password:       viper.GetString("database.password"),
			Host:           viper.GetString("database.host"),
			Port:           viper.GetString("database.port"),
			DbName:         viper.GetString("database.db_name"),
			MigrationsPath: viper.GetString("database.migrations_path"),
		},
		UserService: UserService{
			Address: viper.GetString("user_service.address"),
			Port:    viper.GetInt("user_service.port"),
		},
		EventTypes: EventTypes{
			FollowCreated: viper.GetString("event_types.follow_created"),
			FollowDeleted: viper.GetString("event_types.follow_deleted"),
		},
		Kafka: Kafka{
			Brokers:                   viper.GetString("kafka.brokers"),
			Topic:                     viper.GetString("kafka.topic"),
			Acks:                      viper.GetString("kafka.acks"),
			Retries:                   viper.GetInt("kafka.retries"),
			RetryBackoffMs:            viper.GetInt("kafka.retry_backoff_ms"),
			DeliveryTimeoutMs:         viper.GetInt("kafka.delivery_timeout_ms"),
			QueueBufferingMaxMessages: viper.GetInt("kafka.queue_buffering_max_messages"),
			QueueBufferingMaxMs:       viper.GetInt("kafka.queue_buffering_max_ms"),
			CompressionType:           viper.GetString("kafka.compression_type"),
			BatchSize:                 viper.GetInt("kafka.batch_size"),
			LingerMs:                  viper.GetInt("kafka.linger_ms"),
		},
	}

	return config
}
