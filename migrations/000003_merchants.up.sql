CREATE TABLE IF NOT EXISTS merchants (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    webhook_url TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT merchants_status_chk CHECK (status IN ('active', 'suspended')),
    CONSTRAINT merchants_name_len_chk CHECK (char_length(name) BETWEEN 1 AND 200)
);

CREATE INDEX IF NOT EXISTS idx_merchants_status ON merchants (status);

CREATE TABLE IF NOT EXISTS merchant_api_keys (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL REFERENCES merchants (id),
    name TEXT NOT NULL,
    key_prefix TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ NULL,
    CONSTRAINT merchant_api_keys_status_chk CHECK (status IN ('active', 'revoked')),
    CONSTRAINT merchant_api_keys_name_len_chk CHECK (char_length(name) BETWEEN 1 AND 100),
    CONSTRAINT merchant_api_keys_key_hash_uq UNIQUE (key_hash)
);

CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_merchant_id
    ON merchant_api_keys (merchant_id);

CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_merchant_status
    ON merchant_api_keys (merchant_id, status);
