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
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func NewRouter(estimatorSvc *service.Estimator, logger *zap.Logger, db *sql.DB, cache *cache.PricingCache, apiConfig config.APIConfig) http.Handler {
	router := mux.NewRouter()

	// Handlers
	estimateHandler := handlers.NewEstimateHandler(estimatorSvc, logger)
	statusHandler := handlers.NewStatusHandler(logger, db)
	healthHandler := handlers.NewHealthHandler(db, cache, logger)
	historyHandler := handlers.NewHistoryHandler(db, logger)

	// Protected estimate route
	protectedEstimateHandler := middleware.APIKeyAuthMiddleware(apiConfig.APIKeys)(estimateHandler)

	// Routing
	router.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	router.Handle("/estimate", protectedEstimateHandler)
	router.Handle("/status", statusHandler)
	router.HandleFunc("/health/live", healthHandler.LivenessProbe)
	router.HandleFunc("/health/ready", healthHandler.ReadinessProbe)
	router.Handle("/metrics", promhttp.Handler())
	router.Handle("/history/{owner}/{repo}", historyHandler)

	// Middleware chaining
	var handler http.Handler = router
	handler = middleware.RateLimitMiddleware(apiConfig.RateLimitPerSecond, apiConfig.RateLimitBurst)(handler)
	handler = middleware.MetricsMiddleware()(handler)
	handler = middleware.TimeoutMiddleware(30*time.Second)(handler)
	handler = middleware.RecoveryMiddleware(logger)(handler)
	handler = middleware.LoggingMiddleware(logger)(handler)
	handler = middleware.RequestIDMiddleware()(handler)
	handler = otelhttp.NewHandler(handler, "http.server")

	return handler
}
