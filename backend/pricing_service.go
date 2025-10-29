package main

import (
	"cloudcostguard/backend/database"
	"cloudcostguard/backend/pricing"
	"encoding/json"
	"fmt"
	"time"
)

func startPricingService() {
	fmt.Println("Starting pricing service...")
	go func() {
		for {
			updatePricingData()
			// Update once every 24 hours
			time.Sleep(24 * time.Hour)
		}
	}()
}

func updatePricingData() {
	fmt.Println("Updating AWS pricing data...")
	urls := []string{
		"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/index.json",
		"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonRDS/current/index.json",
		"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonS3/current/index.json",
		"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonElasticLoadBalancing/current/index.json",
		"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonVPC/current/index.json",
	}

	for _, url := range urls {
		priceList := pricing.NewPriceList()
		if err := priceList.LoadFromURL(url); err != nil {
			fmt.Printf("Warning: could not load pricing data from %s: %v\n", url, err)
			continue
		}

		storePriceList(priceList)
	}
	fmt.Println("Finished updating AWS pricing data.")
}

func storePriceList(priceList *pricing.PriceList) {
	for sku, product := range priceList.Products {
		productJSON, err := json.Marshal(product)
		if err != nil {
			fmt.Printf("Warning: failed to marshal product for SKU %s: %v\n", sku, err)
			continue
		}
		termsJSON, err := json.Marshal(priceList.Terms.OnDemand[sku])
		if err != nil {
			fmt.Printf("Warning: failed to marshal terms for SKU %s: %v\n", sku, err)
			continue
		}

		_, err = database.DB.Exec(`
			INSERT INTO aws_prices (sku, product_json, terms_json, last_updated)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (sku) DO UPDATE
			SET product_json = $2, terms_json = $3, last_updated = NOW();
		`, sku, productJSON, termsJSON)

		if err != nil {
			fmt.Printf("Warning: failed to insert/update SKU %s: %v\n", sku, err)
		}
	}
}
