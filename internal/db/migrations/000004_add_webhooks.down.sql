DROP INDEX IF EXISTS idx_webhook_dispatch_dispatched_at;
DROP INDEX IF EXISTS idx_webhook_dispatch_project;
DROP INDEX IF EXISTS idx_webhook_dispatch_owner;
DROP TABLE IF EXISTS webhook_dispatch;

ALTER TABLE url_config DROP COLUMN IF EXISTS webhooks;
