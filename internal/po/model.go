package po

import (
	"time"

	"gorm.io/gorm"
)

const (
	MultimodalModel = 1
	LanguageModel   = 2
	VectorModel     = 3
	VisualModel     = 4
	VoiceModel      = 5
)

// AIModel maps to the ai_model table defined in migrations/create_ai_model.sql.
type AIModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除时间，不暴露给前端

	ModelName string `gorm:"type:varchar(50);not null;default:''" json:"modelName"` // 模型名称
	ApiKey    string `gorm:"type:varchar(500);not null" json:"apiKey"`              // 模型 API Key
	BaseURL   string `gorm:"type:varchar(100);not null" json:"baseUrl"`             // 模型服务地址
	Type      int32  `gorm:"not null;index" json:"type"`                            // 模型类型
	Provider  string `gorm:"type:varchar(100);not null" json:"provider"`            // 模型提供方
	Enable    int32  `gorm:"not null;index" json:"enable"`                          // 是否启用 0-禁用 1-启用
}

// TableName overrides GORM's default naming (ai_models) to match the migration.
func (AIModel) TableName() string {
	return "ai_model"
}
