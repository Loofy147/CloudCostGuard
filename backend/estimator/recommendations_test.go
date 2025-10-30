package estimator

import (
	"testing"

	"cloudcostguard/backend/terraform"
	"github.com/stretchr/testify/assert"
)

func TestGenerateRecommendations(t *testing.T) {
	plan := &terraform.Plan{
		ResourceChanges: []*terraform.ResourceChange{
			{
				Address: "aws_nat_gateway.main",
				Type:    "aws_nat_gateway",
				Change: terraform.Change{
					Actions: []string{"create"},
				},
			},
			{
				Address: "aws_db_instance.pricing_db",
				Type:    "aws_db_instance",
				After: map[string]interface{}{
					"multi_az":       true,
					"instance_class": "db.t3.medium",
				},
				Change: terraform.Change{
					Actions: []string{"create"},
				},
			},
			{
				Address: "aws_instance.backend_nodes",
				Type:    "aws_instance",
				After: map[string]interface{}{
					"instance_type": "t3.medium",
				},
				Change: terraform.Change{
					Actions: []string{"create"},
				},
			},
			{
				Address: "aws_ebs_volume.app_logs",
				Type:    "aws_ebs_volume",
				After: map[string]interface{}{
					"size": float64(50),
				},
				Change: terraform.Change{
					Actions: []string{"create"},
				},
			},
		},
	}
	usage := &UsageEstimates{
		NATGatewayGBProcessed: 5000,
	}

	recommendations := GenerateRecommendations(plan, usage)

	assert.Len(t, recommendations, 5)
	assert.Contains(t, recommendations, "💡 NAT Gateway data transfer is expensive - consider VPC endpoints for AWS services")
	assert.Contains(t, recommendations, "💡 Consider Reserved Instances for RDS - could save ~40%")
	assert.Contains(t, recommendations, "💡 Use t3.small RDS instances if workload allows - save ~$80/month")
	assert.Contains(t, recommendations, "💡 Use t3.small instances if workload allows - save ~$45/month")
	assert.Contains(t, recommendations, "💡 Enable EBS snapshot lifecycle policy to manage backup costs")
}
