CREATE TABLE IF NOT EXISTS ledger_accounts (
    id UUID PRIMARY KEY,
    owner_type TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    currency TEXT NOT NULL,
    balance TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ledger_accounts_owner_type_chk CHECK (owner_type IN ('system', 'merchant')),
    CONSTRAINT ledger_accounts_kind_chk CHECK (kind IN ('clearing', 'available')),
    CONSTRAINT ledger_accounts_owner_kind_chk CHECK (
        (owner_type = 'system' AND kind = 'clearing')
        OR (owner_type = 'merchant' AND kind = 'available')
    ),
    CONSTRAINT ledger_accounts_balance_len_chk CHECK (char_length(balance) BETWEEN 1 AND 64),
    CONSTRAINT ledger_accounts_currency_len_chk CHECK (char_length(currency) BETWEEN 1 AND 32),
    CONSTRAINT ledger_accounts_owner_id_len_chk CHECK (char_length(owner_id) BETWEEN 1 AND 64),
    CONSTRAINT ledger_accounts_key_uq UNIQUE (owner_type, owner_id, kind, currency)
);

CREATE TABLE IF NOT EXISTS ledger_journals (
    id UUID PRIMARY KEY,
    idempotency_key TEXT NOT NULL,
    merchant_id UUID NOT NULL REFERENCES merchants (id),
    reference_type TEXT NOT NULL,
    reference_id TEXT NOT NULL,
    amount TEXT NOT NULL,
    currency TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ledger_journals_idempotency_uq UNIQUE (idempotency_key),
    CONSTRAINT ledger_journals_idempotency_len_chk CHECK (char_length(idempotency_key) BETWEEN 1 AND 128),
    CONSTRAINT ledger_journals_reference_type_len_chk CHECK (char_length(reference_type) BETWEEN 1 AND 64),
    CONSTRAINT ledger_journals_reference_id_len_chk CHECK (char_length(reference_id) BETWEEN 1 AND 128),
    CONSTRAINT ledger_journals_amount_len_chk CHECK (char_length(amount) BETWEEN 1 AND 64),
    CONSTRAINT ledger_journals_currency_len_chk CHECK (char_length(currency) BETWEEN 1 AND 32)
);

CREATE INDEX IF NOT EXISTS idx_ledger_journals_merchant_id
    ON ledger_journals (merchant_id);

CREATE INDEX IF NOT EXISTS idx_ledger_journals_merchant_created
    ON ledger_journals (merchant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id UUID PRIMARY KEY,
    journal_id UUID NOT NULL REFERENCES ledger_journals (id),
    account_id UUID NOT NULL REFERENCES ledger_accounts (id),
    side TEXT NOT NULL,
    amount TEXT NOT NULL,
    currency TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ledger_entries_side_chk CHECK (side IN ('debit', 'credit')),
    CONSTRAINT ledger_entries_amount_len_chk CHECK (char_length(amount) BETWEEN 1 AND 64),
    CONSTRAINT ledger_entries_currency_len_chk CHECK (char_length(currency) BETWEEN 1 AND 32)
);

CREATE INDEX IF NOT EXISTS idx_ledger_entries_journal_id
    ON ledger_entries (journal_id);

CREATE INDEX IF NOT EXISTS idx_ledger_entries_account_id
    ON ledger_entries (account_id);
