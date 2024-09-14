-- Create account table
CREATE TABLE account (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    last_login TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create an index on email and is_active
CREATE INDEX idx_account_email_active ON account (email, is_active);

-- Create trigger function to update 'updated_at' on row update
CREATE OR REPLACE FUNCTION update_account_updated_at()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger on the account table for automatic 'updated_at' updates
CREATE TRIGGER trigger_account_updated_at
BEFORE UPDATE ON account
FOR EACH ROW
EXECUTE FUNCTION update_account_updated_at();
