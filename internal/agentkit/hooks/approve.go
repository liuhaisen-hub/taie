package hooks

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"taie/internal/po"
)

// PermissionRepo 权限中间件依赖的最小接口，由 data 层实现（wire 注入）。
// 接口定义在本包、实现在 data：依赖方向 data → agentkit 单向，
// hooks 禁止 import data（跨层用接口注入，同 UsageRepo 的做法）。
type PermissionRepo interface {
	GetMode() (string, error)
	GetPermission() (*po.PermissionSetting, error)
}

// ApprovalInfo 首跑中断时放进 checkpoint 的审批请求（用户可见的中断信息）。
type ApprovalInfo struct {
	ToolName        string // 被拦截的工具名
	ArgumentsInJSON string // 原始入参 JSON，续跑直接用，不依赖模型重发
}

// ApprovalResult 审批裁决。ResumeWithParams 按 interrupt_id 投递，
// 工具续跑时经 GetResumeContext 取回。
type ApprovalResult struct {
	Approved         bool
	DisapproveReason *string // 拒绝理由，Approved=false 时生效
}

func init() {
	// checkpoint 是 gob 编码：进 checkpoint 的自定义类型必须注册，
	// 否则不在当下报错，而是 Resume 解码时才失败（官方 approval_wrapper 同款）
	schema.Register[*ApprovalInfo]()
	schema.Register[*ApprovalResult]()
}

// approvalMiddleware 工具执行前的权限审批中间件。
// 无状态：同一批实例挂在根 agent 和所有 task 子代理上、跨会话共享，
// 权限判定所需的 mode/开关每次工具调用实时从 repo 读（system.json 很小，
// Save 是 tmp+rename 原子替换，读安全；设置页改开关即时生效）。
type approvalMiddleware struct {
	*adk.BaseChatModelAgentMiddleware
	repo PermissionRepo
}

func newApprovalMiddleware(repo PermissionRepo) adk.ChatModelAgentMiddleware {
	return &approvalMiddleware{repo: repo}
}

// needsApproval 判定这次执行是否需要人工审批（规则见文档 §1.1）：
// 不在清单 → 放行；auto → 放行；plan → 写类必审、读类放行；
// manual → 开关 true 放行 / false 审批；配置读不到 → 清单内一律审批（fail-closed）。
func (m *approvalMiddleware) needsApproval(name string) bool {
	if !po.IsPermissionTool(name) {
		return false // web_search / write_todos / task / ls 等不受权限管辖
	}
	mode, err := m.repo.GetMode()
	if err != nil {
		return true
	}
	switch mode {
	case po.ModeAuto:
		return false
	case po.ModePlan:
		return po.IsWriteTool(name)
	default: // manual
		perm, err := m.repo.GetPermission()
		if err != nil {
			return true
		}
		allowed, ok := perm.Lookup(name)
		if !ok {
			return false
		}
		return !allowed
	}
}

// WrapInvokableToolCall 拦截同步工具（read_file/write_file/edit_file/glob/grep）。
// eino 在"工具即将执行"的请求时刻回调，tCtx.Name 是工具名。
func (m *approvalMiddleware) WrapInvokableToolCall(_ context.Context,
	endpoint adk.InvokableToolCallEndpoint, tCtx *adk.ToolContext,
) (adk.InvokableToolCallEndpoint, error) {
	if !m.needsApproval(tCtx.Name) {
		return endpoint, nil // 放行：原样返回，不加任何包装
	}
	return func(ctx context.Context, args string, opts ...tool.Option) (string, error) {
		// GetInterruptState：本次执行是首跑(false)还是中断后续跑(true)；
		// state 是首跑 StatefulInterrupt 存下的入参，续跑直接用
		wasInterrupted, _, storedArgs := tool.GetInterruptState[string](ctx)
		if !wasInterrupted {
			// 首跑：发起中断。整个 agent 在工具执行前停下，runner 落 checkpoint，
			// 事件流以 Action.Interrupted 收口，前端据此弹审批卡片
			return "", tool.StatefulInterrupt(ctx, &ApprovalInfo{
				ToolName:        tCtx.Name,
				ArgumentsInJSON: args,
			}, args)
		}

		// 续跑：ResumeWithParams.Targets 以 interrupt_id 投递的裁决出现在这里
		isTarget, hasData, result := tool.GetResumeContext[*ApprovalResult](ctx)
		if isTarget && hasData {
			if result.Approved {
				// 批准：用存下的原始入参执行真工具
				return endpoint(ctx, storedArgs, opts...)
			}
			// 拒绝：把理由作为工具结果（字符串而非 error）喂回模型，模型会改口
			if result.DisapproveReason != nil {
				return fmt.Sprintf("工具 %s 被用户拒绝，原因：%s", tCtx.Name, *result.DisapproveReason), nil
			}
			return fmt.Sprintf("工具 %s 被用户拒绝", tCtx.Name), nil
		}

		// 非本轮恢复目标（同批多工具先后审批，只批了别的那个）：
		// 原样再中断保持状态，等下一轮 ResumeWithParams 指到它
		isTarget, _, _ = tool.GetResumeContext[any](ctx)
		if !isTarget {
			return "", tool.StatefulInterrupt(ctx, &ApprovalInfo{
				ToolName:        tCtx.Name,
				ArgumentsInJSON: storedArgs,
			}, storedArgs)
		}

		// 是本轮目标但未携带数据（隐式恢复）：默认放行
		return endpoint(ctx, storedArgs, opts...)
	}, nil
}

// WrapStreamableToolCall 拦截流式工具。taie 配了 StreamingShell，
// execute 注册的是 StreamableTool，只写 Invokable 版拦不住它（见文档 §0）。
// 状态机与 WrapInvokableToolCall 完全同构，仅返回类型不同。
func (m *approvalMiddleware) WrapStreamableToolCall(_ context.Context,
	endpoint adk.StreamableToolCallEndpoint, tCtx *adk.ToolContext,
) (adk.StreamableToolCallEndpoint, error) {
	if !m.needsApproval(tCtx.Name) {
		return endpoint, nil
	}
	return func(ctx context.Context, args string, opts ...tool.Option) (*schema.StreamReader[string], error) {
		wasInterrupted, _, storedArgs := tool.GetInterruptState[string](ctx)
		if !wasInterrupted {
			return nil, tool.StatefulInterrupt(ctx, &ApprovalInfo{
				ToolName:        tCtx.Name,
				ArgumentsInJSON: args,
			}, args)
		}

		isTarget, hasData, result := tool.GetResumeContext[*ApprovalResult](ctx)
		if isTarget && hasData {
			if result.Approved {
				return endpoint(ctx, storedArgs, opts...)
			}
			if result.DisapproveReason != nil {
				return singleChunkReader(fmt.Sprintf("工具 %s 被用户拒绝，原因：%s", tCtx.Name, *result.DisapproveReason)), nil
			}
			return singleChunkReader(fmt.Sprintf("工具 %s 被用户拒绝", tCtx.Name)), nil
		}

		isTarget, _, _ = tool.GetResumeContext[any](ctx)
		if !isTarget {
			return nil, tool.StatefulInterrupt(ctx, &ApprovalInfo{
				ToolName:        tCtx.Name,
				ArgumentsInJSON: storedArgs,
			}, storedArgs)
		}

		return endpoint(ctx, storedArgs, opts...)
	}, nil
}

// singleChunkReader 把整段文本包成单 chunk 只读流（拒绝回执走流式出口用）
func singleChunkReader(s string) *schema.StreamReader[string] {
	sr, sw := schema.Pipe[string](1)
	sw.Send(s, nil)
	sw.Close()
	return sr
}
