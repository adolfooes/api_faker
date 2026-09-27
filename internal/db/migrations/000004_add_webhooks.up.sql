-- Add scenario-declared chained webhooks to url_config
ALTER TABLE url_config ADD COLUMN webhooks JSONB;

-- Log of every outbound webhook dispatch (admin-triggered or scenario-chained)
CREATE TABLE webhook_dispatch (
    id SERIAL PRIMARY KEY,
    owner_id INT NOT NULL,
    project_id INT NULL,
    target_url TEXT NOT NULL,
    method VARCHAR(10) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    request_headers JSONB,
    request_body TEXT,
    response_status INT,
    response_headers JSONB,
    response_body TEXT,
    error TEXT,
    duration_ms INT,
    dispatched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    FOREIGN KEY (owner_id) REFERENCES account(id) ON DELETE CASCADE
);

CREATE INDEX idx_webhook_dispatch_owner ON webhook_dispatch (owner_id);
CREATE INDEX idx_webhook_dispatch_project ON webhook_dispatch (project_id);
CREATE INDEX idx_webhook_dispatch_dispatched_at ON webhook_dispatch (dispatched_at DESC);
