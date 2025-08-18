package output

import "time"

type MetricsProvider interface {
	IncrementGRPCRequests(method, status string)
	RecordGRPCRequestDuration(method, status string, duration time.Duration)

	IncrementDatabaseQueries(queryType string, success bool)
	RecordDatabaseQueryDuration(queryType string, duration time.Duration)

	IncrementRelationOperations(operation string, success bool)

	IncrementKafkaMessages(topic, operation string, success bool)
	RecordKafkaMessageDuration(topic, operation string, duration time.Duration)

	IncrementOutboxOperations(operation string, success bool)

	SetActiveConnections(count int)
	SetServiceHealth(healthy bool)
}
