CREATE TABLE IF NOT EXISTS withdrawal_withdrawals (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL REFERENCES merchants (id),
    idempotency_key TEXT NOT NULL,
    amount TEXT NOT NULL,
    currency TEXT NOT NULL,
    network TEXT NOT NULL,
    to_address TEXT NOT NULL,
    status TEXT NOT NULL,
    tx_hash TEXT NULL,
    journal_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT withdrawal_withdrawals_idempotency_uq UNIQUE (idempotency_key),
    CONSTRAINT withdrawal_withdrawals_status_chk CHECK (
        status IN ('requested', 'completed', 'rejected')
    ),
    CONSTRAINT withdrawal_withdrawals_idempotency_len_chk CHECK (char_length(idempotency_key) BETWEEN 1 AND 128),
    CONSTRAINT withdrawal_withdrawals_amount_len_chk CHECK (char_length(amount) BETWEEN 1 AND 64),
    CONSTRAINT withdrawal_withdrawals_currency_len_chk CHECK (char_length(currency) BETWEEN 1 AND 32),
    CONSTRAINT withdrawal_withdrawals_network_len_chk CHECK (char_length(network) BETWEEN 1 AND 64),
    CONSTRAINT withdrawal_withdrawals_to_address_len_chk CHECK (char_length(to_address) BETWEEN 1 AND 256),
    CONSTRAINT withdrawal_withdrawals_tx_hash_len_chk CHECK (
        tx_hash IS NULL OR char_length(tx_hash) BETWEEN 1 AND 128
    )
);

CREATE INDEX IF NOT EXISTS idx_withdrawal_withdrawals_merchant_created
    ON withdrawal_withdrawals (merchant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_withdrawal_withdrawals_merchant_status
    ON withdrawal_withdrawals (merchant_id, status);
