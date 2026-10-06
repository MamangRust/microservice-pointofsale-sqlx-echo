-- +goose Up
-- +goose StatementBegin
-- merchant_id (pos_merchant) and cashier_id (pos_merchant) live in another
-- logical database; both are validated by the cashier/merchant adapters over
-- gRPC instead of cross-database foreign keys.
CREATE TABLE "orders" (
    "order_id" SERIAL PRIMARY KEY,
    "merchant_id" INT NOT NULL,
    "cashier_id" INT NOT NULL,
    "total_price" BIGINT NOT NULL,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_orders_merchant_id ON orders (merchant_id);

CREATE INDEX idx_orders_cashier_id ON orders (cashier_id);

CREATE INDEX idx_orders_created_at ON orders (created_at);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_orders_merchant_id ON orders (merchant_id);

DROP INDEX IF EXISTS idx_orders_cashier_id ON orders (cashier_id);

DROP INDEX IF EXISTS idx_orders_created_at ON orders (created_at);

DROP TABLE IF EXISTS "orders";

-- +goose StatementEnd
