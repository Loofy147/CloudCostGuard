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
		cost, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, cost, 0.01)
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
		cost, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, cost, 0.01)
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
		cost, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, cost, 0.01)
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
		cost, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, cost, 0.01)
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
		cost, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, cost, 0.01)
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
		cost, err := Estimate(plan, mockPrices, usEastRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, cost, 0.01)
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
		cost, err := Estimate(plan, mockPrices, usEastRegion, usage)
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, cost, 0.01)
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
		cost, err := Estimate(plan, mockPrices, euWestRegion, &UsageEstimates{})
		assert.NoError(t, err)
		assert.InDelta(t, expectedCost, cost, 0.01)
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
		cost, err := Estimate(plan, mockPrices, "ap-southeast-2", &UsageEstimates{})
		assert.NoError(t, err)
		assert.Equal(t, 0.0, cost)
	})
}
