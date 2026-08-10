DROP INDEX IF EXISTS idx_invoices_status_expires_at;

CREATE INDEX IF NOT EXISTS idx_invoices_expires_at
    ON invoices (expires_at)
    WHERE status = 'pending';
