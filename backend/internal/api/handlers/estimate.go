package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"cloudcostguard/backend/estimator"
	"cloudcostguard/backend/internal/service"
	"cloudcostguard/backend/terraform"
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

	if err := validatePlan(plan); err != nil {
		h.logger.Error("Invalid plan", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	repo := r.URL.Query().Get("repo")
	prNumberStr := r.URL.Query().Get("prNumber")
	prNumber, _ := strconv.Atoi(prNumberStr)

	if err := h.estimator.SaveEstimation(repo, prNumber, cost.TotalMonthlyCost); err != nil {
		h.logger.Error("Failed to save estimation", zap.Error(err))
		// We don't return an error to the user, as the cost estimation itself was successful.
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cost); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func validatePlan(plan *terraform.Plan) error {
	if plan == nil {
		return fmt.Errorf("plan cannot be nil")
	}

	if len(plan.ResourceChanges) > 1000 {
		return fmt.Errorf("too many resources in plan (max 1000)")
	}

	for _, rc := range plan.ResourceChanges {
		if rc.Address == "" {
			return fmt.Errorf("resource address cannot be empty")
		}
		if len(rc.Address) > 256 {
			return fmt.Errorf("resource address too long")
		}
	}

	return nil
}
