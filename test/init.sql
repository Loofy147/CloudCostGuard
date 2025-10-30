CREATE TABLE aws_prices (
    sku TEXT PRIMARY KEY,
    product_json JSONB,
    terms_json JSONB,
    last_updated TIMESTAMPTZ NOT NULL
);
