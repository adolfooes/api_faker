-- Drop the trigger for 'updated_at'
DROP TRIGGER IF EXISTS trigger_account_updated_at ON account;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_account_updated_at();

-- Drop the account table
DROP TABLE IF EXISTS account;

-- Drop the index
DROP INDEX IF EXISTS idx_account_email_active;
