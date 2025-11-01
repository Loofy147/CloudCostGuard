// Package api provides the HTTP routing for the backend service.
package api

import (
	"database/sql"
	"net/http"
	"time"

	"cloudcostguard/backend/internal/api/handlers"
	"cloudcostguard/backend/internal/api/middleware"
	"cloudcostguard/backend/internal/cache"
	"cloudcostguard/backend/internal/config"
	"cloudcostguard/backend/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.uber.org/zap"
)

// NewRouter creates a new http.Handler with all the routes and middleware configured.
//
// Parameters:
//   estimatorSvc: The estimator service.
//   logger: The logger.
//   db: The database connection.
//   cache: The pricing cache.
//   apiConfig: The API configuration.
//
// Returns:
//   An http.Handler with all the routes and middleware configured.
func NewRouter(estimatorSvc *service.Estimator, logger *zap.Logger, db *sql.DB, cache *cache.PricingCache, apiConfig config.APIConfig) http.Handler {
	mux := http.NewServeMux()

	// Handlers
	estimateHandler := handlers.NewEstimateHandler(estimatorSvc, logger)
	statusHandler := handlers.NewStatusHandler(logger, db)
	healthHandler := handlers.NewHealthHandler(db, cache, logger)

	// Protected estimate route
	protectedEstimateHandler := middleware.APIKeyAuthMiddleware(apiConfig.APIKeys)(estimateHandler)

	// Routing
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	mux.Handle("/estimate", protectedEstimateHandler)
	mux.Handle("/status", statusHandler)
	mux.HandleFunc("/health/live", healthHandler.LivenessProbe)
	mux.HandleFunc("/health/ready", healthHandler.ReadinessProbe)
	mux.Handle("/metrics", promhttp.Handler())

	// Middleware chaining
	var handler http.Handler = mux
	handler = middleware.RateLimitMiddleware(apiConfig.RateLimitPerSecond, apiConfig.RateLimitBurst)(handler)
	handler = middleware.MetricsMiddleware()(handler)
	handler = middleware.TimeoutMiddleware(30*time.Second)(handler)
	handler = middleware.RecoveryMiddleware(logger)(handler)
	handler = middleware.LoggingMiddleware(logger)(handler)
	handler = middleware.RequestIDMiddleware()(handler)
	handler = otelhttp.NewHandler(handler, "http.server")

	return handler
}
