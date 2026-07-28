CREATE TABLE IF NOT EXISTS invoice_invoices (
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
    CONSTRAINT invoice_invoices_status_chk CHECK (
        status IN ('pending', 'confirming', 'paid', 'expired', 'cancelled')
    ),
    CONSTRAINT invoice_invoices_amount_len_chk CHECK (char_length(amount) BETWEEN 1 AND 64),
    CONSTRAINT invoice_invoices_currency_len_chk CHECK (char_length(currency) BETWEEN 1 AND 32),
    CONSTRAINT invoice_invoices_network_len_chk CHECK (char_length(network) BETWEEN 1 AND 64),
    CONSTRAINT invoice_invoices_address_len_chk CHECK (char_length(address) BETWEEN 1 AND 256)
);

CREATE INDEX IF NOT EXISTS idx_invoice_invoices_merchant_id
    ON invoice_invoices (merchant_id);

CREATE INDEX IF NOT EXISTS idx_invoice_invoices_merchant_status
    ON invoice_invoices (merchant_id, status);

CREATE INDEX IF NOT EXISTS idx_invoice_invoices_address
    ON invoice_invoices (address);
