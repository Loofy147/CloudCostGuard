package main

import (
	"cloudcostguard/backend/estimator"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormatComment(t *testing.T) {
	t.Run("formats a comment with resources", func(t *testing.T) {
		result := estimator.EstimationResponse{
			TotalMonthlyCost: 123.45,
			Currency:         "USD",
			Resources: []estimator.ResourceCost{
				{
					Address:      "aws_instance.web",
					MonthlyCost:  100.0,
					CostBreakdown: "t2.micro @ $0.123/hr",
				},
				{
					Address:      "aws_ebs_volume.data",
					MonthlyCost:  23.45,
					CostBreakdown: "100 GB @ $0.2345/GB-mo",
				},
			},
		}

		comment := formatComment(result)

		assert.Contains(t, comment, "Estimated Monthly Cost Impact: **$123.45**")
		assert.Contains(t, comment, "| Resource | Monthly Cost | Details |")
		assert.Contains(t, comment, "| `aws_instance.web` | `$100.00` | t2.micro @ $0.123/hr |")
		assert.Contains(t, comment, "| `aws_ebs_volume.data` | `$23.45` | 100 GB @ $0.2345/GB-mo |")
	})

	t.Run("formats a comment with no resources", func(t *testing.T) {
		result := estimator.EstimationResponse{
			TotalMonthlyCost: 0.0,
			Currency:         "USD",
			Resources:        []estimator.ResourceCost{},
		}

		comment := formatComment(result)

		assert.Contains(t, comment, "Estimated Monthly Cost Impact: **$0.00**")
		assert.NotContains(t, comment, "| Resource | Monthly Cost | Details |")
	})
}
