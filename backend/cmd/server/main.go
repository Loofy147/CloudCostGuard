package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloudcostguard/backend/internal/api"
	"cloudcostguard/backend/internal/cache"
	"cloudcostguard/backend/internal/config"
	"cloudcostguard/backend/internal/repository/postgres"
	"cloudcostguard/backend/internal/service"
	"cloudcostguard/backend/internal/service/pricing"
	"database/sql"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}


	// Initialize dependencies
	db, err := postgres.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	createTable(db, logger)

	// Initialize services
	pricingRepo := postgres.NewPricingRepository(db, logger)
	pricingCache := cache.NewPricingCache(pricingRepo, logger, cfg.Cache.RefreshInterval)
	estimatorSvc := service.NewEstimator(pricingCache, logger)
	pricingStorer := pricing.NewPostgresPricingDataStorer(db)
	pricingSvc := pricing.NewService(logger, pricingStorer)
	pricingSvc.Start(context.Background())


	// Initialize HTTP server
	router := api.NewRouter(estimatorSvc, logger, db, pricingCache)
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func createTable(db *sql.DB, logger *zap.Logger) {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS aws_prices (
		sku TEXT PRIMARY KEY,
		product_json JSONB,
		terms_json JSONB,
		last_updated TIMESTAMPTZ NOT NULL
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		logger.Fatal("Failed to create prices table", zap.Error(err))
	}
}
