package metrics

import "github.com/prometheus/client_golang/prometheus"

// MessageMetrics holds counters and histograms related to message operations.
type MessageMetrics struct {
	// TotalMessagesPublished counts the total number of messages published to NATS.
	TotalMessagesPublished prometheus.Counter
	// TotalMessagesReceived counts the total number of messages received from NATS subscriptions.
	TotalMessagesReceived prometheus.Counter
	// FailedPublishAttempts counts the total number of failed attempts to publish messages to NATS.
	FailedPublishAttempts prometheus.Counter
	// MessageProcessingDuration measures the duration of processing messages.
	MessageProcessingDuration prometheus.Histogram
	// MessagePublishLatency measures the latency involved in publishing messages.
	MessagePublishLatency prometheus.Histogram
}

// SubscriptionMetrics holds gauges related to subscriptions.
type SubscriptionMetrics struct {
	// ActiveSubscriptions indicates the number of active NATS subscriptions.
	ActiveSubscriptions prometheus.Gauge
}

// ConnectionMetrics holds metrics related to the NATS connection.
type ConnectionMetrics struct {
	// NATSConnectionStatus indicates the current status of the NATS connection (1 for active, 0 for inactive).
	NATSConnectionStatus prometheus.Gauge
	// NATSConnectionAttempts counts the total number of attempts to connect to the NATS server.
	NATSConnectionAttempts prometheus.Counter
	// NATSConnectionFailures counts the total number of failed connection attempts.
	NATSConnectionFailures prometheus.Counter
}

// Metrics aggregates all metric groups and the Prometheus registry.
type Metrics struct {
	// Message holds all message related metrics.
	Message MessageMetrics
	// Subscription holds all subscription related metrics.
	Subscription SubscriptionMetrics
	// Connection holds all connection related metrics.
	Connection ConnectionMetrics
	// Registry is the Prometheus registry that holds all the metrics.
	Registry *prometheus.Registry
}

// NewMetrics creates all metrics, registers them with a custom Prometheus registry.
// This function automatically registers the metrics to ensure they are exposed via the registry.
func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()

	metrics := &Metrics{
		Message: MessageMetrics{
			TotalMessagesPublished: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "nats_messages_published_total",
				Help: "The total number of messages published to NATS",
			}),
			TotalMessagesReceived: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "nats_messages_received_total",
				Help: "The total number of messages received from NATS subscriptions",
			}),
			FailedPublishAttempts: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "nats_publish_failures_total",
				Help: "The total number of failed attempts to publish messages to NATS",
			}),
			MessageProcessingDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
				Name:    "nats_message_processing_duration_seconds",
				Help:    "Histogram for the duration of message processing",
				Buckets: prometheus.DefBuckets,
			}),
			MessagePublishLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
				Name:    "nats_message_publish_latency_seconds",
				Help:    "Histogram for the latency of publishing messages to NATS",
				Buckets: prometheus.DefBuckets,
			}),
		},
		Subscription: SubscriptionMetrics{
			ActiveSubscriptions: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "nats_active_subscriptions",
				Help: "The number of active NATS subscriptions",
			}),
		},
		Connection: ConnectionMetrics{
			NATSConnectionStatus: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "nats_connection_status",
				Help: "Current status of the NATS connection (1 for active, 0 for inactive)",
			}),
			NATSConnectionAttempts: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "nats_connection_attempts_total",
				Help: "Total number of attempts to connect to the NATS server",
			}),
			NATSConnectionFailures: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "nats_connection_failures_total",
				Help: "Total number of failed attempts to connect to the NATS server",
			}),
		},
		Registry: registry,
	}

	metrics.register()
	return metrics
}

// register registers all the defined metrics with the Prometheus registry.
func (m *Metrics) register() {
	m.Registry.MustRegister(
		m.Message.TotalMessagesPublished,
		m.Message.TotalMessagesReceived,
		m.Message.FailedPublishAttempts,
		m.Message.MessageProcessingDuration,
		m.Message.MessagePublishLatency,
		m.Subscription.ActiveSubscriptions,
		m.Connection.NATSConnectionStatus,
		m.Connection.NATSConnectionAttempts,
		m.Connection.NATSConnectionFailures,
	)
}
