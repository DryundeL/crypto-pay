CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY,
    invoice_id UUID NOT NULL REFERENCES invoices (id),
    merchant_id UUID NOT NULL REFERENCES merchants (id),
    status TEXT NOT NULL,
    network TEXT NOT NULL,
    tx_hash TEXT NOT NULL,
    to_address TEXT NOT NULL,
    amount TEXT NOT NULL,
    currency TEXT NOT NULL,
    confirmations INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payments_status_chk CHECK (
        status IN ('detected', 'confirming', 'confirmed', 'failed')
    ),
    CONSTRAINT payments_network_tx_uq UNIQUE (network, tx_hash),
    CONSTRAINT payments_amount_len_chk CHECK (char_length(amount) BETWEEN 1 AND 64),
    CONSTRAINT payments_currency_len_chk CHECK (char_length(currency) BETWEEN 1 AND 32),
    CONSTRAINT payments_network_len_chk CHECK (char_length(network) BETWEEN 1 AND 64),
    CONSTRAINT payments_tx_hash_len_chk CHECK (char_length(tx_hash) BETWEEN 1 AND 128),
    CONSTRAINT payments_to_address_len_chk CHECK (char_length(to_address) BETWEEN 1 AND 256),
    CONSTRAINT payments_confirmations_chk CHECK (confirmations >= 0)
);

CREATE INDEX IF NOT EXISTS idx_payments_invoice_id
    ON payments (invoice_id);

CREATE INDEX IF NOT EXISTS idx_payments_merchant_id
    ON payments (merchant_id);
