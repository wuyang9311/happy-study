package llm

import (
	"context"
	"fmt"
	"os"
)

// Config LLM 供应商配置
type Config struct {
	// Provider 供应商类型：deepseek / openai
	Provider ProviderType `json:"provider"`

	// Model 模型名称
	Model string `json:"model"`

	// APIKey API 密钥
	APIKey string `json:"api_key"`

	// BaseURL API 地址（可选）
	BaseURL string `json:"base_url"`
}

// DefaultConfig 从环境变量读取配置
// 优先读取 DEEPSEEK_* 系列变量，兼容旧配置
func DefaultConfig() *Config {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	model := os.Getenv("DEEPSEEK_MODEL")

	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	if baseURL == "" {
		baseURL = os.Getenv("OPENAI_BASE_URL")
	}

	provider := ProviderDeepSeek
	if model == "" {
		model = "deepseek-chat"
	}

	return &Config{
		Provider: provider,
		Model:    model,
		APIKey:   apiKey,
		BaseURL:  baseURL,
	}
}

// Validate 验证配置是否完整
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required: set DEEPSEEK_API_KEY or OPENAI_API_KEY")
	}
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	return nil
}

// NewProvider 根据配置创建 Provider（自动包装重试）
func NewProvider(ctx context.Context, config *Config) (Provider, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	var inner Provider
	switch config.Provider {
	case ProviderDeepSeek:
		p, err := NewDeepSeekProvider(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("create deepseek provider: %w", err)
		}
		inner = p
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	// 包装重试装饰器
	return NewRetryProvider(inner, nil), nil
}
