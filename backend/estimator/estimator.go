package estimator

import (
	"fmt"
	"strconv"
	"strings"
	"cloudcostguard/backend/pricing"
	"cloudcostguard/backend/terraform"
)

// Cost represents a monetary cost with a value, a unit and a breakdown.
type Cost struct {
	Value    float64
	Unit     string // "hourly" or "monthly"
	Breakdown string
}

// Estimate calculates the estimated monthly cost impact of a Terraform plan.
// It iterates through the resource changes in the plan, estimates the cost of each change,
// and aggregates them into a total monthly cost.
//
// Parameters:
//   plan: The Terraform plan to estimate the cost of.
//   priceList: The list of AWS prices to use for the estimation.
//
// Returns:
//   A detailed breakdown of the estimated monthly cost impact.
func Estimate(plan *terraform.Plan, priceList *pricing.PriceList, region string, usage *UsageEstimates) (*EstimationResponse, error) {
	location := toLocation(region)
	response := &EstimationResponse{
		Currency:  "USD",
		Resources: []ResourceCost{},
	}

	for _, rc := range plan.ResourceChanges {
		cost, err := estimateResourceChange(rc, priceList, location, usage, plan)
		if err != nil {
			fmt.Printf("Warning: skipping unsupported resource %s: %v\n", rc.Address, err)
			continue
		}

		monthlyCost := cost.Value
		if cost.Unit == "hourly" {
			monthlyCost *= 730
		}

		if monthlyCost != 0 {
			response.Resources = append(response.Resources, ResourceCost{
				Address:      rc.Address,
				MonthlyCost:  monthlyCost,
				CostBreakdown: cost.Breakdown,
			})
		}
	}

	for _, resource := range response.Resources {
		response.TotalMonthlyCost += resource.MonthlyCost
	}

	response.Recommendations = GenerateRecommendations(plan, usage)

	return response, nil
}

func estimateResourceChange(rc *terraform.ResourceChange, priceList *pricing.PriceList, region string, usage *UsageEstimates, plan *terraform.Plan) (*Cost, error) {
	costChange := &Cost{Value: 0, Unit: "monthly"} // Default to monthly for aggregation
	actions := rc.Change.Actions
	isCreate := len(actions) == 1 && actions[0] == "create"
	isDelete := len(actions) == 1 && actions[0] == "delete"
	isUpdate := (len(actions) == 1 && actions[0] == "update") || (len(actions) == 2 && actions[0] == "delete" && actions[1] == "create")

	if isCreate || isUpdate {
		cost, err := getResourceCost(rc, rc.After, priceList, region, usage, plan)
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
		cost, err := getResourceCost(rc, rc.Before, priceList, region, usage, plan)
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

func getResourceCost(rc *terraform.ResourceChange, attributes map[string]interface{}, priceList *pricing.PriceList, region string, usage *UsageEstimates, plan *terraform.Plan) (*Cost, error) {
	if priceList == nil {
		return nil, fmt.Errorf("pricing data is nil")
	}

	switch rc.Type {
	case "aws_instance":
		return costForEC2(attributes, priceList, region)
	case "aws_db_instance":
		price, err := costForRDS(attributes, priceList, region)
		return &Cost{Value: price, Unit: "hourly"}, err
	case "aws_ebs_volume":
		price, err := costForEBS(attributes, priceList, region)
		return &Cost{Value: price, Unit: "monthly"}, err
	case "aws_lb":
		price, err := costForELB(attributes, priceList, region)
		return &Cost{Value: price, Unit: "hourly"}, err
	case "aws_s3_bucket":
		return costForS3(attributes, priceList, region, usage)
	case "aws_nat_gateway":
		price, err := costForNATGateway(attributes, priceList, region, usage)
		return &Cost{Value: price, Unit: "hourly"}, err
	case "aws_lambda_function":
		return costForLambda(attributes, priceList, region, usage)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", rc.Type)
	}
}

func costForNATGateway(attributes map[string]interface{}, priceList *pricing.PriceList, region string, usage *UsageEstimates) (float64, error) {
	var hourlyPrice, dataProcessingPrice float64
	var err error

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode != "AmazonVPC" || attr.Location != region {
			continue
		}

		if attr.Group == "NAT Gateway" {
			hourlyPrice, err = getPriceFromTerms(sku, priceList)
			if err != nil {
				return 0, fmt.Errorf("could not get hourly price for NAT Gateway: %w", err)
			}
		}

		if strings.Contains(attr.UsageType, "NatGateway-Bytes") {
			dataProcessingPrice, err = getPriceFromTerms(sku, priceList)
			if err != nil {
				return 0, fmt.Errorf("could not get data processing price for NAT Gateway: %w", err)
			}
		}
	}

	if hourlyPrice == 0 {
		return 0, fmt.Errorf("could not find hourly pricing for NAT Gateway")
	}

	totalHourlyCost := hourlyPrice
	if dataProcessingPrice > 0 && usage != nil {
		// Convert monthly GB processed to hourly GB processed
		hourlyGBProcessed := float64(usage.NATGatewayGBProcessed) / 730
		totalHourlyCost += hourlyGBProcessed * dataProcessingPrice
	}

	return totalHourlyCost, nil
}

// ... (costForEC2, costForRDS, etc. now return float64, error)
func costForEC2(attributes map[string]interface{}, priceList *pricing.PriceList, region string) (*Cost, error) {
	instanceType, _ := attributes["instance_type"].(string)
	if instanceType == "" {
		return nil, fmt.Errorf("missing instance_type")
	}

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode == "AmazonEC2" && attr.InstanceType == instanceType && attr.Location == region && attr.OperatingSystem == "Linux" && strings.HasPrefix(attr.UsageType, "BoxUsage") {
			price, err := getPriceFromTerms(sku, priceList)
			if err != nil {
				return nil, err
			}
			return &Cost{
				Value:    price,
				Unit:     "hourly",
				Breakdown: fmt.Sprintf("%s @ $%.4f/hr", instanceType, price),
			}, nil
		}
	}
	return nil, fmt.Errorf("could not find pricing for EC2 instance type: %s", instanceType)
}

func costForRDS(attributes map[string]interface{}, priceList *pricing.PriceList, region string) (float64, error) {
	instanceClass, _ := attributes["instance_class"].(string)
	if instanceClass == "" { return 0, fmt.Errorf("missing instance_class") }

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode == "AmazonRDS" && attr.InstanceClass == instanceClass && attr.Location == region {
			return getPriceFromTerms(sku, priceList)
		}
	}
	return 0, fmt.Errorf("could not find pricing for RDS instance class: %s", instanceClass)
}

func costForEBS(attributes map[string]interface{}, priceList *pricing.PriceList, region string) (float64, error) {
	volumeType, _ := attributes["type"].(string)
	if volumeType == "" { volumeType = "gp2" }
	size, _ := attributes["size"].(float64)

	apiName := "gp2"
	if volumeType == "gp3" { apiName = "gp3" }

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode == "AmazonEC2" && attr.VolumeAPIName == apiName && attr.Location == region {
			price, err := getPriceFromTerms(sku, priceList)
			if err != nil { continue }
			return price * size, nil
		}
	}
	return 0, fmt.Errorf("could not find pricing for EBS volume type: %s", volumeType)
}

func costForELB(attributes map[string]interface{}, priceList *pricing.PriceList, region string) (float64, error) {
	lbType, _ := attributes["load_balancer_type"].(string)
	if lbType == "" { lbType = "application" }

	group := ""
	if lbType == "application" { group = "ELB-Application" }

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode == "AWSELB" && attr.Group == group && attr.Location == region {
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

func costForLambda(attributes map[string]interface{}, priceList *pricing.PriceList, region string, usage *UsageEstimates) (*Cost, error) {
	memorySize, _ := attributes["memory_size"].(float64)
	if memorySize == 0 {
		memorySize = 128 // Default memory size
	}

	var requestPrice, gbSecondPrice float64
	var err error

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode != "AWSLambda" || attr.Location != region {
			continue
		}

		if strings.Contains(attr.UsageType, "Request") {
			requestPrice, err = getPriceFromTerms(sku, priceList)
			if err != nil {
				return nil, fmt.Errorf("could not get request price for Lambda: %w", err)
			}
		}

		if strings.Contains(attr.UsageType, "GB-Second") {
			gbSecondPrice, err = getPriceFromTerms(sku, priceList)
			if err != nil {
				return nil, fmt.Errorf("could not get GB-Second price for Lambda: %w", err)
			}
		}
	}

	if requestPrice == 0 || gbSecondPrice == 0 {
		return nil, fmt.Errorf("could not find pricing for Lambda")
	}

	totalMonthlyCost := 0.0
	breakdown := "No usage data provided"
	if usage != nil {
		// Free tier adjustment
		monthlyRequests := float64(usage.LambdaMonthlyRequests)
		gbSeconds := (memorySize / 1024) * (float64(usage.LambdaAvgDurationMS) / 1000) * monthlyRequests

		requestCost := (monthlyRequests - 1000000) * requestPrice
		if requestCost < 0 {
			requestCost = 0
		}

		gbSecondCost := (gbSeconds - 400000) * gbSecondPrice
		if gbSecondCost < 0 {
			gbSecondCost = 0
		}

		totalMonthlyCost = requestCost + gbSecondCost
		breakdown = fmt.Sprintf("%d requests/month @ %dms avg duration", usage.LambdaMonthlyRequests, usage.LambdaAvgDurationMS)
	}

	return &Cost{
		Value:    totalMonthlyCost,
		Unit:     "monthly",
		Breakdown: breakdown,
	}, nil
}
func costForS3(attributes map[string]interface{}, priceList *pricing.PriceList, region string, usage *UsageEstimates) (*Cost, error) {
	var storagePrice, putRequestPrice float64
	var err error

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode != "AmazonS3" || attr.Location != region {
			continue
		}

		if attr.StorageClass == "General Purpose" && strings.Contains(attr.UsageType, "TimedStorage-ByteHrs") {
			storagePrice, err = getPriceFromTerms(sku, priceList)
			if err != nil {
				return nil, fmt.Errorf("could not get storage price for S3: %w", err)
			}
		}

		if attr.Group == "S3-Request-Tier1" {
			putRequestPrice, err = getPriceFromTerms(sku, priceList)
			if err != nil {
				return nil, fmt.Errorf("could not get put request price for S3: %w", err)
			}
		}
	}

	if storagePrice == 0 || putRequestPrice == 0 {
		return nil, fmt.Errorf("could not find pricing for S3")
	}

	totalMonthlyCost := 0.0
	breakdown := "No usage data provided"
	if usage != nil {
		storageCost := float64(usage.S3StorageGB) * storagePrice
		requestCost := (float64(usage.S3MonthlyPutRequests) / 1000) * putRequestPrice
		totalMonthlyCost = storageCost + requestCost
		breakdown = fmt.Sprintf("%d GB storage @ $%.4f/GB + %d PUT requests @ $%.4f/1000", usage.S3StorageGB, storagePrice, usage.S3MonthlyPutRequests, putRequestPrice)
	}

	return &Cost{
		Value:    totalMonthlyCost,
		Unit:     "monthly",
		Breakdown: breakdown,
	}, nil
}
