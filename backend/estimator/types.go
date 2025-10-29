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
