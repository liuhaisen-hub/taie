package hooks

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/cloudwego/eino/adk"
)

const projectKey = "taie_project"

// WithBizCode 将业务标识存入ctx
func WithProjectCode(ctx context.Context, projectCode string) context.Context {
	return context.WithValue(ctx, projectKey, projectCode)
}

// GetBizCode 从ctx取出业务标识
func GetProjectCode(ctx context.Context) string {
	val, ok := ctx.Value(projectKey).(string)
	if !ok {
		return ""
	}
	return val
}

// staticSectionPrompt 静态段是常量，进程内只构建一次；
// 动态段（env）依赖业务上下文，每次 run 现算
var staticSectionPrompt = sync.OnceValue(getStaticSection)

// 注入系统提示词，静态 + 动态 （规划环境，mcp列表描述，不同的业务模式，操作系统，shell环境等按需添加）
// 用户提示词等。middleware 无状态：同一批实例挂在根 agent 和所有 task 子代理上，
// 缓存结构体字段会互相踩。
type dynamicPromptToolMiddleware struct {
	*adk.BaseChatModelAgentMiddleware
}

func (m *dynamicPromptToolMiddleware) BeforeAgent(ctx context.Context, runCtx *adk.ChatModelAgentContext) (context.Context, *adk.ChatModelAgentContext, error) {
	// 根据不同的业务产生不同的系统提示词
	// bizCode := rctx.GetBizCode(ctx)
	sysPrompt := staticSectionPrompt()
	if env := buildEnv(); env != "" {
		sysPrompt += systemPromptDynamicBoundary + env
	}
	// 追加而不是覆盖：deep 内置 handler 已往 Instruction 写入基础指令与
	// task 工具说明，覆盖会全部丢失；execContext 每 run 从配置重建，追加不会跨轮累积
	runCtx.Instruction = sysPrompt + "\n\n" + runCtx.Instruction
	return ctx, runCtx, nil
}

// SystemPromptDynamicBoundary 是动态内容的切分标记。
// 是否出现在该标记之下；出现在之上的即视为违规。
const systemPromptDynamicBoundary = "<<<SYSTEM_PROMPT_DYNAMIC_BOUNDARY>>>"

type StaticSection struct {
	Role        string // 角色定义
	Operational string // 系统规范
	Philosophy  string //回答「什么叫把事做完」——例如：聚焦用户请求、反对范围蔓延、反对为炫技而抽象。
	Risk        string // 风险提示
	Tools       string //工具规范
	Tone        string // 语气语调
}

func (s StaticSection) BuildStaticSection() string {
	var sb strings.Builder
	sb.WriteString(s.Role)
	sb.WriteString("\n\n## 工作方式\n")
	sb.WriteString(s.Operational)
	sb.WriteString("\n\n## 做事理念\n")
	sb.WriteString(s.Philosophy)
	sb.WriteString("\n\n## 风险提示\n")
	sb.WriteString(s.Risk)
	sb.WriteString("\n\n## 工具使用规范\n")
	sb.WriteString(s.Tools)
	sb.WriteString("\n\n## 语气语调\n")
	sb.WriteString(s.Tone)
	return sb.String()

}

// 后续传入bizcode 动态构建
func getStaticSection() string {
	// 当前只有一种业务，后续可拓展。先定架构
	section := StaticSection{
		Role: "你是一个全能、靠谱的通用智能助手，能够回答知识问题、提供思路建议、撰写文案、分析信息、答疑解惑。你会理解用户意图，兼顾深度和易懂性",
		Operational: `
		  - 以用户当前提问为核心，围绕需求作答，不随意跑偏；
          - 用户需求模糊时，优先简短提问澄清，不要自行脑补额外需求；
          - 输出结构随场景自适应：简单问题简短回答，复杂问题整理成任务列表，一项项的执行；
          - 不主动追加用户没有要求的额外功能、拓展话题；
          - 严格遵守内容安全准则，拒绝违规、有害请求。
	  `,
		Philosophy: `
	     - 优先解决用户当下真正想要解决的问题，拒绝范围蔓延；
         - 不刻意堆砌专业术语、不炫技，不做多余的复杂表达；
         - 够用就好，优先交付可用结果，不提前假设用户未来的需求；
         - 交付标准：满足用户原始诉求，不画蛇添足增加无关内容。
	  `,
		Risk: `
		- 禁止输出和隐私相关的信息
		- 涉及法律，医疗，资金投资等内容，需申明给出的结论仅供参考
		- 拒绝高危险操作，如自行调用工具进行操作
		- 涉及人身安全的问题，拒绝回答
	  `,
		Tools: `
		- 通过工具获取的结果必须真实的返回给用户，不得编造事实
		- 工具返回的内容需整理清晰，梳理清楚结果，不能直接输出给用户
	  `,
		Tone: `
	   - 亲和自然，理性克制
	   - 不生硬，不过度客套
	   - 如果是用户输入错误，需温和指正
	  `,
	}
	return section.BuildStaticSection()
}

// 当前只有一种agent，所以只有一个

type EnvSpace struct {
	workdir string // 工作目录
	os      string //操作系统
	shell   string // shell环境
}

// 后续传入bizcode 动态构建
func buildEnv() string {
	env := &EnvSpace{
		workdir: "",
		os:      "",
		shell:   "",
	}
	var sb strings.Builder

	if env.workdir != "" && env.shell != "" && env.os != "" {
		fmt.Fprintf(&sb, "\n\n ## 工作目录信息\n操作系统是：%s; shell环境是：%s; 工作目录是：%s;", env.os, env.shell, env.workdir)
	}
	return sb.String()
}
