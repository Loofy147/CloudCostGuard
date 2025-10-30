package pricing

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"cloudcostguard/backend/pricing"
	"time"

	"go.uber.org/zap"
)

var awsPricingURLs = []string{
	"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/index.json",
	"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonRDS/current/index.json",
	"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonS3/current/index.json",
	"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AWSELB/current/index.json",
	"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEBS/current/index.json",
	"https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonVPC/current/index.json",
}

type PricingDataStorer interface {
	StorePricingData(ctx context.Context, priceList *pricing.PriceList) error
}

type Service struct {
	logger *zap.Logger
	storer PricingDataStorer
}

func NewService(logger *zap.Logger, storer PricingDataStorer) *Service {
	return &Service{
		logger: logger,
		storer: storer,
	}
}

func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Starting pricing service")
	go s.run(ctx)
}

func (s *Service) run(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		s.logger.Info("Fetching and storing pricing data")
		if err := s.fetchAndStorePricingData(ctx); err != nil {
			s.logger.Error("Failed to fetch and store pricing data", zap.Error(err))
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			s.logger.Info("Stopping pricing service")
			return
		}
	}
}

func (s *Service) fetchAndStorePricingData(ctx context.Context) error {
	priceList := pricing.NewPriceList()

	for _, url := range awsPricingURLs {
		s.logger.Info("Fetching pricing data", zap.String("url", url))
		if err := priceList.LoadFromURL(url); err != nil {
			return fmt.Errorf("failed to load pricing data from %s: %w", url, err)
		}
	}

	return s.storer.StorePricingData(ctx, priceList)
}

type PostgresPricingDataStorer struct {
	db *sql.DB
}

func NewPostgresPricingDataStorer(db *sql.DB) *PostgresPricingDataStorer {
	return &PostgresPricingDataStorer{db: db}
}

func (s *PostgresPricingDataStorer) StorePricingData(ctx context.Context, priceList *pricing.PriceList) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO aws_prices (sku, product_json, terms_json, last_updated)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (sku) DO UPDATE SET
			product_json = EXCLUDED.product_json,
			terms_json = EXCLUDED.terms_json,
			last_updated = EXCLUDED.last_updated
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()

	for sku, product := range priceList.Products {
		productJSON, err := json.Marshal(product)
		if err != nil {
			return fmt.Errorf("failed to marshal product for SKU %s: %w", sku, err)
		}

		terms := priceList.Terms.OnDemand[sku]
		termsJSON, err := json.Marshal(terms)
		if err != nil {
			return fmt.Errorf("failed to marshal terms for SKU %s: %w", sku, err)
		}

		if _, err := stmt.ExecContext(ctx, sku, productJSON, termsJSON, now); err != nil {
			return fmt.Errorf("failed to execute statement for SKU %s: %w", sku, err)
		}
	}

	return tx.Commit()
}
