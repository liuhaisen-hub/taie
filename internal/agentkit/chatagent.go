package agentkit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"taie/internal/agentkit/hooks"
	"taie/internal/agentkit/provider"
	"taie/internal/agentkit/rctx"
	"taie/internal/agentkit/toolkit"
	"taie/internal/agentkit/trace"

	"taie/internal/pkg/events"
	"taie/internal/po"
	"time"

	"github.com/cloudwego/eino-ext/adk/backend/local"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type AgentRepo interface {
	GetHistory(ctx context.Context, sessionD uint, limit int) ([]*po.ChatMessage, error)
	AppendHistory(ctx context.Context, m *po.ChatMessage) error
	AppendUserMessage(ctx context.Context, sessionID uint, content string) error
	AppendAssistantMessage(ctx context.Context, sessionID uint, content string, status int) error
	AppendTokenUsage(ctx context.Context, u *po.SessionTokenUsage) error
	ListEnableTools() ([]*po.ToolConfig, error)
}
type entry struct {
	build *provider.BuildParams
	fp    string // 构建参数指纹
	agent adk.ResumableAgent
	// runMu 串行化同一会话的运行：entry 里的 agent 与 middleware 实例被同会话
	// 所有请求共享，并发 Run 会踩彼此的状态，也和历史消息/checkpoint 交错
	runMu sync.Mutex
	// lastAccess 供空闲淘汰判断，getEntry 命中或新建时刷新
	lastAccess time.Time
}

// 聊天模型
type ChatAgent struct {
	repo AgentRepo
	// mu 保护 entries：ChatAgent 是单例，多个请求 goroutine 同时读写 map
	// 不加锁会触发 concurrent map read/write 的 fatal panic
	mu        sync.Mutex
	entries   map[uint]*entry
	cpStore   adk.CheckPointStore
	tracer    *trace.Tracer
	hooks     []adk.TypedChatModelAgentMiddleware[*schema.Message]
	toolsInfo map[string]string // 工具集缓存
	logger    *slog.Logger
}

func NewChatAgent(repo AgentRepo,
	cpStore adk.CheckPointStore,
	tracer *trace.Tracer,
	handler []adk.TypedChatModelAgentMiddleware[*schema.Message],
	logger *slog.Logger,
) *ChatAgent {
	return &ChatAgent{
		repo:    repo,
		cpStore: cpStore,
		tracer:  tracer,
		hooks:   handler,
		entries: make(map[uint]*entry),
		logger:  logger,
	}
}

func (c *ChatAgent) Runner(ctx context.Context, sessionID uint, query string, b *provider.BuildParams, onEvent func(events.Event) error) error {
	e, err := c.getEntry(ctx, sessionID, b)
	if err != nil {
		return onEvent(events.Error(fmt.Sprintf("初始化失败: %v", err)))
	}
	// 同一会话的请求在此排队，杜绝共享 agent 状态与历史消息交错
	e.runMu.Lock()
	defer e.runMu.Unlock()
	// 用户消息先落库，防止其他问题中断
	if err := c.repo.AppendUserMessage(ctx, sessionID, query); err != nil {
		return onEvent(events.Error(fmt.Sprintf("记录用户消息失败: %v", err)))
	}
	// 获取历史记录， 已经包含最新的消息
	message, err := c.buildMessage(ctx, sessionID)
	if err != nil {
		return onEvent(events.Error(fmt.Sprintf("读取历史消息失败: %v", err)))
	}
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           e.agent,
		EnableStreaming: true,
		CheckPointStore: c.cpStore,
	})
	// 注入sessionId与模型名，usage 中间件落库时从 ctx 读取归属信息
	ctx = rctx.WithSessionId(ctx, int64(sessionID))
	ctx = rctx.WithModelName(ctx, e.build.ModalName)
	agevs := runner.Run(ctx, message,
		adk.WithSessionValues(map[string]any{
			"Now": time.Now().Format("2006-01-02 15:04:05"),
		}),
		// 接入日志
		adk.WithCallbacks(c.tracer),
		// 中断/恢复的定位键：不带走就没有 checkpoint，审批无从谈起
		adk.WithCheckPointID(checkpointID(sessionID)),
	)
	em := events.NewEmitter(sessionID, onEvent)
	return c.pump(ctx, sessionID, agevs, em, onEvent)
}

// checkpointID 用会话 ID 作 checkpoint 键：一次会话同时只有一个挂起中断，
// 同会话下一轮 Run 的 Set（upsert）会自然覆盖。
func checkpointID(sessionID uint) string {
	return strconv.FormatUint(uint64(sessionID), 10)
}

// pump 消费 Runner 事件流，首轮 Run 与审批续跑 Resume 共用。
// 三类事件：Err / Action(审批中断) / Output(消息与流)。
func (c *ChatAgent) pump(ctx context.Context, sessionID uint,
	agevs *adk.AsyncIterator[*adk.AgentEvent], em *events.Emitter,
	onEvent func(events.Event) error,
) error {
	for {
		agev, ok := agevs.Next()
		if !ok {
			break
		}
		if agev.Err != nil {
			if ctx.Err() != nil {
				// 中断：已生成的部分按退出状态落库，并尽力回推 incomplete 收口
				_ = c.repo.AppendAssistantMessage(ctx, sessionID, em.Text(), po.MessageExitStatus)
				_ = em.Interrupted()
				return ctx.Err()
			}
			return onEvent(events.Error(agev.Err.Error()))
		}
		// 审批中断：工具执行前停下，发 approval 事件后本轮结束
		//（中断事件是该次迭代的最后一个事件，之后 Next() 即结束）。
		// 判断只看 Action，不要依赖该事件上的 Output（可能为 nil）。
		if agev.Action != nil && agev.Action.Interrupted != nil {
			name, args, icID, ok := approvalInterrupt(agev.Action.Interrupted)
			if !ok {
				// 非审批类中断，v1 统一按错误处理
				return onEvent(events.Error("执行被中断"))
			}
			// 挂起轮不落库、不发 done：回合等审批续跑后继续
			return em.Approval(name, args, icID)
		}
		if agev.Output == nil || agev.Output.MessageOutput == nil {
			continue
		}
		out := agev.Output.MessageOutput
		// Role 挂在 MessageVariant 上，流式模式下不消费流就能判断
		if out.Role == schema.Tool {
			// 工具回执：call_id 在 tool 消息上，前端按 call_id 把 function_call 置 completed
			content, callID := "", ""
			if out.IsStreaming && out.MessageStream != nil {
				content, callID = drainToolStream(out.MessageStream)
			} else if out.Message != nil {
				content, callID = out.Message.Content, out.Message.ToolCallID
			}
			if err := em.ToolResult(callID, content); err != nil {
				return err
			}
			continue
		}
		// assistant 消息：流式增量推给前端，同时累积全文与内容序列供 done 重组
		if out.IsStreaming {
			if err := em.ConsumeStream(out.MessageStream); err != nil {
				if ctx.Err() != nil {
					_ = c.repo.AppendAssistantMessage(ctx, sessionID, em.Text(), po.MessageExitStatus)
					return ctx.Err()
				}
				return onEvent(events.Error(fmt.Sprintf("执行错误: %v", err)))
			}
		} else if out.Message != nil {
			// 非流式兜底：必须与流式分支同级，嵌进流式分支会吞掉非流式模型的输出
			if err := em.AssistantMessage(out.Message); err != nil {
				return err
			}
		}
	}
	// 完成落库，Done 事件与落库内容同源
	_ = c.repo.AppendAssistantMessage(ctx, sessionID, em.Text(), po.MessageFinishStatus)
	return em.Done()
}

// approvalInterrupt 从中断链里挑出审批中断点（Info 为 *hooks.ApprovalInfo 的那个，
// 多个命中时优先 IsRootCause），其 ID 是 ResumeWithParams 的定位键。
func approvalInterrupt(info *adk.InterruptInfo) (toolName, argsJSON, interruptID string, ok bool) {
	for _, ic := range info.InterruptContexts {
		ai, is := ic.Info.(*hooks.ApprovalInfo)
		if !is {
			continue
		}
		if !ok || ic.IsRootCause {
			toolName, argsJSON, interruptID, ok = ai.ToolName, ai.ArgumentsInJSON, ic.ID, true
			if ic.IsRootCause {
				break
			}
		}
	}
	return
}

// Resume 审批裁决后续跑：从 checkpoint 恢复，把裁决精准投递给中断的那个工具
//（Targets 的 key 就是 approval 事件里原样回传的 interrupt_id），
// 之后的事件流（delta/tool_result/done，乃至下一个 approval）与首轮完全一致，
// 因此前端可复用同一套事件消费逻辑。
// Run 与 Resume 可跨进程/重启，前提是 CheckPointStore 持久化且 CheckPointID 一致
//——桌面应用重启后仍能审批，架构成立。
func (c *ChatAgent) Resume(ctx context.Context, sessionID uint, b *provider.BuildParams,
	interruptID string, approved bool, reason string, onEvent func(events.Event) error) error {

	e, err := c.getEntry(ctx, sessionID, b)
	if err != nil {
		return onEvent(events.Error(fmt.Sprintf("初始化失败: %v", err)))
	}
	// 与 Runner 同款串行化：entry 里的 agent 与 middleware 实例被同会话共享
	e.runMu.Lock()
	defer e.runMu.Unlock()
	res := &hooks.ApprovalResult{Approved: approved}
	if !approved && reason != "" {
		res.DisapproveReason = &reason
	}
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           e.agent,
		EnableStreaming: true,
		CheckPointStore: c.cpStore,
	})
	ctx = rctx.WithSessionId(ctx, int64(sessionID))
	ctx = rctx.WithModelName(ctx, e.build.ModalName)
	agevs, err := runner.ResumeWithParams(ctx, checkpointID(sessionID), &adk.ResumeParams{
		// key 必须是 approval 事件里原样回传的 interrupt_id
		Targets: map[string]any{interruptID: res},
	}, adk.WithCallbacks(c.tracer))
	if err != nil {
		return onEvent(events.Error(fmt.Sprintf("恢复执行失败: %v", err)))
	}
	em := events.NewEmitter(sessionID, onEvent)
	return c.pump(ctx, sessionID, agevs, em, onEvent)
}

// entries 的有界化：条目数超过 maxEntries 时，在 getEntry 插入新条目前顺带
// 清理空闲超过 idleTTL 的 entry。无后台 goroutine/定时器，清理成本摊在低频
// 的"新建 entry"路径上。用户极少删除会话，靠删除入口回收不可靠，缓存必须自清理。
const (
	maxEntries = 32
	idleTTL    = 30 * time.Minute
)

// evictIdleLocked 清理长期不用的 entry，调用方需持有 c.mu。
// runMu 占用中的 entry 正在运行，跳过：删 map 不影响进行中的 run，但会让
// 同会话下一个请求重建 entry、与旧 run 失去串行化，不如不删。
func (c *ChatAgent) evictIdleLocked(now time.Time) {
	if len(c.entries) <= maxEntries {
		return
	}
	for id, e := range c.entries {
		if now.Sub(e.lastAccess) < idleTTL {
			continue
		}
		if !e.runMu.TryLock() {
			continue
		}
		e.runMu.Unlock()
		delete(c.entries, id)
	}
}

func (c *ChatAgent) getEntry(ctx context.Context, sessionID uint, b *provider.BuildParams) (*entry, error) {
	now := time.Now()
	fpStr, err := json.Marshal(&b)
	if err != nil {
		return nil, fmt.Errorf("构建指纹失败")
	}
	fp, err := buildFingerprint(fpStr)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	// 顺带清理一轮空闲 entry，entries 上限由此维持
	//（必须持 c.mu 调用：遍历/删除 entries 与其他请求的读写并发，锁外是数据竞争）
	c.evictIdleLocked(now)
	// 指纹一致直接复用缓存
	if e, ok := c.entries[sessionID]; ok && e.fp == fp {
		e.lastAccess = now
		return e, nil
	}
	// 不存在或指纹变化，重建
	chat, err := provider.NsewOpenAIChatModel(ctx, b)
	if err != nil {
		return nil, err
	}
	backend, err := local.NewBackend(ctx, &local.Config{})
	if err != nil {
		return nil, err
	}
	// 工具集
	toolInfo, err := c.repo.ListEnableTools()
	if err != nil {
		c.logger.Error("工具注入失败", "error", "sessionId", err, sessionID)
	}
	if len(toolInfo) <= 0 {
		c.logger.Info("暂无注入工具", "sessionID", sessionID)
	}
	info := make(map[string][]po.CommonJson, len(toolInfo))
	if len(toolInfo) > 0 {

		for _, item := range toolInfo {
			info[item.ToolName] = item.Args
		}
	}
	tools := toolkit.BuildTools(info)
	agent, err := deep.New(ctx, &deep.Config{
		Name:           "智能助手",
		Instruction:    "",
		ChatModel:      chat,
		Handlers:       c.hooks,
		Backend:        backend,
		StreamingShell: backend,
		MaxIteration:   50,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		},
		ModelRetryConfig: &adk.ModelRetryConfig{
			// 连接api超时重试配置
			MaxRetries: 5,
			ShouldRetry: func(ctx context.Context, retryCtx *adk.TypedRetryContext[*schema.Message]) *adk.TypedRetryDecision[*schema.Message] {
				// ========== 1. 优先处理ctx取消，一律不重试 ==========
				if ctx.Err() != nil {
					return &adk.TypedRetryDecision[*schema.Message]{
						Retry: false,
					}
				}
				// ========== 2. 错误场景：网络/429限流/5xx临时错误，允许重试 ==========
				if retryCtx.Err != nil {
					errMsg := retryCtx.Err.Error()
					isTransientErr := strings.Contains(errMsg, "429") ||
						strings.Contains(errMsg, "Too Many Requests") ||
						strings.Contains(errMsg, "500") ||
						strings.Contains(errMsg, "502") ||
						strings.Contains(errMsg, "503") ||
						strings.Contains(errMsg, "timeout") ||
						strings.Contains(errMsg, "connection refused")

					if isTransientErr {
						// 指数退避示例：每次重试等待时间 = 2^attempt * 200ms
						backoff := time.Duration(1<<retryCtx.RetryAttempt) * 200 * time.Millisecond
						return &adk.TypedRetryDecision[*schema.Message]{
							Retry:   true,
							Backoff: backoff,
						}
					}
					// 其它错误（400/401/403/内容安全拦截）直接终止重试
					return &adk.TypedRetryDecision[*schema.Message]{Retry: false}
				}

				// ========== 3. 无错误但模型返回结果异常（业务校验失败，可重试） ==========
				// 例如模型返回空内容、tool_call格式非法，可选择重试并修改prompt
				// ========== 修复这里：ToolCall存在时，Content为空是正常的，不要重试 ==========
				msg := retryCtx.OutputMessage
				if msg != nil && msg.Content == "" && len(msg.ToolCalls) == 0 {
					// 只有：文本为空 + 没有工具调用，才判定无效输出
					return &adk.TypedRetryDecision[*schema.Message]{
						Retry: true,
						ModifiedInputMessages: append(
							retryCtx.InputMessages,
							schema.UserMessage("请返回有效内容，不要返回空"),
						),
						PersistModifiedInputMessages: false,
					}
				}

				// ========== 4. 一切正常，不重试 ==========
				return &adk.TypedRetryDecision[*schema.Message]{Retry: false}
			},
		},
	})
	if err != nil {
		return nil, err
	}
	e := &entry{
		build:      b,
		fp:         fp,
		agent:      agent,
		lastAccess: now,
	}
	// 写回缓存，指纹不变时下一轮直接复用
	c.entries[sessionID] = e
	return e, nil
}

// buildFingerprint 对决定 agent 构建结果的输入做摘要：模型接入参数（key/base/model）。
// 工具列表不参与指纹：工具由 hooks 在每轮 BeforeAgent 动态加载，变更无需重建 agent。
func buildFingerprint(raw []byte) (string, error) {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

// // 包含用户最新的输入
func (c *ChatAgent) buildMessage(ctx context.Context, sessionID uint) ([]adk.Message, error) {
	rows, err := c.repo.GetHistory(ctx, sessionID, 40)
	if err != nil {
		return nil, err
	}
	msgs := make([]adk.Message, 0, len(rows))
	for _, r := range rows {
		switch r.Role {
		case "user":
			msgs = append(msgs, schema.UserMessage(r.Content))
		case "assistant":
			// v1 不回放 tool_calls；v2 全保真时改为 schema.AssistantMessage(content, toolCalls)
			msgs = append(msgs, schema.AssistantMessage(r.Content, nil))
		case "tool":
			// v1 不落库 tool 消息，此处仅防御脏数据
			msgs = append(msgs, schema.ToolMessage(r.Content, r.ToolCallID))
		}
	}
	return msgs, nil
}

// drainToolStream 消费一条流式 tool 回执（罕见路径，防御性支持）：拼接全文并取 call_id。
func drainToolStream(sr *schema.StreamReader[*schema.Message]) (string, string) {
	defer sr.Close()
	var content, callID strings.Builder
	for {
		chunk, err := sr.Recv()
		if errors.Is(err, io.EOF) {
			return content.String(), callID.String()
		}
		if err != nil {
			return content.String(), callID.String()
		}
		if chunk == nil {
			continue
		}
		content.WriteString(chunk.Content)
		if chunk.ToolCallID != "" {
			callID.Reset()
			callID.WriteString(chunk.ToolCallID)
		}
	}
}

// Evict 在会话删除后淘汰内存中缓存的 agent，防止 entries 只增不减。
// 正在运行中的 run 不受影响（旧 entry 由运行中的 goroutine 持有，跑完即释放）。
func (c *ChatAgent) Evict(sessionID uint) {
	c.mu.Lock()
	delete(c.entries, sessionID)
	c.mu.Unlock()
}
