CREATE TABLE IF NOT EXISTS transactions (
                                            id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    order_id    UUID REFERENCES orders(id),
    amount      DECIMAL(10,2) NOT NULL,
    type        VARCHAR(20) NOT NULL CHECK (type IN ('debit', 'credit')),
    description VARCHAR(255) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_order_id ON transactions(order_id);