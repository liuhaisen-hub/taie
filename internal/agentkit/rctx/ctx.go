package rctx

import "context"

type ctxKey string

const bizKey ctxKey = "taie_biz_code"
const sessionkey ctxKey = "taie_session"
const messagekey ctxKey = "taie_message"
const modelkey ctxKey = "taie_model_name"

func WithBizCode(ctx context.Context, bizCode string) context.Context {
	return context.WithValue(ctx, bizKey, bizCode)
}

// GetBizCode 从ctx取出业务标识
func GetBizCode(ctx context.Context) string {
	val, ok := ctx.Value(bizKey).(string)
	if !ok {
		return ""
	}
	return val
}

// 对外暴露Set方法，业务层用来把sessionID、messageID塞到ctx
func WithSessionId(ctx context.Context, sid int64) context.Context {
	return context.WithValue(ctx, sessionkey, sid)
}

func GetSessionID(ctx context.Context) (int64, bool) {
	val, ok := ctx.Value(sessionkey).(int64)
	return val, ok
}

func WithMessageID(ctx context.Context, mid int64) context.Context {
	return context.WithValue(ctx, messagekey, mid)
}

func GetMessageID(ctx context.Context) (int64, bool) {
	val, ok := ctx.Value(messagekey).(int64)
	return val, ok
}

func WithModelName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, modelkey, name)
}

func GetModelName(ctx context.Context) string {
	val, _ := ctx.Value(modelkey).(string)
	return val
}
