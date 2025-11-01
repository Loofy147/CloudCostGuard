package estimator

import (
	"cloudcostguard/backend/pricing"
	"cloudcostguard/backend/terraform"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createMockPriceList() *pricing.PriceList {
	priceList := pricing.NewPriceList()
	priceList.Terms.OnDemand = make(map[string]map[string]pricing.Term)

	// Mock EC2 t2.micro price: $10/hr
	priceList.Products["ec2-t2-micro-sku"] = pricing.Product{
		SKU: "ec2-t2-micro-sku",
		Attributes: pricing.ProductAttributes{
			ServiceCode:     "AmazonEC2",
			InstanceType:    "t2.micro",
			Location:        "US East (N. Virginia)",
			OperatingSystem: "Linux",
			UsageType:       "BoxUsage:t2.micro",
		},
	}
	pd1 := pricing.PriceDimension{}
	pd1.PricePerUnit.USD = "10.0"
	priceList.Terms.OnDemand["ec2-t2-micro-sku"] = map[string]pricing.Term{
		"term1": {
			PriceDimensions: map[string]pricing.PriceDimension{
				"dim1": pd1,
			},
		},
	}

	// Mock NAT Gateway price: $0.045/hr
	priceList.Products["nat-gateway-sku"] = pricing.Product{
		SKU: "nat-gateway-sku",
		Attributes: pricing.ProductAttributes{
			ServiceCode: "AmazonVPC",
			Group:       "NAT Gateway",
			Location:    "US East (N. Virginia)",
		},
	}
	pd4 := pricing.PriceDimension{}
	pd4.PricePerUnit.USD = "0.045"
	priceList.Terms.OnDemand["nat-gateway-sku"] = map[string]pricing.Term{
		"term1": {
			PriceDimensions: map[string]pricing.PriceDimension{
				"dim1": pd4,
			},
		},
	}

	// Mock NAT Gateway Data Processing price: $0.045/GB
	priceList.Products["nat-gateway-dp-sku"] = pricing.Product{
		SKU: "nat-gateway-dp-sku",
		Attributes: pricing.ProductAttributes{
			ServiceCode: "AmazonVPC",
			Location:    "US East (N. Virginia)",
			UsageType:   "NatGateway-Bytes",
		},
	}
	pd6 := pricing.PriceDimension{}
	pd6.PricePerUnit.USD = "0.045"
	priceList.Terms.OnDemand["nat-gateway-dp-sku"] = map[string]pricing.Term{
		"term1": {
			PriceDimensions: map[string]pricing.PriceDimension{
				"dim1": pd6,
			},
		},
	}

	// Mock EC2 t2.small price: $20/hr
	priceList.Products["ec2-t2-small-sku"] = pricing.Product{
		SKU: "ec2-t2-small-sku",
		Attributes: pricing.ProductAttributes{
			ServiceCode:     "AmazonEC2",
			InstanceType:    "t2.small",
			Location:        "US East (N. Virginia)",
			OperatingSystem: "Linux",
			UsageType:       "BoxUsage:t2.small",
		},
	}
	pd2 := pricing.PriceDimension{}
	pd2.PricePerUnit.USD = "20.0"
	priceList.Terms.OnDemand["ec2-t2-small-sku"] = map[string]pricing.Term{
		"term1": {
			PriceDimensions: map[string]pricing.PriceDimension{
				"dim1": pd2,
			},
		},
	}

	// Mock EBS gp2 price: $0.10/GB-month
	priceList.Products["ebs-gp2-sku"] = pricing.Product{
		SKU: "ebs-gp2-sku",
		Attributes: pricing.ProductAttributes{
			ServiceCode:   "AmazonEC2",
			VolumeAPIName: "gp2",
			Location:      "US East (N. Virginia)",
		},
	}
	pd3 := pricing.PriceDimension{}
	pd3.PricePerUnit.USD = "0.10"
	priceList.Terms.OnDemand["ebs-gp2-sku"] = map[string]pricing.Term{
		"term1": {
			PriceDimensions: map[string]pricing.PriceDimension{
				"dim1": pd3,
			},
		},
	}

	// Mock EC2 t2.micro price in a different region: $12/hr
	priceList.Products["ec2-t2-micro-eu-sku"] = pricing.Product{
		SKU: "ec2-t2-micro-eu-sku",
		Attributes: pricing.ProductAttributes{
			ServiceCode:     "AmazonEC2",
			InstanceType:    "t2.micro",
			Location:        "EU (Ireland)",
			OperatingSystem: "Linux",
			UsageType:       "BoxUsage:t2.micro",
		},
	}
	pd5 := pricing.PriceDimension{}
	pd5.PricePerUnit.USD = "12.0"
	priceList.Terms.OnDemand["ec2-t2-micro-eu-sku"] = map[string]pricing.Term{
		"term1": {
			PriceDimensions: map[string]pricing.PriceDimension{
				"dim1": pd5,
			},
		},
	}

	// Mock Fargate vCPU price: $0.04048/hr
	priceList.Products["fargate-vcpu-sku"] = pricing.Product{
		SKU: "fargate-vcpu-sku",
		Attributes: pricing.ProductAttributes{
			ServiceCode: "AmazonECS",
			Location:    "US East (N. Virginia)",
			UsageType:   "Fargate-vCPU-Hours:perCPU",
		},
	}
	pd7 := pricing.PriceDimension{}
	pd7.PricePerUnit.USD = "0.04048"
	priceList.Terms.OnDemand["fargate-vcpu-sku"] = map[string]pricing.Term{
		"term1": {
			PriceDimensions: map[string]pricing.PriceDimension{
				"dim1": pd7,
			},
		},
	}

	// Mock Fargate Memory price: $0.004445/GB-hr
	priceList.Products["fargate-memory-sku"] = pricing.Product{
		SKU: "fargate-memory-sku",
		Attributes: pricing.ProductAttributes{
			ServiceCode: "AmazonECS",
			Location:    "US East (N. Virginia)",
			UsageType:   "Fargate-GB-Hours:perGB",
		},
	}
	pd8 := pricing.PriceDimension{}
	pd8.PricePerUnit.USD = "0.004445"
	priceList.Terms.OnDemand["fargate-memory-sku"] = map[string]pricing.Term{
		"term1": {
			PriceDimensions: map[string]pricing.PriceDimension{
				"dim1": pd8,
			},
		},
	}

	// Mock EKS Control Plane price: $0.10/hr
	priceList.Products["eks-control-plane-sku"] = pricing.Product{
		SKU: "eks-control-plane-sku",
		Attributes: pricing.ProductAttributes{
			ServiceCode: "AmazonEKS",
			Location:    "US East (N. Virginia)",
			UsageType:   "EKS-Hours:perCluster",
		},
	}
	pd9 := pricing.PriceDimension{}
	pd9.PricePerUnit.USD = "0.10"
	priceList.Terms.OnDemand["eks-control-plane-sku"] = map[string]pricing.Term{
		"term1": {
			PriceDimensions: map[string]pricing.PriceDimension{
				"dim1": pd9,
			},
		},
	}

	return priceList
}

func TestEstimate(t *testing.T) {
	mockPrices := createMockPriceList()
	usEastRegion := "US East (N. Virginia)"
	euWestRegion := "EU (Ireland)"

	t.Run("estimates cost for new EC2 instance", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_instance.web",
					Type:    "aws_instance",
					Change:  terraform.Change{Actions: []string{"create"}},
					After:   map[string]interface{}{"instance_type": "t2.micro"},
				},
			},
		}

		// Expected cost: $10/hr * 730 hrs/month = $7300
		expectedCost := 10.0 * 730
		result, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, result.TotalMonthlyCost, 0.01)
		assert.Len(t, result.Resources, 1)
		assert.Equal(t, "aws_instance.web", result.Resources[0].Address)
		assert.InDelta(t, expectedCost, result.Resources[0].MonthlyCost, 0.01)
	})

	t.Run("estimates cost for deleted EC2 instance", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_instance.web",
					Type:    "aws_instance",
					Change:  terraform.Change{Actions: []string{"delete"}},
					Before:  map[string]interface{}{"instance_type": "t2.micro"},
				},
			},
		}

		// Expected cost: -$10/hr * 730 hrs/month = -$7300
		expectedCost := -10.0 * 730
		result, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, result.TotalMonthlyCost, 0.01)
		assert.Len(t, result.Resources, 1)
	})

	t.Run("estimates cost for updated EC2 instance", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_instance.web",
					Type:    "aws_instance",
					Change:  terraform.Change{Actions: []string{"update"}},
					Before:  map[string]interface{}{"instance_type": "t2.micro"},
					After:   map[string]interface{}{"instance_type": "t2.small"},
				},
			},
		}

		// Expected cost: ($20/hr - $10/hr) * 730 hrs/month = $7300
		expectedCost := (20.0 - 10.0) * 730
		result, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, result.TotalMonthlyCost, 0.01)
		assert.Len(t, result.Resources, 1)
	})

	t.Run("estimates cost for multiple resources", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_instance.web",
					Type:    "aws_instance",
					Change:  terraform.Change{Actions: []string{"create"}},
					After:   map[string]interface{}{"instance_type": "t2.micro"},
				},
				{
					Address: "aws_ebs_volume.data",
					Type:    "aws_ebs_volume",
					Change:  terraform.Change{Actions: []string{"create"}},
					After:   map[string]interface{}{"type": "gp2", "size": float64(100)},
				},
			},
		}

		// Expected cost: ($10/hr * 730) + ($0.10/GB * 100 GB) = 7300 + 10 = $7310
		expectedCost := (10.0 * 730) + (0.10 * 100)
		result, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, result.TotalMonthlyCost, 0.01)
		assert.Len(t, result.Resources, 2)
	})

	t.Run("skips unsupported resources", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "null_resource.foo",
					Type:    "null_resource",
					Change:  terraform.Change{Actions: []string{"create"}},
					After:   map[string]interface{}{},
				},
				{
					Address: "aws_instance.web",
					Type:    "aws_instance",
					Change:  terraform.Change{Actions: []string{"create"}},
					After:   map[string]interface{}{"instance_type": "t2.micro"},
				},
			},
		}

		expectedCost := 10.0 * 730
		result, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, result.TotalMonthlyCost, 0.01)
		assert.Len(t, result.Resources, 1)
	})

	t.Run("estimates cost for a new NAT Gateway", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_nat_gateway.gw",
					Type:    "aws_nat_gateway",
					Change:  terraform.Change{Actions: []string{"create"}},
					After:   map[string]interface{}{},
				},
			},
		}

		// Expected cost: $0.045/hr * 730 hrs/month = $32.85
		expectedCost := 0.045 * 730
		result, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, result.TotalMonthlyCost, 0.01)
		assert.Len(t, result.Resources, 1)
	})

	t.Run("estimates cost for a new NAT Gateway with usage", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_nat_gateway.gw",
					Type:    "aws_nat_gateway",
					Change:  terraform.Change{Actions: []string{"create"}},
					After:   map[string]interface{}{},
				},
			},
		}

		usage := &UsageEstimates{NATGatewayGBProcessed: 1000} // 1000 GB/month

		// Expected fixed cost: $0.045/hr * 730 hrs/month = $32.85
		// Expected usage cost: 1000 GB * $0.045/GB = $45
		// Total: $77.85
		expectedCost := (0.045 * 730) + (1000 * 0.045)
		result, err := Estimate(plan, mockPrices, usEastRegion, usage)
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, result.TotalMonthlyCost, 0.01)
		assert.Len(t, result.Resources, 1)
	})

	t.Run("uses the correct region for pricing", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_instance.web",
					Type:    "aws_instance",
					Change:  terraform.Change{Actions: []string{"create"}},
					After:   map[string]interface{}{"instance_type": "t2.micro"},
				},
			},
		}

		// Expected cost for eu-west-1: $12/hr * 730 hrs/month = $8760
		expectedCost := 12.0 * 730
		result, err := Estimate(plan, mockPrices, euWestRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, result.TotalMonthlyCost, 0.01)
		assert.Len(t, result.Resources, 1)
	})

	t.Run("returns zero cost for resources in a region with no pricing data", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_instance.web",
					Type:    "aws_instance",
					Change:  terraform.Change{Actions: []string{"create"}},
					After:   map[string]interface{}{"instance_type": "t2.micro"},
				},
			},
		}

		// We have no mock data for ap-southeast-2, so the cost should be 0
		result, err := Estimate(plan, mockPrices, "ap-southeast-2", &UsageEstimates{})
		assert.NoError(t, err)
		assert.Equal(t, 0.0, result.TotalMonthlyCost)
		assert.Len(t, result.Resources, 0)
	})

	t.Run("estimates cost for a new Lambda function", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_lambda_function.test",
					Type:    "aws_lambda_function",
					After: map[string]interface{}{
						"memory_size": float64(512),
					},
					Change: terraform.Change{
						Actions: []string{"create"},
					},
				},
			},
		}
		priceList := &pricing.PriceList{
			Products: map[string]pricing.Product{
				"lambda-request": {
					Attributes: pricing.ProductAttributes{
						ServiceCode: "AWSLambda",
						UsageType:   "Request",
						Location:    "US East (N. Virginia)",
					},
				},
				"lambda-gb-second": {
					Attributes: pricing.ProductAttributes{
						ServiceCode: "AWSLambda",
						UsageType:   "GB-Second",
						Location:    "US East (N. Virginia)",
					},
				},
			},
			Terms: struct {
				OnDemand map[string]map[string]pricing.Term `json:"OnDemand"`
			}{
				OnDemand: map[string]map[string]pricing.Term{
					"lambda-request": {
						"term": {
							PriceDimensions: map[string]pricing.PriceDimension{
								"dim": {
									PricePerUnit: struct {
										USD string `json:"USD"`
									}{
										USD: "0.0000002",
									},
								},
							},
						},
					},
					"lambda-gb-second": {
						"term": {
							PriceDimensions: map[string]pricing.PriceDimension{
								"dim": {
									PricePerUnit: struct {
										USD string `json:"USD"`
									}{
										USD: "0.0000166667",
									},
								},
							},
						},
					},
				},
			},
		}
		usage := &UsageEstimates{
			LambdaMonthlyRequests: 2000000,
			LambdaAvgDurationMS:   500,
		}

		resp, err := Estimate(plan, priceList, "us-east-1", usage)
		assert.NoError(t, err)
		assert.Len(t, resp.Resources, 1)
		assert.InDelta(t, 1.87, resp.Resources[0].MonthlyCost, 0.01)
	})

	t.Run("estimates cost for a new ECS service on Fargate", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_ecs_service.web",
					Type:    "aws_ecs_service",
					Change:  terraform.Change{Actions: []string{"create"}},
					After: map[string]interface{}{
						"desired_count":   float64(2),
						"launch_type":     "FARGATE",
						"task_definition": "aws_ecs_task_definition.web",
					},
				},
				{
					Address: "aws_ecs_task_definition.web",
					Type:    "aws_ecs_task_definition",
					After: map[string]interface{}{
						"cpu":    "1024",
						"memory": "2048",
					},
				},
			},
		}

		// vCPU cost: (1024 / 1024) * $0.04048/hr = $0.04048/hr
		// Memory cost: (2048 / 1024) * $0.004445/hr = $0.00889/hr
		// Total hourly cost per task: $0.04937
		// Total monthly cost for 2 tasks: $0.04937 * 2 * 730 = $72.08
		expectedCost := ((1 * 0.04048) + (2 * 0.004445)) * 2 * 730
		result, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, result.TotalMonthlyCost, 0.01)
		assert.Len(t, result.Resources, 1)
	})

	t.Run("estimates cost for a new ECS service on EC2", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_ecs_service.web",
					Type:    "aws_ecs_service",
					Change:  terraform.Change{Actions: []string{"create"}},
					After: map[string]interface{}{
						"desired_count": float64(1),
						"launch_type":   "EC2",
					},
				},
				{
					Address: "aws_instance.ecs_instance",
					Type:    "aws_instance",
					Change:  terraform.Change{Actions: []string{"create"}},
					After: map[string]interface{}{
						"instance_type": "t2.micro",
					},
				},
			},
		}

		// Expected cost: 1 t2.micro instance * $10/hr * 730 hrs/month = $7300
		expectedCost := 10.0 * 730
		result, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, result.TotalMonthlyCost, 0.01)
	})

	t.Run("does not double-count ECS task definitions", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_ecs_task_definition.web",
					Type:    "aws_ecs_task_definition",
					Change:  terraform.Change{Actions: []string{"create"}},
					After: map[string]interface{}{
						"cpu":    "1024",
						"memory": "2048",
					},
				},
			},
		}

		result, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.Equal(t, 0.0, result.TotalMonthlyCost)
	})

	t.Run("estimates cost for a new EKS cluster with a node group", func(t *testing.T) {
		plan := &terraform.Plan{
			ResourceChanges: []*terraform.ResourceChange{
				{
					Address: "aws_eks_cluster.main",
					Type:    "aws_eks_cluster",
					Change:  terraform.Change{Actions: []string{"create"}},
					After:   map[string]interface{}{},
				},
				{
					Address: "aws_eks_node_group.workers",
					Type:    "aws_eks_node_group",
					Change:  terraform.Change{Actions: []string{"create"}},
					After: map[string]interface{}{
						"instance_types": []interface{}{"t2.micro"},
						"scaling_config": []interface{}{
							map[string]interface{}{
								"desired_size": float64(3),
							},
						},
					},
				},
			},
		}

		// EKS Control Plane: $0.10/hr * 730 hrs/month = $73
		// EKS Node Group: 3 * t2.micro ($10/hr) * 730 hrs/month = $21900
		// Total: $21973
		expectedCost := (0.10 * 730) + (3 * 10.0 * 730)
		result, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, result.TotalMonthlyCost, 0.01)
		assert.Len(t, result.Resources, 2)
	})
}
