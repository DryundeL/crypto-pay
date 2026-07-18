-- Platform transactional outbox (shared infrastructure, not a BC table).
CREATE TABLE IF NOT EXISTS outbox_messages (
    id UUID PRIMARY KEY,
    event_name TEXT NOT NULL,
    aggregate_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ NULL,
    CONSTRAINT outbox_messages_status_chk CHECK (status IN ('pending', 'sent', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_outbox_messages_pending
    ON outbox_messages (created_at ASC)
    WHERE status = 'pending';
