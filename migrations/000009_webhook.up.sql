CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL REFERENCES merchants (id),
    event_name TEXT NOT NULL,
    source_id TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    url TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL,
    attempt_count INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 8,
    next_attempt_at TIMESTAMPTZ NOT NULL,
    last_error TEXT NULL,
    last_status_code INT NULL,
    delivered_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT webhook_deliveries_idempotency_uq UNIQUE (idempotency_key),
    CONSTRAINT webhook_deliveries_status_chk CHECK (
        status IN ('pending', 'sent', 'failed')
    ),
    CONSTRAINT webhook_deliveries_event_name_len_chk CHECK (char_length(event_name) BETWEEN 1 AND 128),
    CONSTRAINT webhook_deliveries_source_id_len_chk CHECK (char_length(source_id) BETWEEN 1 AND 128),
    CONSTRAINT webhook_deliveries_idempotency_len_chk CHECK (char_length(idempotency_key) BETWEEN 1 AND 256),
    CONSTRAINT webhook_deliveries_url_len_chk CHECK (char_length(url) BETWEEN 1 AND 2048),
    CONSTRAINT webhook_deliveries_attempts_chk CHECK (
        attempt_count >= 0 AND max_attempts >= 1 AND attempt_count <= max_attempts
    )
);

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_due
    ON webhook_deliveries (status, next_attempt_at)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_merchant_created
    ON webhook_deliveries (merchant_id, created_at DESC);
