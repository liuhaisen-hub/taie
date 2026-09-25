package chat

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
)

// 后续做兼容层
func NewOpenAiChatModel(ctx context.Context, apiKey, baseUrl, modelName string) (*openai.ChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		BaseURL: baseUrl,
		Model:   modelName,
	})
}
