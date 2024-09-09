-- Create account table
CREATE TABLE
    account (
        id SERIAL PRIMARY KEY, -- Auto-incrementing primary key for each user
        email VARCHAR(255) NOT NULL UNIQUE, -- Unique email for each user
        password VARCHAR(255) NOT NULL, -- Hashed and salted password
        last_login TIMESTAMP NULL, -- Timestamp of the user's last login
        is_active BOOLEAN DEFAULT TRUE -- Indicates if the account is active or not
        z_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Timestamp of account creation
        z_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, -- Timestamp of last update
    );

-- Create an index on email and is_active
CREATE INDEX idx_account_email_active ON account (email, is_active);