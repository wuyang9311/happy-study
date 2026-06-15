package llm

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// DeepSeekProvider DeepSeek LLM 供应商实现
type DeepSeekProvider struct {
	model  *openai.ChatModel
	config *Config
}

// NewDeepSeekProvider 创建 DeepSeek provider
func NewDeepSeekProvider(ctx context.Context, config *Config) (*DeepSeekProvider, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid deepseek config: %w", err)
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   config.Model,
		APIKey:  config.APIKey,
		BaseURL: baseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create deepseek chat model: %w", err)
	}

	return &DeepSeekProvider{
		model:  chatModel,
		config: config,
	}, nil
}

// Generate 发送消息到 DeepSeek 并获取回复
func (p *DeepSeekProvider) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	resp, err := p.model.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("deepseek generate: %w", err)
	}
	return resp, nil
}

// GenerateStream 流式生成
func (p *DeepSeekProvider) GenerateStream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	stream, err := p.model.Stream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("deepseek stream: %w", err)
	}
	return stream, nil
}
