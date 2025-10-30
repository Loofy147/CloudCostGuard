CREATE TABLE IF NOT EXISTS aws_prices (
    sku TEXT PRIMARY KEY,
    product_json JSONB NOT NULL,
    terms_json JSONB NOT NULL,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_aws_prices_last_updated ON aws_prices(last_updated);
CREATE INDEX IF NOT EXISTS idx_aws_prices_product_service ON aws_prices USING GIN ((product_json->'attributes'->>'servicecode'));
CREATE INDEX IF NOT EXISTS idx_aws_prices_product_location ON aws_prices USING GIN ((product_json->'attributes'->>'location'));
