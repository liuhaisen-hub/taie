package hooks

import (
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// BuildHooks 返回挂到 deep agent 的全部 middleware。
// 注意：同一批实例会同时挂在根 agent 和 task 子代理上，
// middleware 必须无状态，跨 run 数据放 ctx（见 toolscall.go / rctx）。
func BuildHooks(repo UsageRepo, perm PermissionRepo) []adk.TypedChatModelAgentMiddleware[*schema.Message] {
	return []adk.ChatModelAgentMiddleware{
		&safeToolMiddleware{},
		&dynamicPromptToolMiddleware{},
		// 审批中间件放在 safeTool 之内（注册序在后 = 包装更内层）：
		// StatefulInterrupt 的中断错误要经 safeTool 的 IsInterruptRerunError
		// 直通分支原样穿透，不能被它转成 "[tool error] ..." 字符串
		newApprovalMiddleware(perm),
		&usageMiddleware{
			repo: repo,
		},
	}
}
