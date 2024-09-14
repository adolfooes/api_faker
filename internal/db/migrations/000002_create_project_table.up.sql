-- Define the ENUM type for access levels
CREATE TYPE access_level_enum AS ENUM ('read', 'write', 'admin');

-- Create the project table
CREATE TABLE project (
    id SERIAL PRIMARY KEY,
    owner_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    removed_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES account(id) ON DELETE CASCADE
);

-- Create the project_users table
CREATE TABLE project_users (
    project_id INT NOT NULL,
    account_id INT NOT NULL,
    access_level access_level_enum DEFAULT 'read',
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    removed_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (project_id, account_id),
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE
);

-- Create an index on project_users (account_id, project_id)
CREATE INDEX idx_project_users_account_project ON project_users (account_id, project_id);

-- Create trigger function to update 'updated_at' on row update for project
CREATE OR REPLACE FUNCTION update_project_updated_at()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger function to update 'updated_at' on row update for project_users
CREATE OR REPLACE FUNCTION update_project_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for the project table
CREATE TRIGGER trigger_project_updated_at
BEFORE UPDATE ON project
FOR EACH ROW
EXECUTE FUNCTION update_project_updated_at();

-- Create triggers for the project_users table
CREATE TRIGGER trigger_project_users_updated_at
BEFORE UPDATE ON project_users
FOR EACH ROW
EXECUTE FUNCTION update_project_users_updated_at();
