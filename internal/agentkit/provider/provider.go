package provider

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
)

type BuildParams struct {
	ApiKey      string
	ModalName   string
	BaseUrl     string
	MaxMemories int32
}

// 将来通过这里适配不同的模型接入
// 在eino 中chatmodel 是无状态的
func NsewOpenAIChatModel(ctx context.Context, b *BuildParams) (*openai.ChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: b.BaseUrl,
		APIKey:  b.ApiKey,
		Model:   b.ModalName,
	})
}
