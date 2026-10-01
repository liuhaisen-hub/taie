package logger

import (
	"context"
	"log/slog"
)

type AsyncHandler struct {
	inner slog.Handler
	ch    chan slog.Record
}

func NewAsyncHandler(inner slog.Handler, queueSize int) *AsyncHandler {
	h := &AsyncHandler{
		inner: inner,
		ch:    make(chan slog.Record, queueSize),
	}
	go func() {
		for r := range h.ch {
			_ = h.inner.Handle(context.Background(), r)
		}
	}()
	return h
}

func (a *AsyncHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return a.inner.Enabled(ctx, level)
}

func (a *AsyncHandler) Handle(ctx context.Context, r slog.Record) error {
	select {
	case a.ch <- r:
	default:
		// 队列满，丢弃日志，防止阻塞业务
	}
	return nil
}

func (a *AsyncHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &AsyncHandler{inner: a.inner.WithAttrs(attrs), ch: a.ch}
}

func (a *AsyncHandler) WithGroup(name string) slog.Handler {
	return &AsyncHandler{inner: a.inner.WithGroup(name), ch: a.ch}
}
