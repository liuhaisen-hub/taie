package po

import (
	"time"

	"gorm.io/gorm"
)

type ChatSession struct {
	ID           uint           `json:"id"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-"` // 软删除时间，不暴露给前端
	Title        string         `json:"title" gorm:"type:VARCHAR(128);comment:会话标题"`
	SystemPrompt string         `json:"systemPrompt" gorm:"type:TEXT;comment:会话system prompt"`
	ModelName    string         `json:"modelName" gorm:"type:varchar(50);not null;default:''"` // 模型名称
	ApiKey       string         `json:"apiKey" gorm:"type:varchar(500);not null"`              // 模型 API Key
	BaseURL      string         `json:"baseUrl" gorm:"type:varchar(100);not null"`             // 模型服务地址
	Meta         []byte         `json:"meta" gorm:"type:BLOB;comment:扩展元信息json bytes"`
	EndedAt      *time.Time     `json:"endedAt" gorm:"comment:会话结束时间，可为null"`
}

func (ChatSession) TableName() string {
	return "chat_session"
}
