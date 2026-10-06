-- +goose Up
-- +goose StatementBegin
-- outbox_events lives in the publisher's own logical database. This DDL is
-- repeated (idempotently) in the identity and sales contexts because auth,
-- order and transaction publish outbox events from their own databases; the
-- IF NOT EXISTS guards keep a single-database fallback environment working.
CREATE TABLE IF NOT EXISTS "outbox_events" (
    "outbox_id" BIGSERIAL PRIMARY KEY,
    "topic" VARCHAR(255) NOT NULL,
    "event_key" VARCHAR(255) NOT NULL,
    "payload" JSONB NOT NULL,
    "status" VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK ("status" IN ('pending', 'delivered', 'dead')),
    "attempts" INT NOT NULL DEFAULT 0,
    "next_attempt_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_outbox_events_status_next_attempt" ON "outbox_events"("status", "next_attempt_at");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "outbox_events";
-- +goose StatementEnd
