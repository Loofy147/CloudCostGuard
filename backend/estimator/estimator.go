// Package estimator provides the core logic for estimating the cost of Terraform plans.
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
//   region: The AWS region to use for pricing.
//   usage: A struct containing usage estimates for various resources.
//
// Returns:
//   A pointer to an EstimationResponse struct containing a detailed breakdown of the estimated monthly cost impact.
//   An error if the estimation fails.
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

// estimateResourceChange calculates the cost impact of a single resource change.
// It determines whether the resource is being created, deleted, or updated, and calculates the cost delta accordingly.
//
// Parameters:
//   rc: The resource change to estimate the cost of.
//   priceList: The list of AWS prices to use for the estimation.
//   region: The AWS region to use for pricing.
//   usage: A struct containing usage estimates for various resources.
//   plan: The full Terraform plan.
//
// Returns:
//   A pointer to a Cost struct representing the cost delta of the resource change.
//   An error if the estimation fails.
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

// getResourceCost calculates the cost of a single resource based on its attributes.
// It delegates to the appropriate cost function based on the resource type.
//
// Parameters:
//   rc: The resource change to get the cost of.
//   attributes: The attributes of the resource.
//   priceList: The list of AWS prices to use for the estimation.
//   region: The AWS region to use for pricing.
//   usage: A struct containing usage estimates for various resources.
//   plan: The full Terraform plan.
//
// Returns:
//   A pointer to a Cost struct representing the cost of the resource.
//   An error if the estimation fails.
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
	case "aws_ecs_service":
		return costForECSService(rc, attributes, priceList, region, plan)
	case "aws_ecs_task_definition":
		// Cost is calculated as part of the ECS service, not standalone.
		return &Cost{Value: 0, Unit: "monthly"}, nil
	case "aws_eks_cluster":
		return costForEKS(attributes, priceList, region)
	case "aws_eks_node_group":
		return costForEKSNodeGroup(attributes, priceList, region)
	case "aws_elasticache_cluster":
		return costForElastiCache(attributes, priceList, region)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", rc.Type)
	}
}

// costForNATGateway calculates the cost of an AWS NAT Gateway.
// It includes both the hourly price and the data processing price.
//
// Parameters:
//   attributes: The attributes of the NAT Gateway resource.
//   priceList: The list of AWS prices.
//   region: The AWS region.
//   usage: Usage estimates, which may include NAT Gateway data processing volume.
//
// Returns:
//   The estimated hourly cost of the NAT Gateway.
//   An error if the pricing data cannot be found.
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

// costForEC2 calculates the cost of an AWS EC2 instance.
//
// Parameters:
//   attributes: The attributes of the EC2 instance resource.
//   priceList: The list of AWS prices.
//   region: The AWS region.
//
// Returns:
//   A pointer to a Cost struct representing the hourly cost of the EC2 instance.
//   An error if the pricing data cannot be found.
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

// costForRDS calculates the cost of an AWS RDS instance.
//
// Parameters:
//   attributes: The attributes of the RDS instance resource.
//   priceList: The list of AWS prices.
//   region: The AWS region.
//
// Returns:
//   The estimated hourly cost of the RDS instance.
//   An error if the pricing data cannot be found.
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

// costForEBS calculates the cost of an AWS EBS volume.
//
// Parameters:
//   attributes: The attributes of the EBS volume resource.
//   priceList: The list of AWS prices.
//   region: The AWS region.
//
// Returns:
//   The estimated monthly cost of the EBS volume.
//   An error if the pricing data cannot be found.
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

// costForELB calculates the cost of an AWS Elastic Load Balancer.
//
// Parameters:
//   attributes: The attributes of the ELB resource.
//   priceList: The list of AWS prices.
//   region: The AWS region.
//
// Returns:
//   The estimated hourly cost of the ELB.
//   An error if the pricing data cannot be found.
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

// getPriceFromTerms extracts the price from the terms of a product.
//
// Parameters:
//   sku: The SKU of the product.
//   priceList: The list of AWS prices.
//
// Returns:
//   The price of the product.
//   An error if the price cannot be extracted.
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

// costForLambda calculates the cost of an AWS Lambda function.
// It includes both the request price and the GB-second price, and accounts for the free tier.
//
// Parameters:
//   attributes: The attributes of the Lambda function resource.
//   priceList: The list of AWS prices.
//   region: The AWS region.
//   usage: Usage estimates, which may include Lambda monthly requests and average duration.
//
// Returns:
//   A pointer to a Cost struct representing the monthly cost of the Lambda function.
//   An error if the pricing data cannot be found.
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

// costForECSService calculates the cost of an AWS ECS service.
// It currently only supports the Fargate launch type.
//
// Parameters:
//   rc: The resource change for the ECS service.
//   attributes: The attributes of the ECS service resource.
//   priceList: The list of AWS prices.
//   region: The AWS region.
//   plan: The full Terraform plan.
//
// Returns:
//   A pointer to a Cost struct representing the hourly cost of the ECS service.
//   An error if the pricing data cannot be found.
func costForECSService(rc *terraform.ResourceChange, attributes map[string]interface{}, priceList *pricing.PriceList, region string, plan *terraform.Plan) (*Cost, error) {
	launchType, _ := attributes["launch_type"].(string)
	if launchType != "FARGATE" {
		// For EC2 launch type, cost is in the EC2 instances, not the service.
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

// parseFloat converts a value to a float64.
// It can handle float64 and string types.
//
// Parameters:
//   val: The value to convert.
//
// Returns:
//   The converted float64 value.
//   An error if the conversion fails.
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

// costForS3 calculates the cost of an AWS S3 bucket.
// It includes both storage and request costs.
//
// Parameters:
//   attributes: The attributes of the S3 bucket resource.
//   priceList: The list of AWS prices.
//   region: The AWS region.
//   usage: Usage estimates, which may include S3 storage and request data.
//
// Returns:
//   A pointer to a Cost struct representing the monthly cost of the S3 bucket.
//   An error if the pricing data cannot be found.
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

// costForEKS calculates the cost of an AWS EKS cluster.
// It includes the hourly price for the control plane.
//
// Parameters:
//   attributes: The attributes of the EKS cluster resource.
//   priceList: The list of AWS prices.
//   region: The AWS region.
//
// Returns:
//   A pointer to a Cost struct representing the hourly cost of the EKS cluster.
//   An error if the pricing data cannot be found.
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

// costForEKSNodeGroup calculates the cost of an AWS EKS node group.
// It calculates the cost of the EC2 instances in the node group.
//
// Parameters:
//   attributes: The attributes of the EKS node group resource.
//   priceList: The list of AWS prices.
//   region: The AWS region.
//
// Returns:
//   A pointer to a Cost struct representing the hourly cost of the EKS node group.
//   An error if the pricing data cannot be found.
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

// costForElastiCache calculates the cost of an AWS ElastiCache cluster.
//
// Parameters:
//   attributes: The attributes of the ElastiCache cluster resource.
//   priceList: The list of AWS prices.
//   region: The AWS region.
//
// Returns:
//   A pointer to a Cost struct representing the hourly cost of the ElastiCache cluster.
//   An error if the pricing data cannot be found.
func costForElastiCache(attributes map[string]interface{}, priceList *pricing.PriceList, region string) (*Cost, error) {
	nodeType, _ := attributes["node_type"].(string)
	if nodeType == "" {
		return nil, fmt.Errorf("missing node_type")
	}
	numCacheNodes, _ := attributes["num_cache_nodes"].(float64)
	if numCacheNodes == 0 {
		numCacheNodes = 1
	}

	for sku, product := range priceList.Products {
		attr := product.Attributes
		if attr.ServiceCode == "AmazonElastiCache" && attr.InstanceType == nodeType && attr.Location == region {
			price, err := getPriceFromTerms(sku, priceList)
			if err != nil {
				return nil, err
			}
			totalCost := price * numCacheNodes
			return &Cost{
				Value:    totalCost,
				Unit:     "hourly",
				Breakdown: fmt.Sprintf("%d x %s @ $%.4f/hr", int(numCacheNodes), nodeType, price),
			}, nil
		}
	}
	return nil, fmt.Errorf("could not find pricing for ElastiCache node type: %s", nodeType)
}
