-- Create ENUM for access levels
CREATE TYPE access_level_enum AS ENUM ('read', 'write', 'admin');

-- Create the project table
CREATE TABLE project (
    id SERIAL PRIMARY KEY,                  -- Auto-incrementing primary key for each project
    owner_id INT NOT NULL,                  -- Foreign key to the account table (owner/creator)
    name VARCHAR(255) NOT NULL,             -- Name of the project
    description TEXT,                       -- Description of the project
    removed_at TIMESTAMP NULL,              -- Timestamp of project removal (soft delete)
    is_active BOOLEAN DEFAULT TRUE,         -- Indicates if the project is active
    FOREIGN KEY (owner_id) REFERENCES account(id) ON DELETE CASCADE -- Foreign key referencing account
    z_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- Timestamp of project creation
    z_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,  -- Timestamp of last update
);

-- Create project_users table to allow sharing project with other users
CREATE TABLE project_users (
    project_id INT NOT NULL,                -- Foreign key to project
    account_id INT NOT NULL,                -- Foreign key to account
    access_level access_level_enum DEFAULT 'read', -- Access level (ENUM type)
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Timestamp when the user was added
    removed_at TIMESTAMP NULL,              -- Timestamp when the user was removed (soft delete)
    is_active BOOLEAN DEFAULT TRUE,         -- Indicates if the access is still active
    PRIMARY KEY (project_id, account_id),   -- Composite primary key
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE, -- Foreign key referencing project
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE  -- Foreign key referencing account
);

-- Create index on account_id and project_id for better performance
CREATE INDEX idx_project_users_account_project ON project_users (account_id, project_id);
