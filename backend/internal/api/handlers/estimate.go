package handlers

import (
	"encoding/json"
	"net/http"

	"cloudcostguard/backend/estimator"
	"cloudcostguard/backend/internal/service"
	"go.uber.org/zap"
)

type EstimateHandler struct {
	estimator *service.Estimator
	logger    *zap.Logger
}

func NewEstimateHandler(estimator *service.Estimator, logger *zap.Logger) *EstimateHandler {
	return &EstimateHandler{
		estimator: estimator,
		logger:    logger,
	}
}

func (h *EstimateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	region := r.URL.Query().Get("region")
	if region == "" {
		region = "us-east-1"
	}

	var requestBody estimator.EstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	plan := requestBody.Plan

	cost, err := h.estimator.Estimate(plan, region, &requestBody.UsageEstimates)
	if err != nil {
        if _, ok := err.(*service.ServiceUnavailableError); ok {
            http.Error(w, err.Error(), http.StatusServiceUnavailable)
        } else {
            h.logger.Error("Failed to estimate cost", zap.Error(err))
            http.Error(w, "Failed to estimate cost", http.StatusInternalServerError)
        }
        return
    }

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cost); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
