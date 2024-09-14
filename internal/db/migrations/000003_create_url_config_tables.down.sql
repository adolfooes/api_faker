-- Drop triggers
DROP TRIGGER IF EXISTS trigger_url_config_updated_at ON url_config;
DROP TRIGGER IF EXISTS trigger_url_http_status_updated_at ON url_http_status;
DROP TRIGGER IF EXISTS trigger_response_model_updated_at ON response_model;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_url_config_updated_at;
DROP FUNCTION IF EXISTS update_url_http_status_updated_at;
DROP FUNCTION IF EXISTS update_response_model_updated_at;

-- Drop the response_model, url_http_status, and url_config tables
DROP TABLE IF EXISTS response_model;
DROP TABLE IF EXISTS url_http_status;
DROP TABLE IF EXISTS url_config;

-- Drop the ENUM type for HTTP methods
DROP TYPE IF EXISTS http_method_enum;

-- Drop the index
DROP INDEX IF EXISTS idx_response_model_http_status;
