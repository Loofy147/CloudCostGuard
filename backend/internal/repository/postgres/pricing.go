package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"cloudcostguard/backend/pricing"
)

type PricingRepository struct {
	db *sql.DB
}

func NewPricingRepository(db *sql.DB) *PricingRepository {
	return &PricingRepository{db: db}
}

func (r *PricingRepository) LoadPricing(ctx context.Context) (*pricing.PriceList, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT sku, product_json, terms_json FROM aws_prices")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	newPriceList := pricing.NewPriceList()
	for rows.Next() {
		var sku string
		var productJSON, termsJSON []byte
		if err := rows.Scan(&sku, &productJSON, &termsJSON); err != nil {
			return nil, err
		}

		var product pricing.Product
		if err := json.Unmarshal(productJSON, &product); err != nil {
			fmt.Printf("Warning: could not unmarshal product for SKU %s: %v\n", sku, err)
			continue
		}
		newPriceList.Products[sku] = product

		var terms map[string]pricing.Term
		if err := json.Unmarshal(termsJSON, &terms); err != nil {
			fmt.Printf("Warning: could not unmarshal terms for SKU %s: %v\n", sku, err)
			continue
		}
		newPriceList.Terms.OnDemand[sku] = terms
	}

	return newPriceList, nil
}
