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
		// S3 pricing is per GB/month. For MVP, we use a flat $1.00/month baseline.
		return &Cost{Value: 1.0, Unit: "monthly"}, nil
	case "aws_nat_gateway":
		price, err := costForNATGateway(attributes, priceList, region, usage)
		return &Cost{Value: price, Unit: "hourly"}, err
	case "aws_lambda_function":
		return costForLambda(attributes, priceList, region, usage)
	case "aws_ecs_service":
		return costForECSService(rc, attributes, priceList, region, plan)
	case "aws_ecs_task_definition":
		// Cost is calculated as part of the ECS service, not standalone.
		return &Cost{Value: 0, Unit: "monthly"}, nil
	case "aws_eks_cluster":
		return costForEKS(attributes, priceList, region)
	case "aws_eks_node_group":
		return costForEKSNodeGroup(attributes, priceList, region)
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
	}

	return &Cost{
		Value:    totalMonthlyCost,
		Unit:     "monthly",
		Breakdown: fmt.Sprintf("%d requests/month @ %dms avg duration", usage.LambdaMonthlyRequests, usage.LambdaAvgDurationMS),
	}, nil
}

func costForECSService(rc *terraform.ResourceChange, attributes map[string]interface{}, priceList *pricing.PriceList, region string, plan *terraform.Plan) (*Cost, error) {
	launchType, _ := attributes["launch_type"].(string)
	if launchType != "FARGATE" {
		return &Cost{Value: 0, Unit: "monthly"}, nil
	}

	desiredCount, _ := attributes["desired_count"].(float64)
	if desiredCount == 0 {
		desiredCount = 1
	}

	taskDefinitionArn, _ := attributes["task_definition"].(string)
	if taskDefinitionArn == "" {
		return nil, fmt.Errorf("missing task_definition for Fargate service")
	}

	var taskDef *terraform.ResourceChange
	for _, r := range plan.ResourceChanges {
		if r.Address == taskDefinitionArn {
			taskDef = r
			break
		}
	}

	if taskDef == nil {
		return nil, fmt.Errorf("could not find task definition: %s", taskDefinitionArn)
	}

	cpu, err := parseFloat(taskDef.After["cpu"])
	if err != nil {
		return nil, fmt.Errorf("could not parse cpu from task definition: %w", err)
	}

	memory, err := parseFloat(taskDef.After["memory"])
	if err != nil {
		return nil, fmt.Errorf("could not parse memory from task definition: %w", err)
	}

	var vcpuPrice, memoryPrice float64

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode != "AmazonECS" || attr.Location != region {
			continue
		}

		if strings.Contains(attr.UsageType, "vCPU-Hours") {
			vcpuPrice, err = getPriceFromTerms(sku, priceList)
			if err != nil {
				return nil, fmt.Errorf("could not get vCPU price for Fargate: %w", err)
			}
		}

		if strings.Contains(attr.UsageType, "GB-Hours") {
			memoryPrice, err = getPriceFromTerms(sku, priceList)
			if err != nil {
				return nil, fmt.Errorf("could not get memory price for Fargate: %w", err)
			}
		}
	}

	if vcpuPrice == 0 || memoryPrice == 0 {
		return nil, fmt.Errorf("could not find pricing for Fargate")
	}

	hourlyCost := (cpu/1024)*vcpuPrice + (memory/1024)*memoryPrice
	totalHourlyCost := hourlyCost * desiredCount

	return &Cost{
		Value:    totalHourlyCost,
		Unit:     "hourly",
		Breakdown: fmt.Sprintf("%d tasks @ %.2f vCPU / %.2f GB", int(desiredCount), cpu/1024, memory/1024),
	}, nil
}

func parseFloat(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("unsupported type for float conversion")
	}
}

func costForEKS(attributes map[string]interface{}, priceList *pricing.PriceList, region string) (*Cost, error) {
	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode == "AmazonEKS" && attr.Location == region && strings.Contains(attr.UsageType, "EKS-Hours:perCluster") {
			price, err := getPriceFromTerms(sku, priceList)
			if err != nil {
				return nil, err
			}
			return &Cost{
				Value:    price,
				Unit:     "hourly",
				Breakdown: fmt.Sprintf("EKS Control Plane @ $%.4f/hr", price),
			}, nil
		}
	}
	return nil, fmt.Errorf("could not find pricing for EKS control plane in region: %s", region)
}

func costForEKSNodeGroup(attributes map[string]interface{}, priceList *pricing.PriceList, region string) (*Cost, error) {
	instanceTypes, ok := attributes["instance_types"].([]interface{})
	if !ok || len(instanceTypes) == 0 {
		return nil, fmt.Errorf("missing instance_types")
	}
	instanceType := instanceTypes[0].(string)

	scalingConfig, ok := attributes["scaling_config"].([]interface{})
	if !ok || len(scalingConfig) == 0 {
		return nil, fmt.Errorf("missing scaling_config")
	}
	desiredSize, ok := scalingConfig[0].(map[string]interface{})["desired_size"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing desired_size")
	}

	ec2Cost, err := costForEC2(map[string]interface{}{"instance_type": instanceType}, priceList, region)
	if err != nil {
		return nil, err
	}

	totalCost := ec2Cost.Value * desiredSize
	return &Cost{
		Value:    totalCost,
		Unit:     "hourly",
		Breakdown: fmt.Sprintf("%d x %s @ $%.4f/hr", int(desiredSize), instanceType, ec2Cost.Value),
	}, nil
}
