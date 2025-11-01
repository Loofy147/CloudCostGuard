package api_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"cloudcostguard/backend/internal/api"
	"cloudcostguard/backend/internal/cache"
	"cloudcostguard/backend/internal/config"
	"cloudcostguard/backend/internal/repository/postgres"
	"cloudcostguard/backend/internal/service"
	"github.com/ory/dockertest/v3"
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var db *sql.DB

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.BuildAndRun("postgres-test", "../../test/Dockerfile.postgres", []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=test"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("postgres", fmt.Sprintf("postgres://postgres:secret@localhost:%s/test?sslmode=disable", resource.GetPort("5432/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestEstimateHandler_Integration(t *testing.T) {
	logger, _ := zap.NewProduction()
	apiConfig := config.APIConfig{
		APIKeys: []string{"test-key"},
	}

	pricingRepo := postgres.NewPricingRepository(db, logger)
	pricingCache := cache.NewPricingCache(pricingRepo, logger, time.Hour)
	estimatorSvc := service.NewEstimator(pricingCache, logger, db)
	router := api.NewRouter(estimatorSvc, logger, db, pricingCache, apiConfig)

	// Test with no data in DB
	req1, _ := http.NewRequest("POST", "/estimate", bytes.NewBufferString(`{"plan": {"resource_changes": []}}`))
	req1.Header.Set("Authorization", "Bearer test-key")
	rr1 := httptest.NewRecorder()
	router.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusServiceUnavailable, rr1.Code)

	// Add some data to the DB
	_, err := db.Exec(`CREATE TABLE aws_prices (
		sku TEXT PRIMARY KEY,
		product_json JSONB,
		terms_json JSONB,
		last_updated TIMESTAMPTZ NOT NULL
	);`)
	assert.NoError(t, err)
	_, err = db.Exec(`INSERT INTO aws_prices (sku, product_json, terms_json, last_updated) VALUES ('test-sku', '{}', '{}', NOW())`)
	assert.NoError(t, err)

	// Refresh cache
	err = pricingCache.Refresh(context.Background())
	assert.NoError(t, err)

	// Test with data in DB
	req2, _ := http.NewRequest("POST", "/estimate", bytes.NewBufferString(`{"plan": {"resource_changes": []}}`))
	req2.Header.Set("Authorization", "Bearer test-key")
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)
}
