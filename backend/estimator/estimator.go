package estimator

import (
	"fmt"
	"strconv"
	"strings"
	"cloudcostguard/backend/pricing"
	"cloudcostguard/backend/terraform"
)

type Cost struct {
	Value float64
	Unit  string // "hourly" or "monthly"
}

// Estimate calculates the estimated monthly cost impact of a Terraform plan.
func Estimate(plan *terraform.Plan, priceList *pricing.PriceList) (float64, error) {
	var totalMonthlyCost float64
	for _, rc := range plan.ResourceChanges {
		cost, err := estimateResourceChange(rc, priceList)
		if err != nil {
			fmt.Printf("Warning: skipping unsupported resource %s: %v\n", rc.Address, err)
			continue
		}

		if cost.Unit == "hourly" {
			totalMonthlyCost += cost.Value * 730
		} else {
			totalMonthlyCost += cost.Value
		}
	}
	return totalMonthlyCost, nil
}

func estimateResourceChange(rc *terraform.ResourceChange, priceList *pricing.PriceList) (*Cost, error) {
	costChange := &Cost{Value: 0, Unit: "monthly"} // Default to monthly for aggregation
	actions := rc.Change.Actions
	isCreate := len(actions) == 1 && actions[0] == "create"
	isDelete := len(actions) == 1 && actions[0] == "delete"
	isUpdate := (len(actions) == 1 && actions[0] == "update") || (len(actions) == 2 && actions[0] == "delete" && actions[1] == "create")

	if isCreate || isUpdate {
		cost, err := getResourceCost(rc, rc.After, priceList)
		if err != nil {
			return nil, err
		}
		if cost.Unit == "hourly" {
			costChange.Value += cost.Value * 730
		} else {
			costChange.Value += cost.Value
		}
	}

	if isDelete || isUpdate {
		cost, err := getResourceCost(rc, rc.Before, priceList)
		if err != nil {
			return nil, err
		}
		if cost.Unit == "hourly" {
			costChange.Value -= cost.Value * 730
		} else {
			costChange.Value -= cost.Value
		}
	}

	return costChange, nil
}

func getResourceCost(rc *terraform.ResourceChange, attributes map[string]interface{}, priceList *pricing.PriceList) (*Cost, error) {
	if priceList == nil {
		return nil, fmt.Errorf("pricing data is nil")
	}

	switch rc.Type {
	case "aws_instance":
		price, err := costForEC2(attributes, priceList)
		return &Cost{Value: price, Unit: "hourly"}, err
	case "aws_db_instance":
		price, err := costForRDS(attributes, priceList)
		return &Cost{Value: price, Unit: "hourly"}, err
	case "aws_ebs_volume":
		price, err := costForEBS(attributes, priceList)
		return &Cost{Value: price, Unit: "monthly"}, err
	case "aws_lb":
		price, err := costForELB(attributes, priceList)
		return &Cost{Value: price, Unit: "hourly"}, err
	case "aws_s3_bucket":
		// S3 pricing is per GB/month. For MVP, we use a flat $1.00/month baseline.
		return &Cost{Value: 1.0, Unit: "monthly"}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", rc.Type)
	}
}

// ... (costForEC2, costForRDS, etc. now return float64, error)
func costForEC2(attributes map[string]interface{}, priceList *pricing.PriceList) (float64, error) {
	instanceType, _ := attributes["instance_type"].(string)
	if instanceType == "" { return 0, fmt.Errorf("missing instance_type") }

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode == "AmazonEC2" && attr.InstanceType == instanceType && attr.Location == "US East (N. Virginia)" && attr.OperatingSystem == "Linux" && strings.HasPrefix(attr.UsageType, "BoxUsage") {
			return getPriceFromTerms(sku, priceList)
		}
	}
	return 0, fmt.Errorf("could not find pricing for EC2 instance type: %s", instanceType)
}

func costForRDS(attributes map[string]interface{}, priceList *pricing.PriceList) (float64, error) {
	instanceClass, _ := attributes["instance_class"].(string)
	if instanceClass == "" { return 0, fmt.Errorf("missing instance_class") }

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode == "AmazonRDS" && attr.InstanceClass == instanceClass && attr.Location == "US East (N. Virginia)" {
			return getPriceFromTerms(sku, priceList)
		}
	}
	return 0, fmt.Errorf("could not find pricing for RDS instance class: %s", instanceClass)
}

func costForEBS(attributes map[string]interface{}, priceList *pricing.PriceList) (float64, error) {
	volumeType, _ := attributes["type"].(string)
	if volumeType == "" { volumeType = "gp2" }
	size, _ := attributes["size"].(float64)

	apiName := "gp2"
	if volumeType == "gp3" { apiName = "gp3" }

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode == "AmazonEC2" && attr.VolumeAPIName == apiName && attr.Location == "US East (N. Virginia)" {
			price, err := getPriceFromTerms(sku, priceList)
			if err != nil { continue }
			return price * size, nil
		}
	}
	return 0, fmt.Errorf("could not find pricing for EBS volume type: %s", volumeType)
}

func costForELB(attributes map[string]interface{}, priceList *pricing.PriceList) (float64, error) {
	lbType, _ := attributes["load_balancer_type"].(string)
	if lbType == "" { lbType = "application" }

	group := ""
	if lbType == "application" { group = "ELB-Application" }

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode == "AWSELB" && attr.Group == group && attr.Location == "US East (N. Virginia)" {
			return getPriceFromTerms(sku, priceList)
		}
	}
	return 0, fmt.Errorf("could not find pricing for Load Balancer type: %s", lbType)
}


func getPriceFromTerms(sku string, priceList *pricing.PriceList) (float64, error) {
	if terms, ok := priceList.Terms.OnDemand[sku]; ok {
		for _, term := range terms {
			for _, dim := range term.PriceDimensions {
				price, err := strconv.ParseFloat(dim.PricePerUnit.USD, 64)
				if err == nil {
					return price, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("could not extract price for SKU %s", sku)
}
