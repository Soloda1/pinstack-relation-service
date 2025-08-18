package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// gRPC metrics
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relation_service_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "relation_service_grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status"},
	)

	// Database metrics
	databaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relation_service_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"query_type", "status"},
	)

	databaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "relation_service_database_query_duration_seconds",
			Help:    "Duration of database queries",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query_type"},
	)

	// Relation-specific metrics
	relationOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relation_service_operations_total",
			Help: "Total number of relation operations",
		},
		[]string{"operation", "status"},
	)

	// Kafka metrics
	kafkaMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relation_service_kafka_messages_total",
			Help: "Total number of Kafka messages",
		},
		[]string{"topic", "operation", "status"},
	)

	kafkaMessageDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "relation_service_kafka_message_duration_seconds",
			Help:    "Duration of Kafka message operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic", "operation"},
	)

	// Outbox metrics
	outboxOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relation_service_outbox_operations_total",
			Help: "Total number of outbox operations",
		},
		[]string{"operation", "status"},
	)

	// System metrics
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "relation_service_active_connections",
			Help: "Number of active connections",
		},
	)

	serviceHealth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "relation_service_health",
			Help: "Service health status (1 = healthy, 0 = unhealthy)",
		},
	)
)
