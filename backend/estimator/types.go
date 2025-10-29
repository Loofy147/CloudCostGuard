package estimator

import (
	"cloudcostguard/backend/terraform"
)

// EstimateRequest defines the structure of the request body for the /estimate endpoint.
type EstimateRequest struct {
	Plan           *terraform.Plan `json:"plan"`
	UsageEstimates UsageEstimates    `json:"usage_estimates"`
}

// UsageEstimates represents the structure of the usage_estimates block in the config file.
type UsageEstimates struct {
	NATGatewayGBProcessed int `yaml:"nat_gateway_gb_processed" json:"nat_gateway_gb_processed"`
}

// EstimationResponse defines the structure of the response body for the /estimate endpoint.
type EstimationResponse struct {
	TotalMonthlyCost float64        `json:"total_monthly_cost"`
	Currency         string         `json:"currency"`
	Resources        []ResourceCost `json:"resources"`
}

// ResourceCost represents the cost of a single resource.
type ResourceCost struct {
	Address      string  `json:"address"`
	MonthlyCost  float64 `json:"monthly_cost"`
	CostBreakdown string  `json:"cost_breakdown"`
}
