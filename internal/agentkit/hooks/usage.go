package hooks

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"taie/internal/agentkit/rctx"
	"taie/internal/po"
)

// UsageRecorder 只依赖落库所需的最小接口，AgentRepo 已隐式满足
// （接口定义在本包、由 data 层实现，避免反向依赖 agentkit 根包）。
type UsageRepo interface {
	AppendTokenUsage(ctx context.Context, u *po.SessionTokenUsage) error
}

// usageCallSeq 全局调用序号：per-turn 计数在多轮会话里会产生重复 CallID，
// 全局递增保证进程内唯一。进程重启归零无妨，CallID 非唯一键。
var usageCallSeq atomic.Uint64

// usageMiddleware 每次模型调用落一行 token 明细：WrapModel 挂在模型调用源头，
// 不依赖前端事件流的消费路径（异常分支、防御路径的调用也必经此处）。
// sessionID/modelName 不放实例字段，从 ctx 读取（rctx，由 Runner 注入），
// 计数器为全局 atomic——实例彻底无状态，deep agent 子代理共享并发安全。
type usageMiddleware struct {
	*adk.BaseChatModelAgentMiddleware
	repo UsageRepo
}

// WrapModel 每次模型调用前被 adk 回调，返回包了一层 usage 观测的模型。
// retry 包在本层之外，每次重试 attempt 都会经过这里；失败的 attempt
// 无 usage，record 的全 0 防御自然跳过。
func (m *usageMiddleware) WrapModel(_ context.Context, mdl model.BaseModel[*schema.Message],
	_ *adk.TypedModelContext[*schema.Message],
) (model.BaseModel[*schema.Message], error) {
	return &usageModel{mw: m, inner: mdl}, nil
}

// usageModel 透明代理：只观测，不改写输入输出。
type usageModel struct {
	mw    *usageMiddleware
	inner model.BaseModel[*schema.Message]
}

func (u *usageModel) Generate(ctx context.Context, input []*schema.Message,
	opts ...model.Option,
) (*schema.Message, error) {
	msg, err := u.inner.Generate(ctx, input, opts...)
	if err == nil && msg != nil && msg.ResponseMeta != nil {
		u.mw.record(ctx, msg.ResponseMeta.Usage)
	}
	return msg, err
}

func (u *usageModel) Stream(ctx context.Context, input []*schema.Message,
	opts ...model.Option,
) (*schema.StreamReader[*schema.Message], error) {
	sr, err := u.inner.Stream(ctx, input, opts...)
	if err != nil {
		return nil, err
	}
	var acc schema.TokenUsage
	seen := false
	// StreamReaderWithConvert：convert 在消费方 goroutine 上对每个 chunk 内联
	// 调用（无额外协程），透明透传 chunk 并抓取 usage；WithOnEOF 在 EOF 时内联
	// 回调一次，即流末尾 usage 帧之后，是同步落库点。
	return schema.StreamReaderWithConvert(sr, func(msg *schema.Message) (*schema.Message, error) {
		if msg != nil && msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
			mergeMaxUsage(&acc, msg.ResponseMeta.Usage)
			seen = true
		}
		return msg, nil
	}, schema.WithOnEOF(func() (any, error) {
		// 落库失败必须吞掉：WithOnEOF 返回非 EOF error 会先投递给消费方，
		// 会让 DB 抖动直接打断推理流。返回 io.EOF 让流正常结束。
		if seen {
			u.mw.record(ctx, &acc)
		}
		return nil, io.EOF
	})), nil
}

// record 落一行 token 明细。usage 是否返回由模型/网关决定，拿不到或全 0 时
// 跳过；落库失败静默，不影响对话主流程。
func (m *usageMiddleware) record(ctx context.Context, u *schema.TokenUsage) {
	if u == nil || (u.PromptTokens == 0 && u.CompletionTokens == 0 && u.TotalTokens == 0) {
		return
	}
	sid, ok := rctx.GetSessionID(ctx)
	if !ok || sid <= 0 {
		return
	}
	_ = m.repo.AppendTokenUsage(ctx, &po.SessionTokenUsage{
		SessionID:        uint64(sid),
		CallID:           fmt.Sprintf("chatmodel-%d-%d", sid, usageCallSeq.Add(1)),
		ModelName:        rctx.GetModelName(ctx),
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
	})
}

// mergeMaxUsage 按 max 合并 usage，与 schema.ConcatMessage 同规则，
// 防个别网关在多个 chunk 上重复携带 usage 导致累加虚高。
func mergeMaxUsage(acc *schema.TokenUsage, u *schema.TokenUsage) {
	if u.PromptTokens > acc.PromptTokens {
		acc.PromptTokens = u.PromptTokens
	}
	if u.CompletionTokens > acc.CompletionTokens {
		acc.CompletionTokens = u.CompletionTokens
	}
	if u.TotalTokens > acc.TotalTokens {
		acc.TotalTokens = u.TotalTokens
	}
}
