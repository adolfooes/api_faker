-- Drop triggers
DROP TRIGGER IF EXISTS trigger_project_updated_at ON project;
DROP TRIGGER IF EXISTS trigger_project_users_updated_at ON project_users;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_project_updated_at;
DROP FUNCTION IF EXISTS update_project_users_updated_at;

-- Drop the project_users and project tables
DROP TABLE IF EXISTS project_users;
DROP TABLE IF EXISTS project;

-- Drop the ENUM type for access levels
DROP TYPE IF EXISTS access_level_enum;

-- Drop the index
DROP INDEX IF EXISTS idx_project_users_account_project;
