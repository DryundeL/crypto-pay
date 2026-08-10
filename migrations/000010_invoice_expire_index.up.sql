-- expire job: status + expires_at lookup (replaces pending-only partial from 000004)
DROP INDEX IF EXISTS idx_invoices_expires_at;

CREATE INDEX IF NOT EXISTS idx_invoices_status_expires_at
    ON invoices (status, expires_at);
