CREATE TABLE account (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    last_login TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_account_email_active ON account (email, is_active);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_updated_at
BEFORE UPDATE ON account
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


CREATE TYPE access_level_enum AS ENUM ('read', 'write', 'admin');

CREATE TABLE project (
    id SERIAL PRIMARY KEY,
    owner_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    removed_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES account(id) ON DELETE CASCADE
);

CREATE TABLE project_users (
    project_id INT NOT NULL,
    account_id INT NOT NULL,
    access_level access_level_enum DEFAULT 'read',
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    removed_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (project_id, account_id),
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_project_users_account_project ON project_users (account_id, project_id);

CREATE TYPE http_method_enum AS ENUM ('GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD');

CREATE TABLE url_config (
    id SERIAL PRIMARY KEY,
    path VARCHAR(255) NOT NULL,
    method http_method_enum NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE url_http_status (
    id SERIAL PRIMARY KEY,
    url_id INT NOT NULL,
    http_status INT NOT NULL,
    percentage INT NOT NULL CHECK (percentage BETWEEN 0 AND 100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (url_id) REFERENCES url_config(id) ON DELETE CASCADE
);

CREATE TABLE response_model (
    id SERIAL PRIMARY KEY,
    url_http_status_id INT NOT NULL,
    model JSONB NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (url_http_status_id) REFERENCES url_http_status(id) ON DELETE CASCADE
);

CREATE INDEX idx_response_model_http_status ON response_model (url_http_status_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_updated_at_url_config
BEFORE UPDATE ON url_config
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_updated_at_url_http_status
BEFORE UPDATE ON url_http_status
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_updated_at_response_model
BEFORE UPDATE ON response_model
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
