package data

import (
	"context"
	"encoding/json"
	"slices"
	"taie/internal/agentkit"
	"taie/internal/po"
)

type agentRepo struct {
	data *Data
}

func NewAgentRepo(data *Data) agentkit.AgentRepo {
	return &agentRepo{
		data: data,
	}
}

// GetHistory 返回某会话最近 limit 条消息，并按时间正序排列供回放；limit <= 0 时返回全部
func (a *agentRepo) GetHistory(ctx context.Context, sessionID uint, limit int) ([]*po.ChatMessage, error) {
	query := a.data.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	var msgs []*po.ChatMessage
	if err := query.Find(&msgs).Error; err != nil {
		return nil, err
	}
	// DESC 取出的是最新 limit 条，反转回时间正序
	slices.Reverse(msgs)
	return msgs, nil
}

// AppendHistory 原样落库一条消息
func (a *agentRepo) AppendHistory(ctx context.Context, m *po.ChatMessage) error {
	return a.data.db.WithContext(ctx).Create(m).Error
}

// AppendUserMessage 落库一条用户消息
func (a *agentRepo) AppendUserMessage(ctx context.Context, sessionID uint, content string) error {
	return a.data.db.WithContext(ctx).Create(&po.ChatMessage{
		SessionID: uint64(sessionID),
		Role:      po.UserRole,
		Content:   content,
	}).Error
}

// AppendAssistantMessage 落库一条 assistant 消息，status 存在 Extra 扩展字段中
func (a *agentRepo) AppendAssistantMessage(ctx context.Context, sessionID uint, content string, status int) error {
	extra, err := json.Marshal(map[string]int{"status": status})
	if err != nil {
		return err
	}
	return a.data.db.WithContext(ctx).Create(&po.ChatMessage{
		SessionID: uint64(sessionID),
		Role:      po.AssistantRole,
		Content:   content,
		Extra:     extra,
	}).Error
}

// AppendTokenUsage 落库一行 token 消耗明细（每次模型调用一条）
func (a *agentRepo) AppendTokenUsage(ctx context.Context, u *po.SessionTokenUsage) error {
	return a.data.db.WithContext(ctx).Create(u).Error
}

// ListEnable 返回所有已启用的工具，读取 .conf/tools.json（与工具设置页同一份配置）
func (a *agentRepo) ListEnableTools() ([]*po.ToolConfig, error) {
	cfgs, err := loadToolConfs()
	if err != nil {
		return nil, err
	}
	tools := make([]*po.ToolConfig, 0, len(cfgs))
	for i := range cfgs {
		if cfgs[i].Enable {
			tools = append(tools, &cfgs[i])
		}
	}
	return tools, nil
}
