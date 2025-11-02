package service

import (
	"cloudcostguard/backend/estimator"
	"cloudcostguard/backend/internal/api/middleware"
	"cloudcostguard/backend/internal/cache"
	"cloudcostguard/backend/terraform"
	"database/sql"
	"go.uber.org/zap"
	"time"
)

type Estimator struct {
	pricingCache *cache.PricingCache
	logger       *zap.Logger
	db           *sql.DB
}

func NewEstimator(pricingCache *cache.PricingCache, logger *zap.Logger, db *sql.DB) *Estimator {
	return &Estimator{
		pricingCache: pricingCache,
		logger:       logger,
		db:           db,
	}
}

func (s *Estimator) Estimate(plan *terraform.Plan, region string, usageEstimates *estimator.UsageEstimates) (*estimator.EstimationResponse, error) {
	startTime := time.Now()
	defer func() {
		middleware.EstimationDuration.Observe(time.Since(startTime).Seconds())
	}()

	priceList := s.pricingCache.Get()
	if priceList == nil {
		s.logger.Error("Pricing data is not available")
		return nil, &ServiceUnavailableError{"Pricing data is not available"}
	}
	return estimator.Estimate(plan, priceList, region, usageEstimates)
}

type ServiceUnavailableError struct {
    Message string
}

func (e *ServiceUnavailableError) Error() string {
    return e.Message
}

func (s *Estimator) SaveEstimation(repo string, prNumber int, totalMonthlyCost float64) error {
	_, err := s.db.Exec("INSERT INTO estimations (repository, pr_number, total_monthly_cost) VALUES ($1, $2, $3)", repo, prNumber, totalMonthlyCost)
	return err
}
