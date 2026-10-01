package po

import (
	"time"

	"gorm.io/gorm"
)

// SessionTokenUsage Token消耗明细
type SessionTokenUsage struct {
	ID               uint           // through Model
	CreatedAt        time.Time      // through Model
	UpdatedAt        time.Time      // through Model
	DeletedAt        gorm.DeletedAt // through Model
	SessionID        uint64         `gorm:"index;comment:会话ID"`
	MessageID        *uint64        `gorm:"index;comment:关联消息ID，可为NULL"`
	CallID           string         `gorm:"type:VARCHAR(128);comment:Eino callback调用ID"`
	ModelName        string         `gorm:"type:VARCHAR(128);comment:模型名称"`
	PromptTokens     int            `gorm:"default:0;comment:输入token"`
	CompletionTokens int            `gorm:"default:0;comment:输出token"`
	TotalTokens      int            `gorm:"default:0;comment:总token"`
}

func (SessionTokenUsage) TableName() string {
	return "session_token_usage"
}
