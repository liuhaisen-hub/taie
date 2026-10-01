-- chat_session 会话表
CREATE TABLE IF NOT EXISTS chat_session (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NULL,
    system_prompt TEXT NULL,
    model_name TEXT NOT NULL DEFAULT '',
    api_key TEXT NOT NULL,
    base_url TEXT NOT NULL,
    meta BLOB NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL
);
CREATE INDEX IF NOT EXISTS idx_chat_session_created_at ON chat_session(created_at);
