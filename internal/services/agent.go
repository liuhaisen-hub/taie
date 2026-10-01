package services

import (
	"context"
	"fmt"
	"taie/internal/agentkit"
	"taie/internal/agentkit/provider"
	"taie/internal/pkg/events"
	"taie/internal/pkg/utils"
	"taie/internal/po"
)

type SessionRepo interface {
	Create(ctx context.Context, ss *po.ChatSession) (*po.ChatSession, error)
	// Update 保存会话的所有业务字段
	// GetByID 按主键查询单条会话记录
	GetByID(ctx context.Context, id uint) (*po.ChatSession, error)
}
type AgentServices struct {
	model   ModelRepo
	session SessionRepo
	chagent *agentkit.ChatAgent
}
type ChatRequest struct {
	SessionID uint   `json:"session_id"` // 0 = 首条消息，自动创建会话
	UserInput string `json:"user_input"`
	// Extra 承载前端配置区（深度思考/模型选择等）的透传值，后端暂不消费。
	// 注意 JSON number 反序列化后是 float64，将来消费时需类型断言。
	Extra map[string]any `json:"extra,omitempty"`
}

func NewAgentServices(model ModelRepo, chatagent *agentkit.ChatAgent, session SessionRepo) *AgentServices {
	return &AgentServices{
		session: session,
		model:   model,
		chagent: chatagent,
	}
}

func (a *AgentServices) ChatWithAgent(ctx context.Context, req *ChatRequest, onEvent func(events.Event) error) {
	sess, err := a.ensureSession(ctx, req)
	if err != nil {
		_ = onEvent(events.Error(fmt.Sprintf("获取会话失败: %v", err)))
		return
	}
	// Runner 启动前先把会话 ID 回推给前端：新会话此时才落库，
	// 前端收到 begin 后才能在后续请求中带上 session_id。
	if err := onEvent(events.Begin(sess.ID)); err != nil {
		return
	}
	a.chagent.Runner(ctx, sess.ID, req.UserInput, &provider.BuildParams{
		ApiKey:    sess.ApiKey,
		BaseUrl:   sess.BaseURL,
		ModalName: sess.ModelName,
	}, func(e events.Event) error {
		return onEvent(e)
	})
}

func (a *AgentServices) ensureSession(ctx context.Context, req *ChatRequest) (*po.ChatSession, error) {
	if req.SessionID == 0 {
		// 新建session
		conf, err := a.getAgentConf(ctx)
		if err != nil {
			return nil, err
		}
		return a.session.Create(ctx, &po.ChatSession{
			Title:     utils.TruncateWithEllipsis(req.UserInput, 8),
			ApiKey:    conf.ApiKey,
			BaseURL:   conf.BaseURL,
			ModelName: conf.ModelName,
		})
	}
	return a.session.GetByID(ctx, req.SessionID)
}

func (a *AgentServices) getAgentConf(ctx context.Context) (*po.AIModel, error) {
	// 将来在这里实现模型并发配置
	list, err := a.model.GetByType(ctx, po.LanguageModel)
	if err != nil {
		return nil, err
	}
	if len(list) <= 0 {
		return nil, fmt.Errorf("当前未配置语言大模型")
	}
	// 当前只返回一个
	return list[0], err
}

// ApprovalRequest 工具审批裁决。interrupt_id 来自 approval 事件，必须原样回传。
type ApprovalRequest struct {
	SessionID   uint   `json:"session_id"`
	InterruptID string `json:"interrupt_id"`
	Approved    bool   `json:"approved"`
	Reason      string `json:"reason"` // 拒绝理由，approved=false 时生效
}

// ApprovalWithAgent 审批后续跑。前端复用与 agent/chat 相同的事件消费逻辑
// （delta/tool_call/tool_result/approval/done/error），approval 会话中可能再次出现
//（同批多工具先后审批）。
func (a *AgentServices) ApprovalWithAgent(ctx context.Context, req *ApprovalRequest, onEvent func(events.Event) error) {
	sess, err := a.session.GetByID(ctx, req.SessionID)
	if err != nil {
		_ = onEvent(events.Error(fmt.Sprintf("获取会话失败: %v", err)))
		return
	}
	a.chagent.Resume(ctx, sess.ID, &provider.BuildParams{
		ApiKey:    sess.ApiKey,
		BaseUrl:   sess.BaseURL,
		ModalName: sess.ModelName,
	}, req.InterruptID, req.Approved, req.Reason, func(e events.Event) error {
		return onEvent(e)
	})
}
