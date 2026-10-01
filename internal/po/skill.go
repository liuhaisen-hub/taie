package po

import (
	"time"

	"gorm.io/gorm"
)

const (
	BuiltinSkill = "builtin"
	MCPToolSkill = "mcp_tool"
	HTTPAPISkill = "http_api"
)

// Skill maps to the skills table defined in migrations/create_skills.sql.
type Skill struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name         string `gorm:"type:text;not null;uniqueIndex:idx_skills_name" json:"name"` // 技能唯一名称，供 agent 调用
	Type         string `gorm:"type:text;not null" json:"type"`                             // builtin | mcp_tool | http_api
	Description  string `gorm:"type:text;not null" json:"description"`                      // 给大模型的工具描述
	InputSchema  string `gorm:"column:input_schema;type:text;not null" json:"inputSchema"`
	OutputSchema string `gorm:"column:output_schema;type:text" json:"outputSchema"`
	Config       string `gorm:"type:text" json:"config"` // JSON，技能配置，不同 type 结构不同
	TimeoutMs    int32  `gorm:"not null;default:30000" json:"timeoutMs"`
	Enabled      bool   `gorm:"not null;default:true;index" json:"enabled"`
	Tags         string `gorm:"type:text" json:"tags"` // JSON []string，标签
}

// TableName overrides GORM's default naming to match the migration.
func (Skill) TableName() string {
	return "skills"
}
