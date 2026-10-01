package events

// 角色常量：wire 层唯一词汇来源，禁止散落魔法字符串
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Message.Status —— 与 Semi AIChatDialogue / OpenAI Response API 的 status 枚举
// 原生取值对齐（in_progress / completed / failed / incomplete），前端收到事件
// 零转换即可塞进气泡。
const (
	StatusLoading    = "in_progress"
	StatusDone       = "completed"
	StatusError      = "failed"
	StatusIncomplete = "incomplete" // 中断（stop/断连）收口
)

// ContentItem.Type —— OpenAI Response Object 子集（即 Semi ContentItem 的原生
// 形态）。output_text 仅作为 message item 的内层元素或 delta 事件的传输载荷出现，
// 不直接放 Message.content 顶层（Semi 无对应内置渲染器）；tool_result 为本协议
// 自定义类型，前端经 renderDialogueContentItem 渲染结果折叠卡。
const (
	ItemMessage      = "message"
	ItemOutputText   = "output_text"
	ItemFunctionCall = "function_call"
	ItemToolResult   = "tool_result" // 自定义
	ItemApproval     = "approval"    // 自定义：工具等待人工审批卡片
)

// MessageItem 把正文包裹成 message item（Semi 内置渲染形态：
// 裸 output_text 放 Message.content 顶层不会被渲染）。
func MessageItem(text string) ContentItem {
	return ContentItem{Type: ItemMessage, Content: []ContentItem{{Type: ItemOutputText, Text: text}}}
}

// ContentItem 是 OpenAI Response Object 子集，字段名与 Semi ContentItem 一致，
// 前端收到后直接作为 Message.content 数组元素塞进 AIChatDialogue，零转换。
// 注意：Semi 只内置渲染 message/reasoning/function_call/custom_tool_call 四种
// 顶层 item，正文必须包成 {type:message, content:[{type:output_text,...}]}，
// 裸 output_text 放顶层不会被渲染。
type ContentItem struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
	Role   string `json:"role,omitempty"`
	// Text 为 output_text 的文本：delta 事件里是增量，done 事件里是全文
	Text string `json:"text,omitempty"`
	// message item 的内层元素（output_text 等），正文经此包裹后 Semi 才渲染
	Content []ContentItem `json:"content,omitempty"`
	// function_call / tool_result 专用（工具入参不回传，前端不展示）
	CallID string `json:"call_id,omitempty"`
	Name   string `json:"name,omitempty"`
	Result string `json:"result,omitempty"`
	// 预留
	Extra map[string]any `json:"extra,omitempty"`
}

// Message 对齐 Semi AIChatDialogue 的 Message：content 统一为 ContentItem 数组，
// 前端按 item.type 归并（output_text 增量追加、function_call/tool_result 追加卡片）。
type Message struct {
	Role       string        `json:"role"`
	ID         string        `json:"id,omitempty"`
	CreateTime int64         `json:"createdAt,omitempty"` // Semi Message.createdAt
	Status     string        `json:"status,omitempty"`
	Content    []ContentItem `json:"content,omitempty"`
	References []Reference   `json:"references,omitempty"`
}

type Reference struct {
	Title   string `json:"title"`
	URL     string `json:"url,omitempty"`
	Content string `json:"content"`
}
