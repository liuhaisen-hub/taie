package po

import (
	"time"

	"gorm.io/gorm"
)

// AIModel maps to the ai_model table defined in migrations/create_ai_model.sql.
type AIModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除时间，不暴露给前端

	Name     string `gorm:"type:varchar(50);not null;default:''" json:"name"` // 模型名称
	Key      string `gorm:"type:varchar(500);not null" json:"key"`            // 模型 API Key
	BaseURL  string `gorm:"type:varchar(100);not null" json:"base_url"`       // 模型服务地址
	Type     int32  `gorm:"not null;index" json:"type"`                       // 模型类型
	Provider string `gorm:"type:varchar(100);not null" json:"provider"`       // 模型提供方
	Enable   int32  `gorm:"not null;index" json:"enable"`                     // 是否启用 0-禁用 1-启用
}

// TableName overrides GORM's default naming (ai_models) to match the migration.
func (AIModel) TableName() string {
	return "ai_model"
}
