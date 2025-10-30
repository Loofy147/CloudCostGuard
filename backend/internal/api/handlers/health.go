package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"cloudcostguard/backend/internal/cache"
	"go.uber.org/zap"
)

type HealthHandler struct {
	db     *sql.DB
	cache  *cache.PricingCache
	logger *zap.Logger
}

func NewHealthHandler(db *sql.DB, cache *cache.PricingCache, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

func (h *HealthHandler) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

func (h *HealthHandler) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := make(map[string]string)
	allHealthy := true

	// Check database
	if err := h.db.PingContext(ctx); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		checks["database"] = "healthy"
	}

	// Check cache
	if !h.cache.IsReady() {
		checks["cache"] = "unhealthy: no pricing data"
		allHealthy = false
	} else {
		checks["cache"] = "healthy"
	}

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": checks,
		"ready":  allHealthy,
	})
}
