-- Define the ENUM type for HTTP methods
CREATE TYPE http_method_enum AS ENUM ('GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD');

-- Create the url_config table with a reference to project (project_id)
CREATE TABLE url_config (
    id SERIAL PRIMARY KEY,
    path VARCHAR(255) NOT NULL,
    method http_method_enum NOT NULL,
    description TEXT,
    project_id INT NOT NULL, -- Add project_id column
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE -- Foreign key to project table
);

-- Create the url_http_status table
CREATE TABLE url_http_status (
    id SERIAL PRIMARY KEY,
    url_id INT NOT NULL,
    http_status INT NOT NULL,
    percentage INT NOT NULL CHECK (percentage BETWEEN 0 AND 100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (url_id) REFERENCES url_config(id) ON DELETE CASCADE
);

-- Create the response_model table
CREATE TABLE response_model (
    id SERIAL PRIMARY KEY,
    url_http_status_id INT NOT NULL,
    model JSONB NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (url_http_status_id) REFERENCES url_http_status(id) ON DELETE CASCADE
);

-- Create an index on response_model (url_http_status_id)
CREATE INDEX idx_response_model_http_status ON response_model (url_http_status_id);

-- Create trigger function to update 'updated_at' on row update for url_config
CREATE OR REPLACE FUNCTION update_url_config_updated_at()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger function to update 'updated_at' on row update for url_http_status
CREATE OR REPLACE FUNCTION update_url_http_status_updated_at()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger function to update 'updated_at' on row update for response_model
CREATE OR REPLACE FUNCTION update_response_model_updated_at()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for the url_config table
CREATE TRIGGER trigger_url_config_updated_at
BEFORE UPDATE ON url_config
FOR EACH ROW
EXECUTE FUNCTION update_url_config_updated_at();

-- Create triggers for the url_http_status table
CREATE TRIGGER trigger_url_http_status_updated_at
BEFORE UPDATE ON url_http_status
FOR EACH ROW
EXECUTE FUNCTION update_url_http_status_updated_at();

-- Create triggers for the response_model table
CREATE TRIGGER trigger_response_model_updated_at
BEFORE UPDATE ON response_model
FOR EACH ROW
EXECUTE FUNCTION update_response_model_updated_at();
