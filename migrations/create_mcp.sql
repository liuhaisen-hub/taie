CREATE TABLE IF NOT EXISTS mcp_servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL, -- MCP服务唯一名称，用于agent引用
    description TEXT,
    transport TEXT NOT NULL CHECK(transport IN ('stdio','sse')),
    command TEXT, -- stdio: 启动命令，例如 npx
    args TEXT,    -- stdio: json数组字符串，命令参数
    env TEXT,     -- json对象，环境变量 map[string]string
    sse_url TEXT, -- sse transport: http地址
    timeout_ms INTEGER DEFAULT 30000, -- 请求超时毫秒
    enabled BOOLEAN NOT NULL DEFAULT true, -- 是否启用
    capabilities TEXT, -- json，缓存mcp服务返回的capabilities
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 唯一索引：服务名称不能重复
CREATE UNIQUE INDEX IF NOT EXISTS idx_mcp_servers_name ON mcp_servers(name) WHERE deleted_at IS NULL;
-- 普通索引：按启用状态查询
CREATE INDEX IF NOT EXISTS idx_mcp_servers_enabled ON mcp_servers(enabled, deleted_at);