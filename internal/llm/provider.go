package llm

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// Provider 抽象 LLM 调用接口
// 所有 LLM 供应商都实现这个接口，Agent 只依赖接口不依赖具体实现
type Provider interface {
	// Generate 发送消息并获取回复
	// ctx 用于超时控制
	// messages 对话历史（含 system + user messages）
	Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error)

	// GenerateStream 流式生成回复
	// 返回 StreamReader，逐 token 读取
	GenerateStream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error)
}

// ProviderType 支持的供应商类型
type ProviderType string

const (
	ProviderDeepSeek ProviderType = "deepseek"
	ProviderOpenAI   ProviderType = "openai"
)
