package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type StatusHandler struct {
	logger *zap.Logger
	db     *sql.DB
}

func NewStatusHandler(logger *zap.Logger, db *sql.DB) *StatusHandler {
	return &StatusHandler{
		logger: logger,
		db:     db,
	}
}

func (h *StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var oldestTimestamp time.Time
	err := h.db.QueryRow("SELECT MIN(last_updated) FROM aws_prices").Scan(&oldestTimestamp)
	if err != nil {
		h.logger.Error("Failed to query database", zap.Error(err))
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"oldest_pricing_data": oldestTimestamp.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
