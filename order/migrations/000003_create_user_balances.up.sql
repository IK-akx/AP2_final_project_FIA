CREATE TABLE IF NOT EXISTS user_balances (
    user_id     UUID PRIMARY KEY,
    balance     DECIMAL(10,2) NOT NULL DEFAULT 1000.00,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );