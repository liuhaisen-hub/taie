package events

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"time"

	"github.com/cloudwego/eino/schema"
)

// Emitter 收口一轮问答的事件发射：持有本轮 assistant 消息 ID 与按发生顺序累积的
// 内容序列。delta/tool_call/tool_result 即时下发，done 时重组完整 ContentItem[]
// 作为最终态整体下发，保证与事件流一致。
type Emitter struct {
	assistantID string
	onEvent     func(Event) error

	textBuf     strings.Builder      // 正文全文（与落库同源）
	seq         []ContentItem        // 内容序列：output_text 全文项 + function_call + tool_result
	callIdx     map[string]int       // call_id -> seq 中 function_call item 下标
	streamText  strings.Builder      // 当前流式消息的正文（EOF 时并入 seq）
	streamIdx   map[int]*toolCallAcc // 当前流式消息的工具调用分片聚合
	streamOrder []int                // 分片出现顺序，保证下发顺序稳定
}

// toolCallAcc 聚合一条工具调用的流式分片：eino 的 chunk.ToolCalls 按 Index 分片，
// Name/ID 可能只在首个分片出现。入参分片直接丢弃（不回传前端）。
type toolCallAcc struct {
	id   string
	name string
}

func NewEmitter(sessionID uint, onEvent func(Event) error) *Emitter {
	return &Emitter{
		assistantID: fmt.Sprintf("assistant-%d-%d", sessionID, time.Now().UnixNano()),
		onEvent:     onEvent,
		callIdx:     make(map[string]int),
		streamIdx:   make(map[int]*toolCallAcc),
	}
}

func (e *Emitter) emit(ev Event) error { return e.onEvent(ev) }

// delta 流式正文增量：即时下发并累积。
func (e *Emitter) delta(chunk string) error {
	e.streamText.WriteString(chunk)
	e.textBuf.WriteString(chunk)
	return e.emit(Delta(e.assistantID, chunk))
}

// assistantMessage 非流式兜底：正文一次性下发；附带的 ToolCalls 防御性下发。
func (e *Emitter) AssistantMessage(msg *schema.Message) error {
	if msg.Content != "" {
		e.textBuf.WriteString(msg.Content)
		// 正文包成 message item，Semi 只内置渲染这种形态（裸 output_text 不显示）
		e.seq = append(e.seq, MessageItem(msg.Content))
		if err := e.emit(Delta(e.assistantID, msg.Content)); err != nil {
			return err
		}
	}
	for i, tc := range msg.ToolCalls {
		callID := tc.ID
		if callID == "" {
			callID = fmt.Sprintf("%s-%d", tc.Function.Name, i)
		}
		if err := e.emitToolCall(callID, tc.Function.Name); err != nil {
			return err
		}
	}
	return nil
}

// emitToolCall 把一条完整的工具调用写入序列并下发（status in_progress，
// 等对应 tool_result 到达后置 completed）。入参不写入序列、不下发。
func (e *Emitter) emitToolCall(callID, name string) error {
	e.callIdx[callID] = len(e.seq)
	e.seq = append(e.seq, ContentItem{
		Type: ItemFunctionCall, CallID: callID,
		Name: name, Status: StatusLoading,
	})
	return e.emit(ToolCall(e.assistantID, callID, name))
}

// streamAcc 取当前流中第 idx 个工具调用分片的聚合器。
func (e *Emitter) streamAcc(idx int) *toolCallAcc {
	acc, ok := e.streamIdx[idx]
	if !ok {
		acc = &toolCallAcc{}
		e.streamIdx[idx] = acc
		e.streamOrder = append(e.streamOrder, idx)
	}
	return acc
}

// streamEOF 一条流式消息消费完毕：正文并入 seq，工具调用统一下发（eino 流式下
// 工具调用与正文分片混排，EOF 时 Name/ID 才拼装完整）。
func (e *Emitter) streamEOF() error {
	if e.streamText.Len() > 0 {
		// 正文包成 message item，Semi 只内置渲染这种形态（裸 output_text 不显示）
		e.seq = append(e.seq, MessageItem(e.streamText.String()))
		e.streamText.Reset()
	}
	for _, idx := range e.streamOrder {
		acc := e.streamIdx[idx]
		callID := acc.id
		if callID == "" {
			callID = fmt.Sprintf("%s-%d", acc.name, idx)
		}
		if err := e.emitToolCall(callID, acc.name); err != nil {
			return err
		}
	}
	// 重置流内聚合，迎接下一轮 assistant 消息
	e.streamIdx = make(map[int]*toolCallAcc)
	e.streamOrder = nil
	return nil
}

// toolResult 工具回执：按 call_id 把对应 function_call 置 completed，
// 回执内容追加为 tool_result item。
func (e *Emitter) ToolResult(callID, result string) error {
	if callID == "" {
		return nil // 无 call_id 无法归位，防御脏数据
	}
	if idx, ok := e.callIdx[callID]; ok {
		e.seq[idx].Status = StatusDone
	}
	e.seq = append(e.seq, ContentItem{Type: ItemToolResult, CallID: callID, Result: result})
	return e.emit(ToolResult(e.assistantID, callID, result))
}

// consumeStream 消费一条流式 assistant 消息：正文增量即时下发，工具调用分片按
// *Index 合并、EOF 后统一下发。流必须恰好 Close 一次。
func (e *Emitter) ConsumeStream(sr *schema.StreamReader[*schema.Message]) error {
	defer sr.Close() // 无论读完还是提前退出，都只 Close 这一次
	for {
		chunk, err := sr.Recv()
		if errors.Is(err, io.EOF) {
			return e.streamEOF()
		}
		if err != nil {
			return err
		}
		if chunk == nil {
			continue
		}
		if chunk.Content != "" {
			// 空 Content 是工具调用参数分片等非正文数据
			if err := e.delta(chunk.Content); err != nil {
				return err
			}
		}
		for i := range chunk.ToolCalls {
			tc := chunk.ToolCalls[i]
			idx := 0
			if tc.Index != nil {
				idx = *tc.Index
			}
			acc := e.streamAcc(idx)
			if tc.ID != "" {
				acc.id = tc.ID
			}
			if tc.Function.Name != "" {
				acc.name = tc.Function.Name
			}
			// tc.Function.Arguments 入参分片不聚合，本协议不回传前端
		}
	}
}

// Text 返回正文全文，供落库（与 Done 事件的 output_text 同源）。
func (e *Emitter) Text() string { return e.textBuf.String() }

// done 正常收口：重组完整内容序列下发。
func (e *Emitter) Done() error {
	return e.emit(Done(e.assistantID, e.seq))
}

// interrupted 中断收口：尽力下发 incomplete 状态（连接可能已关闭，失败静默）。
func (e *Emitter) Interrupted() error {
	_ = e.streamEOF() // 残留正文/分片尽量入列
	return e.emit(Interrupted(e.assistantID, e.seq))
}

// Approval 工具等待人工审批：卡片项追加进内容序列并下发。
// 发出后本轮执行挂起（中断是迭代器最后一个事件），不再有 done；
// 前端裁决后走 agent/approval 新连接续跑。
func (e *Emitter) Approval(toolName, argsJSON, interruptID string) error {
	item := ContentItem{
		Type:   ItemApproval,
		Name:   toolName,
		Result: argsJSON,
		Status: StatusLoading,
		Extra:  map[string]any{"interrupt_id": interruptID},
	}
	e.seq = append(e.seq, item)
	return e.emit(Approval(e.assistantID, toolName, argsJSON, interruptID))
}
