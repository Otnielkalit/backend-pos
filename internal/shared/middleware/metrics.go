package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTP metrics exposed at /metrics and scraped by Prometheus.
// Using promauto so metrics self-register without manual prometheus.MustRegister calls.
var (
	// httpRequestsTotal tracks total HTTP requests, labeled by method, route, and status code.
	// Uses c.FullPath() (e.g. /api/v1/products/:id) instead of c.Request.URL.Path
	// to avoid high cardinality from dynamic path parameters.
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"method", "route", "status"},
	)

	// httpRequestDuration tracks request latency as a histogram.
	// Buckets cover typical API response times: 5ms to 10s.
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "route"},
	)

	// httpRequestsInFlight tracks how many requests are currently being processed.
	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)
)

// PrometheusMetrics returns a Gin middleware that records HTTP metrics for Prometheus.
// Register this AFTER RequestLogger so both run independently.
//
// Excluded paths: /metrics and /health (internal endpoints, not business traffic).
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip internal/operational endpoints to avoid polluting business metrics
		path := c.Request.URL.Path
		if path == "/metrics" || path == "/health" {
			c.Next()
			return
		}

		start := time.Now()
		httpRequestsInFlight.Inc()

		c.Next()

		httpRequestsInFlight.Dec()
		duration := time.Since(start).Seconds()

		// FullPath returns the registered route pattern (e.g. /api/v1/products/:id),
		// not the actual URL — this prevents high cardinality from path params.
		route := c.FullPath()
		if route == "" {
			route = "unmatched" // 404 routes
		}

		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, route, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, route).Observe(duration)
	}
}
