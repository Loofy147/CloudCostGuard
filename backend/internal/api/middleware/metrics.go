package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	EstimationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "estimation_duration_seconds",
			Help:    "Time to compute cost estimation",
			Buckets: []float64{.01, .05, .1, .5, 1, 2.5, 5},
		},
	)
)

func MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start).Seconds()

			httpRequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				strconv.Itoa(wrapped.statusCode),
			).Inc()

			httpRequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(duration)
		})
	}
}
