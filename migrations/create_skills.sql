CREATE TABLE IF NOT EXISTS skills (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL, -- 技能唯一名称，供agent调用
    type TEXT NOT NULL CHECK(type IN ('builtin','mcp_tool','http_api')),
    description TEXT NOT NULL, -- 给大模型的工具描述
    input_schema TEXT NOT NULL, -- json schema，入参
    output_schema TEXT, -- json schema，返回结构
    config TEXT, -- json，技能配置，不同type结构不一样
    timeout_ms INTEGER DEFAULT 30000,
    enabled BOOLEAN NOT NULL DEFAULT true,
    tags TEXT, -- json []string，标签
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 唯一索引：技能名称不能重复（未删除）
CREATE UNIQUE INDEX IF NOT EXISTS idx_skills_name ON skills(name) WHERE deleted_at IS NULL;
-- 索引：按启用状态查询
CREATE INDEX IF NOT EXISTS idx_skills_enabled ON skills(enabled, deleted_at);
