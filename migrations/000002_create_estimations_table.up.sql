CREATE TABLE estimations (
    id SERIAL PRIMARY KEY,
    repository VARCHAR(255) NOT NULL,
    pr_number INTEGER NOT NULL,
    total_monthly_cost NUMERIC(10, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
