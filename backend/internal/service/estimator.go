package service

import (
	"cloudcostguard/backend/estimator"
	"cloudcostguard/backend/internal/cache"
	"cloudcostguard/backend/terraform"
	"go.uber.org/zap"
)

type Estimator struct {
	pricingCache *cache.PricingCache
	logger       *zap.Logger
}

func NewEstimator(pricingCache *cache.PricingCache, logger *zap.Logger) *Estimator {
	return &Estimator{
		pricingCache: pricingCache,
		logger:       logger,
	}
}

func (s *Estimator) Estimate(plan *terraform.Plan, region string, usageEstimates *estimator.UsageEstimates) (*estimator.EstimationResponse, error) {
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
