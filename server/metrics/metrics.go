package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetrics struct {
	authenticatedWebhooks *prometheus.CounterVec
	successfulWebhooks    *prometheus.CounterVec
}

func NewPrometheusMetrics() *PrometheusMetrics {
	metrics := &PrometheusMetrics{
		authenticatedWebhooks: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iot",
			Subsystem: "webhook",
			Name:      "authenticated_total",
			Help:      "The total number of authenticated webhooks",
		},
			[]string{"source"},
		),
		successfulWebhooks: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "iot",
			Subsystem: "webhook",
			Name:      "successful_total",
			Help:      "The total number of successful webhooks",
		},
			[]string{"source"},
		),
	}

	return metrics
}

func (m *PrometheusMetrics) IncAuthenticatedWebhook(source string) {
	m.authenticatedWebhooks.WithLabelValues(source).Inc()
}

func (m *PrometheusMetrics) IncSuccessfulWebhook(source string) {
	m.successfulWebhooks.WithLabelValues(source).Inc()
}
