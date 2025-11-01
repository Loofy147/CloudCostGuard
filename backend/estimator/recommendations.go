// Package estimator provides the core logic for estimating the cost of Terraform plans.
package estimator

import (
	"strings"

	"cloudcostguard/backend/terraform"
)

// GenerateRecommendations analyzes a Terraform plan and suggests cost-saving optimizations.
//
// Parameters:
//   plan: The Terraform plan to analyze.
//   usage: Usage estimates for various resources.
//
// Returns:
//   A slice of strings, where each string is a cost-saving recommendation.
func GenerateRecommendations(plan *terraform.Plan, usage *UsageEstimates) []string {
	if plan == nil {
		return nil
	}
	var recommendations []string
	resourceCounts := make(map[string]int)
	for _, rc := range plan.ResourceChanges {
		resourceCounts[rc.Type]++

		switch rc.Type {
		case "aws_nat_gateway":
			recommendations = append(recommendations, checkNATGateway(rc, usage)...)
		case "aws_db_instance":
			recommendations = append(recommendations, checkDBInstance(rc)...)
		case "aws_instance":
			recommendations = append(recommendations, checkInstance(rc)...)
		case "aws_ebs_volume":
			recommendations = append(recommendations, checkEBSVolume(rc)...)
		}
	}

	return recommendations
}

// checkNATGateway checks for potential cost savings on a NAT Gateway.
func checkNATGateway(rc *terraform.ResourceChange, usage *UsageEstimates) []string {
	if usage != nil && usage.NATGatewayGBProcessed > 100 { // Recommend if over 100GB processed
		return []string{"💡 NAT Gateway data transfer is expensive - consider VPC endpoints for AWS services"}
	}
	return nil
}

// checkDBInstance checks for potential cost savings on an RDS instance.
func checkDBInstance(rc *terraform.ResourceChange) []string {
	var recs []string
	if multiAZ, ok := rc.After["multi_az"].(bool); ok && multiAZ {
		recs = append(recs, "💡 Consider Reserved Instances for RDS - could save ~40%")
	}
	if instanceClass, ok := rc.After["instance_class"].(string); ok {
		if strings.Contains(instanceClass, "t3.medium") {
			recs = append(recs, "💡 Use t3.small RDS instances if workload allows - save ~$80/month")
		}
		if !strings.HasPrefix(instanceClass, "db.r6g") && !strings.HasPrefix(instanceClass, "db.m6g") {
			recs = append(recs, "💡 Consider switching to ARM-based Graviton RDS instances for better price-performance.")
		}
	}
	return recs
}

// checkInstance checks for potential cost savings on an EC2 instance.
func checkInstance(rc *terraform.ResourceChange) []string {
	var recs []string
	if instanceType, ok := rc.After["instance_type"].(string); ok {
		if strings.Contains(instanceType, "t3.medium") {
			recs = append(recs, "💡 Use t3.small instances if workload allows - save ~$45/month")
		}
		if !strings.HasPrefix(instanceType, "t4g") {
			recs = append(recs, "💡 Consider switching to ARM-based Graviton EC2 instances for better price-performance.")
		}
	}
	return recs
}

// checkEBSVolume checks for potential cost savings on an EBS volume.
func checkEBSVolume(rc *terraform.ResourceChange) []string {
	if size, ok := rc.After["size"].(float64); ok && size > 20 {
		return []string{"💡 Enable EBS snapshot lifecycle policy to manage backup costs"}
	}
	return nil
}
