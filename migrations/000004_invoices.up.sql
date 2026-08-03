CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL REFERENCES merchants (id),
    status TEXT NOT NULL,
    amount TEXT NOT NULL,
    currency TEXT NOT NULL,
    network TEXT NOT NULL,
    address TEXT NOT NULL,
    tx_hash TEXT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT invoices_status_chk CHECK (
        status IN ('pending', 'confirming', 'paid', 'expired', 'cancelled')
    ),
    CONSTRAINT invoices_amount_len_chk CHECK (char_length(amount) BETWEEN 1 AND 64),
    CONSTRAINT invoices_currency_len_chk CHECK (char_length(currency) BETWEEN 1 AND 32),
    CONSTRAINT invoices_network_len_chk CHECK (char_length(network) BETWEEN 1 AND 64),
    CONSTRAINT invoices_address_len_chk CHECK (char_length(address) BETWEEN 1 AND 256)
);

CREATE INDEX IF NOT EXISTS idx_invoices_merchant_id
    ON invoices (merchant_id);

CREATE INDEX IF NOT EXISTS idx_invoices_merchant_status
    ON invoices (merchant_id, status);

-- scanner: FindPendingByAddress(network, address)
CREATE INDEX IF NOT EXISTS idx_invoices_network_address
    ON invoices (network, address);

-- worker: expire pending invoices by TTL
CREATE INDEX IF NOT EXISTS idx_invoices_expires_at
    ON invoices (expires_at)
    WHERE status = 'pending';
