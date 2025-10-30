package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"cloudcostguard/backend/pricing"
	"go.uber.org/zap"
)

type PricingRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPricingRepository(db *sql.DB, logger *zap.Logger) *PricingRepository {
	return &PricingRepository{db: db, logger: logger}
}

func (r *PricingRepository) LoadPricing(ctx context.Context) (*pricing.PriceList, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT sku, product_json, terms_json FROM aws_prices")
	if err != nil {
		return nil, fmt.Errorf("cache refresh failed: %w", err)
	}
	defer rows.Close()

	newPriceList := pricing.NewPriceList()
	newPriceList.Terms.OnDemand = make(map[string]map[string]pricing.Term)

	for rows.Next() {
		var sku string
		var productJSON, termsJSON []byte
		if err := rows.Scan(&sku, &productJSON, &termsJSON); err != nil {
			r.logger.Warn("Failed to scan row", zap.Error(err))
			continue
		}

		var product pricing.Product
		if err := json.Unmarshal(productJSON, &product); err != nil {
			r.logger.Warn("Failed to unmarshal product", zap.String("sku", sku), zap.Error(err))
			continue
		}
		newPriceList.Products[sku] = product

		var terms map[string]pricing.Term
		if err := json.Unmarshal(termsJSON, &terms); err != nil {
			r.logger.Warn("Failed to unmarshal terms", zap.String("sku", sku), zap.Error(err))
			continue
		}
		newPriceList.Terms.OnDemand[sku] = terms
	}

	r.logger.Info("Cache refreshed", zap.Int("products_loaded", len(newPriceList.Products)))
	return newPriceList, nil
}
