package po

import (
	"time"

	"gorm.io/gorm"
)

const (
	StdioTransport = "stdio"
	SSETransport   = "sse"
)

// MCPServer maps to the mcp_servers table defined in migrations/create_mcp.sql.
type MCPServer struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name         string `gorm:"type:text;not null;uniqueIndex:idx_mcp_servers_name" json:"name"` // MCP 服务唯一名称，供 agent 引用
	Description  string `gorm:"type:text" json:"description"`
	Transport    string `gorm:"type:text;not null" json:"transport"` // stdio | sse
	Command      string `gorm:"type:text" json:"command"`            // stdio: 启动命令，如 npx
	Args         string `gorm:"type:text" json:"args"`               // stdio: JSON 数组字符串，命令参数
	Env          string `gorm:"type:text" json:"env"`                // JSON 对象 map[string]string
	SSEURL       string `gorm:"column:sse_url;type:text" json:"sseUrl"`
	TimeoutMs    int32  `gorm:"not null;default:30000" json:"timeoutMs"`
	Enabled      bool   `gorm:"not null;default:true;index" json:"enabled"`
	Capabilities string `gorm:"type:text" json:"capabilities"` // 缓存 mcp 服务返回的 capabilities
}

// TableName overrides GORM's default naming (mcp_servers already matches, kept explicit for clarity).
func (MCPServer) TableName() string {
	return "mcp_servers"
}
