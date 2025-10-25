package pricing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// PriceList holds the pricing data for all supported AWS services.
type PriceList struct {
	Products map[string]Product `json:"products"`
	Terms    struct {
		OnDemand map[string]map[string]Term `json:"OnDemand"`
	} `json:"terms"`
}

// Product represents a single product in the AWS catalog.
type Product struct {
	SKU        string            `json:"sku"`
	Attributes ProductAttributes `json:"attributes"`
}

// ProductAttributes contains the detailed attributes of a product.
type ProductAttributes struct {
	ServiceCode     string `json:"servicecode"`
	InstanceType    string `json:"instanceType"`
	InstanceClass   string `json:"instanceClass"`
	Location        string `json:"location"`
	OperatingSystem string `json:"operatingSystem"`
	UsageType       string `json:"usagetype"`
	VolumeAPIName   string `json:"volumeApiName"`
	Group           string `json:"group"`
}

// Term represents the pricing terms for a product.
type Term struct {
	PriceDimensions map[string]PriceDimension `json:"priceDimensions"`
}

// PriceDimension represents a single dimension of pricing for a product.
type PriceDimension struct {
	PricePerUnit struct {
		USD string `json:"USD"`
	} `json:"pricePerUnit"`
}


// NewPriceList creates a new, empty price list.
func NewPriceList() *PriceList {
	return &PriceList{
		Products: make(map[string]Product),
	}
}

// LoadFromURL fetches a pricing file from a URL and merges it into the price list.
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
