package agent

import "os"

// LLMConfig 大模型配置
type LLMConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// DefaultLLMConfig 从环境变量读取默认配置
func DefaultLLMConfig() *LLMConfig {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	model := os.Getenv("DEEPSEEK_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}
	return &LLMConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
	}
}
