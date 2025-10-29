package pricing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// PriceList holds the pricing data for all supported AWS services.
type PriceList struct {
	// Products is a map of SKU to Product.
	Products map[string]Product `json:"products"`
	// Terms is a map of terms, with OnDemand being the only one we care about.
	Terms    struct {
		OnDemand map[string]map[string]Term `json:"OnDemand"`
	} `json:"terms"`
}

// Product represents a single product in the AWS catalog.
type Product struct {
	// SKU is the unique identifier for the product.
	SKU        string            `json:"sku"`
	// Attributes contains the detailed attributes of a product.
	Attributes ProductAttributes `json:"attributes"`
}

// ProductAttributes contains the detailed attributes of a product.
type ProductAttributes struct {
	// ServiceCode is the AWS service code (e.g., "AmazonEC2").
	ServiceCode     string `json:"servicecode"`
	// InstanceType is the EC2 instance type (e.g., "t2.micro").
	InstanceType    string `json:"instanceType"`
	// InstanceClass is the RDS instance class (e.g., "db.t2.micro").
	InstanceClass   string `json:"instanceClass"`
	// Location is the AWS region (e.g., "US East (N. Virginia)").
	Location        string `json:"location"`
	// OperatingSystem is the operating system (e.g., "Linux").
	OperatingSystem string `json:"operatingSystem"`
	// UsageType is the usage type (e.g., "BoxUsage:t2.micro").
	UsageType       string `json:"usagetype"`
	// VolumeAPIName is the EBS volume type (e.g., "gp2").
	VolumeAPIName   string `json:"volumeApiName"`
	// Group is the ELB group (e.g., "ELB-Application").
	Group           string `json:"group"`
}

// Term represents the pricing terms for a product.
type Term struct {
	// PriceDimensions is a map of price dimensions.
	PriceDimensions map[string]PriceDimension `json:"priceDimensions"`
}

// PriceDimension represents a single dimension of pricing for a product.
type PriceDimension struct {
	// PricePerUnit is a map of currency to price.
	PricePerUnit struct {
		USD string `json:"USD"`
	} `json:"pricePerUnit"`
}


// NewPriceList creates a new, empty price list.
//
// Returns:
//   A pointer to a new, empty PriceList.
func NewPriceList() *PriceList {
	return &PriceList{
		Products: make(map[string]Product),
	}
}

// LoadFromURL fetches a pricing file from a URL and merges it into the price list.
//
// Parameters:
//   url: The URL of the pricing file to load.
//
// Returns:
//   An error if the pricing file could not be loaded or parsed, nil otherwise.
func (p *PriceList) LoadFromURL(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download price list from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download price list: received status %s", resp.Status)
	}

	var newPricing PriceList
	if err := json.NewDecoder(resp.Body).Decode(&newPricing); err != nil {
		return fmt.Errorf("failed to parse price list: %w", err)
	}

	p.Merge(&newPricing)
	return nil
}


// LoadFromFile loads a pricing file from a local path and merges it into the price list.
//
// Parameters:
//   path: The local path of the pricing file to load.
//
// Returns:
//   An error if the pricing file could not be loaded or parsed, nil otherwise.
func (p *PriceList) LoadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open pricing file: %w", err)
	}
	defer file.Close()

	var newPricing PriceList
	if err := json.NewDecoder(file).Decode(&newPricing); err != nil {
		return fmt.Errorf("failed to parse pricing file: %w", err)
	}

	p.Merge(&newPricing)
	return nil
}

// Merge merges another price list into the current one.
//
// Parameters:
//   other: The price list to merge into the current one.
func (p *PriceList) Merge(other *PriceList) {
	for sku, product := range other.Products {
		p.Products[sku] = product
	}
	if p.Terms.OnDemand == nil {
		p.Terms.OnDemand = make(map[string]map[string]Term)
	}
	for sku, terms := range other.Terms.OnDemand {
		p.Terms.OnDemand[sku] = terms
	}
}
