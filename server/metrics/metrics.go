package metrics

import (
	"context"
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
)

type identifierWithGauge struct {
	name     string
	gaugeVec *prometheus.GaugeVec
}

func (i *identifierWithGauge) Set(source string, val float64) {
	i.gaugeVec.WithLabelValues(source).Set(val)
}

type PrometheusMetrics struct {
	authenticatedWebhooks identifierWithGauge
	successfulWebhooks    identifierWithGauge
	redisClient           *redis.Client
}

func NewPrometheusMetrics(redisClient *redis.Client) *PrometheusMetrics {
	metrics := &PrometheusMetrics{
		authenticatedWebhooks: identifierWithGauge{
			name: "iot_webhook_authenticated_total",
			gaugeVec: promauto.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: "iot",
				Subsystem: "webhook",
				Name:      "authenticated_total",
				Help:      "The total number of authenticated webhooks",
			},
				[]string{"source"},
			),
		},
		successfulWebhooks: identifierWithGauge{
			name: "iot_webhook_successful_total",
			gaugeVec: promauto.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: "iot",
				Subsystem: "webhook",
				Name:      "successful_total",
				Help:      "The total number of successful webhooks",
			},
				[]string{"source"},
			),
		},
		redisClient: redisClient,
	}

	return metrics
}

func (m *PrometheusMetrics) IncAuthenticatedWebhook(ctx context.Context, source string) {
	m.incrementRedisAndCopyLocally(ctx, source, &m.authenticatedWebhooks)
}

func (m *PrometheusMetrics) IncSuccessfulWebhook(ctx context.Context, source string) {
	m.incrementRedisAndCopyLocally(ctx, source, &m.successfulWebhooks)
}

func (m *PrometheusMetrics) incrementRedisAndCopyLocally(ctx context.Context, source string, gaugeWithId *identifierWithGauge) {
	if m.redisClient != nil {
		key := fmt.Sprintf("%s:%s", gaugeWithId.name, source)
		val, err := m.redisClient.Incr(ctx, key).Result()
		if err != nil {
			log.Printf("Redis INCR error: %v", err)
			return
		}

		gaugeWithId.Set(source, float64(val))
	}
}
