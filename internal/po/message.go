package po

import (
	"time"

	"gorm.io/gorm"
)

type StatusType int

const (
	MessageFinishStatus = 0
	MessageExitStatus   = 1
	MessageErrorStatus  = 3
	UserRole            = "user"
	AssistantRole       = "assistant"
)

type ChatMessage struct {
	ID           uint           // through Model
	CreatedAt    time.Time      // through Model
	UpdatedAt    time.Time      // through Model
	DeletedAt    gorm.DeletedAt // through Model
	SessionID    uint64         `gorm:"index;comment:归属会话ID"`
	Role         string         `gorm:"type:VARCHAR(32);comment:消息角色"`
	Content      string         `gorm:"type:TEXT;comment:消息正文"`
	MultiContent []byte         `gorm:"type:BLOB;comment:多模态内容json bytes"`
	ToolCalls    []byte         `gorm:"type:BLOB;comment:[]schema.ToolCall json bytes"`
	ToolCallID   string         `gorm:"type:VARCHAR(128);comment:工具调用ID"`
	Name         string         `gorm:"type:VARCHAR(128);comment:function名称"`
	Extra        []byte         `gorm:"type:BLOB;comment:扩展信息json bytes"`
}

func (ChatMessage) TableName() string {
	return "chat_message"
}
