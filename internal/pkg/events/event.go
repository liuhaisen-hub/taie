package events

import (
	"strconv"
	"time"
)

// EventKind 标识一条流式事件的类别。
type EventKind string

const (
	EventBegin      EventKind = "begin"       // 会话握手：message.id 为会话 ID（services 构造，agentkit 不发）
	EventDelta      EventKind = "delta"       // 正文增量：content=[{type:output_text, text:<增量>}]
	EventToolCall   EventKind = "tool_call"   // 工具调用：content=[function_call]，status=in_progress
	EventToolResult EventKind = "tool_result" // 工具回执：content=[{type:tool_result, call_id, result}]，前端按 call_id 归位
	EventApproval   EventKind = "approval"    // 工具等待审批：content=[approval]，extra.interrupt_id 必须原样回传 agent/approval
	EventDone       EventKind = "done"        // 收口：content=服务端重组的完整 ContentItem[]，status=completed；中断收口时 status=incomplete
	EventError      EventKind = "error"       // 失败：status=failed，content=[{type:output_text, text:<错误文案>}]
)

type Event struct {
	Kind    EventKind `json:"kind"`
	Message Message   `json:"message"`
}

// Begin 会话就绪握手：session_id=0 发出首问时，后端建会话后第一件事回推本事件，
// 会话 ID 放在 message.id 里。前端收到后记下来，后续请求全部回传。
// services 层构造，agentkit 不发。
func Begin(sessionID uint) Event {
	return Event{Kind: EventBegin, Message: Message{ID: strconv.FormatUint(uint64(sessionID), 10)}}
}

// NewAssistantMessage 生成一条 assistant 消息骨架。assistantID 由 agentkit 在
// Runner 入口生成（落库在结束时才发生，拿不到自增主键），本轮事件全程携带，
// 前端据此把占位气泡归并到同一条消息上。
func NewAssistantMessage(assistantID string) Message {
	return Message{
		Role:       RoleAssistant,
		ID:         assistantID,
		CreateTime: time.Now().UnixMilli(),
		Status:     StatusLoading,
	}
}

// Delta 正文增量：text 追加到前端最后一个 output_text item。
func Delta(assistantID, chunk string) Event {
	msg := NewAssistantMessage(assistantID)
	msg.Content = []ContentItem{{Type: ItemOutputText, Text: chunk}}
	return Event{Kind: EventDelta, Message: msg}
}

// ToolCall 工具调用开始：前端追加 function_call item（自定义状态卡渲染）。
// 工具入参不回传，前端只展示"正在调用xx工具"。
func ToolCall(assistantID, callID, name string) Event {
	msg := NewAssistantMessage(assistantID)
	msg.Content = []ContentItem{{
		Type: ItemFunctionCall, CallID: callID, Name: name, Status: StatusLoading,
	}}
	return Event{Kind: EventToolCall, Message: msg}
}

// ToolResult 工具回执：前端按 call_id 把对应 function_call 置 completed，
// 并把 result 追加为 tool_result item（自定义渲染折叠卡）。
func ToolResult(assistantID, callID, result string) Event {
	msg := NewAssistantMessage(assistantID)
	msg.Content = []ContentItem{{Type: ItemToolResult, CallID: callID, Result: result}}
	return Event{Kind: EventToolResult, Message: msg}
}

// Approval 工具等待人工审批：卡片项 name=工具名、result=入参 JSON、
// extra.interrupt_id 是续跑定位键（eino InterruptCtx.ID），
// 前端裁决后必须原样回传 agent/approval。
func Approval(assistantID, toolName, argsJSON, interruptID string) Event {
	msg := NewAssistantMessage(assistantID)
	msg.Content = []ContentItem{{
		Type:   ItemApproval,
		Name:   toolName,
		Result: argsJSON,
		Status: StatusLoading,
		Extra:  map[string]any{"interrupt_id": interruptID},
	}}
	return Event{Kind: EventApproval, Message: msg}
}

// Done 正常收口：content 为服务端重组的完整 ContentItem[]（与事件流一致的最终态），
// 前端整体替换气泡内容。
func Done(assistantID string, items []ContentItem) Event {
	msg := NewAssistantMessage(assistantID)
	msg.Status = StatusDone
	msg.Content = items
	return Event{Kind: EventDone, Message: msg}
}

// Interrupted 中断收口（ctx 取消，如用户点停止/窗口关闭）：携带已累积 items，
// status=incomplete，前端据此把气泡置 cancelled。连接可能已关闭，发射失败静默。
func Interrupted(assistantID string, items []ContentItem) Event {
	msg := NewAssistantMessage(assistantID)
	msg.Status = StatusIncomplete
	msg.Content = items
	return Event{Kind: EventDone, Message: msg}
}

// Error 错误收口：前端把气泡置 failed，错误文案作为 output_text 追加展示。
func Error(text string) Event {
	msg := NewAssistantMessage("")
	msg.Status = StatusError
	msg.Content = []ContentItem{{Type: ItemOutputText, Text: text}}
	return Event{Kind: EventError, Message: msg}
}
