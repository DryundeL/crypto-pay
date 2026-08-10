CREATE TABLE IF NOT EXISTS blockchain_scan_cursors (
    network TEXT PRIMARY KEY,
    block_number BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT blockchain_scan_cursors_block_number_chk CHECK (block_number >= 0),
    CONSTRAINT blockchain_scan_cursors_network_len_chk CHECK (char_length(network) BETWEEN 1 AND 64)
);
