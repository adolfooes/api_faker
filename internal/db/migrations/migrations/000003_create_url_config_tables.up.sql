-- Create ENUM for valid HTTP methods
CREATE TYPE http_method_enum AS ENUM ('GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD');

-- Create the url_config table to store paths and valid methods
CREATE TABLE url_config (
    id SERIAL PRIMARY KEY,                  -- Auto-incrementing primary key for each URL config
    path VARCHAR(255) NOT NULL,             -- The URL path (e.g., /api/resource)
    method http_method_enum NOT NULL,       -- HTTP method (GET, POST, PUT, DELETE, etc.)
    description TEXT,                       -- Optional description of the URL
    z_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Timestamp of URL creation
    z_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP -- Timestamp of last update
);

-- Create the url_http_status table to store valid HTTP statuses for each URL
CREATE TABLE url_http_status (
    id SERIAL PRIMARY KEY,                  -- Auto-incrementing primary key
    url_id INT NOT NULL,                    -- Foreign key to the url_config table
    http_status INT NOT NULL,               -- Valid HTTP status code (e.g., 200, 404, 500)
    percentage INT NOT NULL CHECK (percentage BETWEEN 0 AND 100), -- Percentage of responses for this status
    z_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Timestamp of association creation
    z_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, -- Timestamp of last update
    FOREIGN KEY (url_id) REFERENCES url_config(id) ON DELETE CASCADE  -- Foreign key referencing url_config
);

-- Create the response_model table to store response models for each status
CREATE TABLE response_model (
    id SERIAL PRIMARY KEY,                  -- Auto-incrementing primary key
    url_http_status_id INT NOT NULL,        -- Foreign key to the url_http_status table
    model JSONB NOT NULL,                   -- Response model stored as JSONB
    description TEXT,                       -- Optional description of the model
    z_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Timestamp of model creation
    z_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, -- Timestamp of last update
    FOREIGN KEY (url_http_status_id) REFERENCES url_http_status(id) ON DELETE CASCADE  -- Foreign key referencing url_http_status
);

-- Create index for faster queries on status models
CREATE INDEX idx_response_model_http_status ON response_model (url_http_status_id);
