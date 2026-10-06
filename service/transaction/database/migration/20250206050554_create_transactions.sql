-- +goose Up
-- +goose StatementBegin
-- order_id stays inside pos_sales (FK kept). merchant_id points at
-- pos_merchant, so it is validated by the merchant adapter over gRPC instead
-- of a cross-database foreign key.
CREATE TABLE "transactions" (
    "transaction_id" SERIAL PRIMARY KEY,
    "order_id" INT NOT NULL REFERENCES "orders" ("order_id"),
    "merchant_id" INT NOT NULL,
    "payment_method" VARCHAR(50) NOT NULL,
    "amount" INT NOT NULL,
    "change_amount" INT DEFAULT 0,
    "payment_status" VARCHAR(20) NOT NULL DEFAULT 'completed',
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_transactions_order_id ON transactions (order_id);

CREATE INDEX idx_transactions_payment_status ON transactions (payment_status);

CREATE INDEX idx_transactions_created_at ON transactions (created_at);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_transactions_order_id ON transactions (order_id);

DROP INDEX IF EXISTS idx_transactions_payment_status ON transactions (payment_status);

DROP INDEX IF EXISTS idx_transactions_created_at ON transactions (created_at);

DROP TABLE IF EXISTS "transactions";

-- +goose StatementEnd
