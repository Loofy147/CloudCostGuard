package cache

import (
    "context"
    "sync"
    "time"

    "cloudcostguard/backend/pricing"
    "go.uber.org/zap"
)

type PricingRepository interface {
    LoadPricing(ctx context.Context) (*pricing.PriceList, error)
}

type PricingCache struct {
    mu       sync.RWMutex
    data     *pricing.PriceList
    repo     PricingRepository
    logger   *zap.Logger
    refreshInterval time.Duration
    stopChan chan struct{}
}

func NewPricingCache(repo PricingRepository, logger *zap.Logger, refreshInterval time.Duration) *PricingCache {
    pc := &PricingCache{
        repo:     repo,
        logger:   logger,
        refreshInterval: refreshInterval,
        stopChan: make(chan struct{}),
    }

    // Initial load
    if err := pc.Refresh(context.Background()); err != nil {
        logger.Error("Failed initial cache load", zap.Error(err))
    }

    // Start background refresh
    go pc.backgroundRefresh()

    return pc
}

func (pc *PricingCache) Get() *pricing.PriceList {
    pc.mu.RLock()
    defer pc.mu.RUnlock()
    return pc.data
}

func (pc *PricingCache) Refresh(ctx context.Context) error {
    priceList, err := pc.repo.LoadPricing(ctx)
    if err != nil {
        return err
    }

    pc.mu.Lock()
    pc.data = priceList
    pc.mu.Unlock()

    pc.logger.Info("Pricing cache refreshed")

    return nil
}

func (pc *PricingCache) backgroundRefresh() {
    ticker := time.NewTicker(pc.refreshInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
            if err := pc.Refresh(ctx); err != nil {
                pc.logger.Error("Background refresh failed", zap.Error(err))
            }
            cancel()
        case <-pc.stopChan:
            return
        }
    }
}

func (pc *PricingCache) Stop() {
    close(pc.stopChan)
}
