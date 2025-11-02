// Package handlers provides the HTTP handlers for the API.
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// HistoryHandler is the HTTP handler for the /history endpoint.
type HistoryHandler struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewHistoryHandler creates a new HistoryHandler.
//
// Parameters:
//   db: The database connection.
//   logger: The logger.
//
// Returns:
//   A pointer to a new HistoryHandler.
func NewHistoryHandler(db *sql.DB, logger *zap.Logger) *HistoryHandler {
	return &HistoryHandler{
		db:     db,
		logger: logger,
	}
}

// ServeHTTP handles the HTTP request for the /history endpoint.
// It retrieves the estimation history from the database and encodes it as JSON.
//
// Parameters:
//   w: The http.ResponseWriter to write the response to.
//   r: The http.Request to handle.
func (h *HistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]
	repo := vars["repo"]
	repository := owner + "/" + repo

	rows, err := h.db.Query("SELECT pr_number, total_monthly_cost, created_at FROM estimations WHERE repository = $1 ORDER BY created_at DESC", repository)
	if err != nil {
		h.logger.Error("Failed to query estimations", zap.Error(err))
		http.Error(w, "Failed to query estimations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Estimation struct {
		PRNumber        int       `json:"pr_number"`
		TotalMonthlyCost float64   `json:"total_monthly_cost"`
		CreatedAt       time.Time `json:"created_at"`
	}

	var estimations []Estimation
	for rows.Next() {
		var estimation Estimation
		if err := rows.Scan(&estimation.PRNumber, &estimation.TotalMonthlyCost, &estimation.CreatedAt); err != nil {
			h.logger.Error("Failed to scan estimation", zap.Error(err))
			http.Error(w, "Failed to scan estimation", http.StatusInternalServerError)
			return
		}
		estimations = append(estimations, estimation)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(estimations); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
