CREATE TABLE IF NOT EXISTS blockchain_network_counters (
    network TEXT PRIMARY KEY,
    next_index BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT blockchain_network_counters_next_index_chk CHECK (next_index >= 0)
);

CREATE TABLE IF NOT EXISTS blockchain_addresses (
    id UUID PRIMARY KEY,
    network TEXT NOT NULL,
    address TEXT NOT NULL,
    derivation_path TEXT NOT NULL,
    invoice_id UUID NOT NULL,
    currency TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT blockchain_addresses_network_address_uq UNIQUE (network, address),
    CONSTRAINT blockchain_addresses_network_invoice_uq UNIQUE (network, invoice_id),
    CONSTRAINT blockchain_addresses_network_len_chk CHECK (char_length(network) BETWEEN 1 AND 64),
    CONSTRAINT blockchain_addresses_address_len_chk CHECK (char_length(address) BETWEEN 1 AND 256),
    CONSTRAINT blockchain_addresses_currency_len_chk CHECK (char_length(currency) BETWEEN 1 AND 32)
);

CREATE INDEX IF NOT EXISTS idx_blockchain_addresses_invoice_id
    ON blockchain_addresses (invoice_id);

CREATE TABLE IF NOT EXISTS blockchain_transactions (
    id UUID PRIMARY KEY,
    network TEXT NOT NULL,
    tx_hash TEXT NOT NULL,
    to_address TEXT NOT NULL,
    amount TEXT NOT NULL,
    currency TEXT NOT NULL,
    confirmations INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT blockchain_transactions_network_tx_uq UNIQUE (network, tx_hash),
    CONSTRAINT blockchain_transactions_status_chk CHECK (status IN ('observed', 'confirmed')),
    CONSTRAINT blockchain_transactions_confirmations_chk CHECK (confirmations >= 0),
    CONSTRAINT blockchain_transactions_network_len_chk CHECK (char_length(network) BETWEEN 1 AND 64),
    CONSTRAINT blockchain_transactions_tx_hash_len_chk CHECK (char_length(tx_hash) BETWEEN 1 AND 128),
    CONSTRAINT blockchain_transactions_to_address_len_chk CHECK (char_length(to_address) BETWEEN 1 AND 256)
);
