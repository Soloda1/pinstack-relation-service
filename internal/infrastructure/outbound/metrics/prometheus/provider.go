package prometheus

import (
	"time"

	"pinstack-relation-service/internal/domain/ports/output"
)

type PrometheusMetricsProvider struct{}

func NewPrometheusMetricsProvider() output.MetricsProvider {
	return &PrometheusMetricsProvider{}
}

func (p *PrometheusMetricsProvider) IncrementGRPCRequests(method, status string) {
	grpcRequestsTotal.WithLabelValues(method, status).Inc()
}

func (p *PrometheusMetricsProvider) RecordGRPCRequestDuration(method, status string, duration time.Duration) {
	grpcRequestDuration.WithLabelValues(method, status).Observe(duration.Seconds())
}

func (p *PrometheusMetricsProvider) IncrementDatabaseQueries(queryType string, success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	databaseQueriesTotal.WithLabelValues(queryType, status).Inc()
}

func (p *PrometheusMetricsProvider) RecordDatabaseQueryDuration(queryType string, duration time.Duration) {
	databaseQueryDuration.WithLabelValues(queryType).Observe(duration.Seconds())
}

func (p *PrometheusMetricsProvider) IncrementRelationOperations(operation string, success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	relationOperationsTotal.WithLabelValues(operation, status).Inc()
}

func (p *PrometheusMetricsProvider) IncrementKafkaMessages(topic, operation string, success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	kafkaMessagesTotal.WithLabelValues(topic, operation, status).Inc()
}

func (p *PrometheusMetricsProvider) RecordKafkaMessageDuration(topic, operation string, duration time.Duration) {
	kafkaMessageDuration.WithLabelValues(topic, operation).Observe(duration.Seconds())
}

func (p *PrometheusMetricsProvider) IncrementOutboxOperations(operation string, success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	outboxOperationsTotal.WithLabelValues(operation, status).Inc()
}

func (p *PrometheusMetricsProvider) SetActiveConnections(count int) {
	activeConnections.Set(float64(count))
}

func (p *PrometheusMetricsProvider) SetServiceHealth(healthy bool) {
	if healthy {
		serviceHealth.Set(1)
	} else {
		serviceHealth.Set(0)
	}
}
