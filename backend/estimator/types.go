// Package estimator provides the core logic for estimating the cost of Terraform plans.
package estimator

import (
	"cloudcostguard/backend/terraform"
)

// EstimateRequest defines the structure of the request body for the /estimate endpoint.
type EstimateRequest struct {
	// Plan is the Terraform plan to estimate.
	Plan           *terraform.Plan `json:"plan"`
	// UsageEstimates contains usage estimates for various resources.
	UsageEstimates UsageEstimates    `json:"usage_estimates"`
}

// UsageEstimates represents the structure of the usage_estimates block in the config file.
type UsageEstimates struct {
	// NATGatewayGBProcessed is the estimated GB of data processed by the NAT Gateway per month.
	NATGatewayGBProcessed int `yaml:"nat_gateway_gb_processed" json:"nat_gateway_gb_processed"`
	// LambdaMonthlyRequests is the estimated number of monthly requests for the Lambda function.
	LambdaMonthlyRequests int `yaml:"lambda_monthly_requests" json:"lambda_monthly_requests"`
	// LambdaAvgDurationMS is the estimated average duration of the Lambda function in milliseconds.
	LambdaAvgDurationMS   int `yaml:"lambda_avg_duration_ms" json:"lambda_avg_duration_ms"`
	// S3StorageGB is the estimated storage in GB for the S3 bucket.
	S3StorageGB           int `yaml:"s3_storage_gb" json:"s3_storage_gb"`
	// S3MonthlyPutRequests is the estimated number of monthly PUT requests for the S3 bucket.
	S3MonthlyPutRequests  int `yaml:"s3_monthly_put_requests" json:"s3_monthly_put_requests"`
}

// EstimationResponse defines the structure of the response body for the /estimate endpoint.
type EstimationResponse struct {
	// TotalMonthlyCost is the total estimated monthly cost of the resources in the plan.
	TotalMonthlyCost float64        `json:"total_monthly_cost"`
	// Currency is the currency of the cost estimate.
	Currency         string         `json:"currency"`
	// Resources is a slice of ResourceCost structs, each representing the cost of a single resource.
	Resources        []ResourceCost `json:"resources"`
	// Recommendations is a slice of strings, where each string is a cost-saving recommendation.
	Recommendations  []string       `json:"recommendations"`
}

// ResourceCost represents the cost of a single resource.
type ResourceCost struct {
	// Address is the address of the resource in the Terraform plan.
	Address      string  `json:"address"`
	// MonthlyCost is the estimated monthly cost of the resource.
	MonthlyCost  float64 `json:"monthly_cost"`
	// CostBreakdown is a string describing the breakdown of the cost.
	CostBreakdown string  `json:"cost_breakdown"`
}
