ALTER TABLE merchants DROP CONSTRAINT IF EXISTS merchants_webhook_secret_len_chk;
ALTER TABLE merchants DROP COLUMN IF EXISTS webhook_secret;
