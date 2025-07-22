package metrics

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

func (m *PrometheusMetrics) HandlerWithRedisUpdate() http.Handler {
	baseHandler := promhttp.Handler()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		m.refreshMetricsFromRedis(ctx)
		baseHandler.ServeHTTP(w, r)
	})
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
		_, err := m.redisClient.Incr(ctx, key).Result()
		if err != nil {
			log.Printf("Redis INCR error: %v", err)
			return
		}
	}
}

func (m *PrometheusMetrics) refreshMetricsFromRedis(ctx context.Context) {
	if m.redisClient == nil {
		return
	}

	m.refreshMetricFromRedis(ctx, &m.authenticatedWebhooks)
	m.refreshMetricFromRedis(ctx, &m.successfulWebhooks)
}

func (m *PrometheusMetrics) refreshMetricFromRedis(ctx context.Context, metric *identifierWithGauge) {
	pattern := fmt.Sprintf("%s:*", metric.name)
	// Realistically nowhere near this limit
	iter := m.redisClient.Scan(ctx, 0, pattern, 100).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 100 {
			m.updateGaugeFromKeys(ctx, metric, keys)
			keys = keys[:0]
		}
	}
	if len(keys) > 0 {
		m.updateGaugeFromKeys(ctx, metric, keys)
	}

	if err := iter.Err(); err != nil {
		log.Printf("Redis SCAN error: %v", err)
	}
}

func (m *PrometheusMetrics) updateGaugeFromKeys(ctx context.Context, metric *identifierWithGauge, keys []string) {
	vals, err := m.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		log.Printf("Redis MGET error: %v", err)
		return
	}

	for i, v := range vals {
		if v == nil {
			continue
		}

		// Sanity check which we shouldn't really need
		valStr, ok := v.(string)
		if !ok {
			log.Printf("Unexpected value type for key %s: %T", keys[i], v)
			continue
		}

		// Parse as integer
		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			log.Printf("Parse error for key %s: %v", keys[i], err)
			continue
		}

		parts := strings.Split(keys[i], ":")
		source := parts[len(parts)-1]

		// Set the gauge value
		metric.gaugeVec.WithLabelValues(source).Set(float64(val))
	}
}
