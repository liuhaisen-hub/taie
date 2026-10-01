package trace

import (
	"context"
	"fmt"
	"log/slog"
	"taie/internal/agentkit/rctx"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/schema"
)

type Tracer struct {
	logger *slog.Logger
	prefix string
}

// 返回具体类型 *Tracer 而非接口，wire 才能把它注入到 NewChatAgent
func NewLogCollectHandler(logger *slog.Logger) *Tracer {
	callbacks.AppendGlobalHandlers()
	return &Tracer{
		logger: logger,
		prefix: "[AGENT_LOG]",
	}
}

// 实现以下接口
// type Handler interface {
//     OnStart(ctx context.Context, info *RunInfo, input CallbackInput) context.Context
//     OnEnd(ctx context.Context, info *RunInfo, output CallbackOutput) context.Context

//     OnError(ctx context.Context, info *RunInfo, err error) context.Context

//	    OnStartWithStreamInput(ctx context.Context, info *RunInfo,
//	        input *schema.StreamReader[CallbackInput]) context.Context
//	    OnEndWithStreamOutput(ctx context.Context, info *RunInfo,
//	        output *schema.StreamReader[CallbackOutput]) context.Context
//	}
//
// 非流失
func (t *Tracer) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	sessionID, _ := rctx.GetSessionID(ctx)

	msg := fmt.Sprintf("%s START sessionId: %d", t.prefix, sessionID)
	t.logger.Info(msg,
		"component", info.Name,
		"type", info.Type,
		"input", input,
	)
	return ctx
}

// 非流失
func (t *Tracer) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	sessionID, _ := rctx.GetSessionID(ctx)
	msg := fmt.Sprintf("%s END sessionId: %d", t.prefix, sessionID)
	t.logger.Info(msg,
		"component", info.Name,
		"type", info.Type,
		"output", output,
	)
	return ctx
}

func (t *Tracer) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	sessionID, _ := rctx.GetSessionID(ctx)
	msg := fmt.Sprintf("%s ERROR sessionId: %d", t.prefix, sessionID)
	t.logger.Error(msg,
		"component", info.Name,
		"type", info.Type,
		"error", err,
	)
	return ctx
}

func (t *Tracer) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo,
	input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	sessionID, _ := rctx.GetSessionID(ctx)
	msg := fmt.Sprintf("%s STREAM_INPUT START sessionId %d", t.prefix, sessionID)
	t.logger.Info(msg,
		"component", info.Name,
		"type", info.Type,
	)
	// input 是流，不建议直接打印，会消费流；如需打印可以单独goroutine读取
	return ctx
}
func (t *Tracer) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	sessionID, _ := rctx.GetSessionID(ctx)
	msg := fmt.Sprintf("%s STREAM_OUTPUT END sessionId %d", t.prefix, sessionID)
	t.logger.Info(msg,
		"component", info.Name,
		"type", info.Type,
	)
	// output 是流，不要直接打印，会吃掉流数据
	return ctx
}
