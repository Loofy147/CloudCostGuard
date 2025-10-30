package api

import (
	"database/sql"
	"net/http"
	"time"

	"cloudcostguard/backend/internal/api/handlers"
	"cloudcostguard/backend/internal/api/middleware"
	"cloudcostguard/backend/internal/cache"
	"cloudcostguard/backend/internal/service"
	"github.com/swaggo/http-swagger"

	"go.uber.org/zap"
)

func NewRouter(estimatorSvc *service.Estimator, logger *zap.Logger, db *sql.DB, cache *cache.PricingCache) http.Handler {
	mux := http.NewServeMux()

	// Handlers
	estimateHandler := handlers.NewEstimateHandler(estimatorSvc, logger)
	statusHandler := handlers.NewStatusHandler(logger, db)
	healthHandler := handlers.NewHealthHandler(db, cache, logger)

	// Routing
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	mux.Handle("/estimate", estimateHandler)
	mux.Handle("/status", statusHandler)
	mux.HandleFunc("/health/live", healthHandler.LivenessProbe)
	mux.HandleFunc("/health/ready", healthHandler.ReadinessProbe)

	// Middleware chaining
	var handler http.Handler = mux
	handler = middleware.TimeoutMiddleware(30*time.Second)(handler)
	handler = middleware.RecoveryMiddleware(logger)(handler)
	handler = middleware.LoggingMiddleware(logger)(handler)

	return handler
}
