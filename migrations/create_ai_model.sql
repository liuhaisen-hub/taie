CREATE TABLE IF NOT EXISTS ai_model (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    name VARCHAR(50) NOT NULL DEFAULT '',
    key VARCHAR(500) NOT NULL,
    base_url VARCHAR(100) NOT NULL,
    type INTEGER NOT NULL,
    provider VARCHAR(100) NOT NULL,
    enable INT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_ai_model_type_enable_del ON ai_model(type, enable, deleted_at);
