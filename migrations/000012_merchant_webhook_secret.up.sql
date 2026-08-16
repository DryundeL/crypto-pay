ALTER TABLE merchants ADD COLUMN IF NOT EXISTS webhook_secret TEXT;

-- Backfill existing rows with a random value they do not know; call rotate after deploy.
UPDATE merchants
SET webhook_secret = 'whsec_' || replace(gen_random_uuid()::text || gen_random_uuid()::text, '-', '')
WHERE webhook_secret IS NULL OR webhook_secret = '';

ALTER TABLE merchants ALTER COLUMN webhook_secret SET NOT NULL;

ALTER TABLE merchants DROP CONSTRAINT IF EXISTS merchants_webhook_secret_len_chk;
ALTER TABLE merchants ADD CONSTRAINT merchants_webhook_secret_len_chk
    CHECK (char_length(webhook_secret) >= 32);
