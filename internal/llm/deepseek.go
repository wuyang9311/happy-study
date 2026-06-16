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

// 用于缓存通过 context 创建的 model 实例
type contextModelCache struct{}

// getModelForCall 根据 context 返回合适的 model（可能基于用户偏好）
func (p *DeepSeekProvider) getModelForCall(ctx context.Context) *openai.ChatModel {
	userModel := UserModelFromContext(ctx)
	if userModel == "" {
		return p.model
	}
	// 如果用户模型和当前模型一致，直接用
	if userModel == p.config.Model {
		return p.model
	}
	// 否则创建新 model
	m, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		Model:   userModel,
		APIKey:  p.config.APIKey,
		BaseURL: p.config.BaseURL,
	})
	if err != nil {
		return p.model // 回退
	}
	return m
}

// Generate 发送消息到 DeepSeek 并获取回复（支持 context 中的用户模型覆盖）
func (p *DeepSeekProvider) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	m := p.getModelForCall(ctx)
	resp, err := m.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("deepseek generate: %w", err)
	}
	return resp, nil
}

// GenerateStream 流式生成（支持 context 中的用户模型覆盖）
func (p *DeepSeekProvider) GenerateStream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	m := p.getModelForCall(ctx)
	stream, err := m.Stream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("deepseek stream: %w", err)
	}
	return stream, nil
}

// WithModel 创建使用指定模型的 DeepSeekProvider 副本
func (p *DeepSeekProvider) WithModel(model string) Provider {
	cfg := *p.config
	cfg.Model = model
	clone := &DeepSeekProvider{
		config: &cfg,
	}
	// 懒初始化：等第一次 Generate/GenerateStream 调用时再创建 model
	// 但为了通用性，还是直接创建
	newModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		Model:   cfg.Model,
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
	})
	if err == nil {
		clone.model = newModel
	} else {
		clone.model = p.model // 回退到原模型
	}
	return clone
}
