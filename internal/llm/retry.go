package llm

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/cloudwego/eino/schema"
)

// RetryConfig 重试策略配置
type RetryConfig struct {
	MaxAttempts int           // 最大重试次数（含首次）
	BaseDelay   time.Duration // 初始退避延迟
	MaxDelay    time.Duration // 最大退避延迟
}

// DefaultRetryConfig 默认重试策略
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   1 * time.Second,
		MaxDelay:    10 * time.Second,
	}
}

// RetryProvider 带重试的 Provider 装饰器
type RetryProvider struct {
	inner Provider
	cfg   *RetryConfig
}

// NewRetryProvider 用重试装饰一个 Provider
func NewRetryProvider(inner Provider, cfg *RetryConfig) *RetryProvider {
	if cfg == nil {
		cfg = DefaultRetryConfig()
	}
	return &RetryProvider{inner: inner, cfg: cfg}
}

// Generate 带指数退避重试的 LLM 调用
func (p *RetryProvider) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	var lastErr error

	for attempt := 1; attempt <= p.cfg.MaxAttempts; attempt++ {
		// 检查 context 是否已取消
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context cancelled on attempt %d: %w", attempt, ctx.Err())
		}

		resp, err := p.inner.Generate(ctx, messages)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// 最后一次尝试不等待
		if attempt < p.cfg.MaxAttempts {
			delay := calcBackoff(attempt, p.cfg.BaseDelay, p.cfg.MaxDelay)
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(delay):
			}
		}
	}

	return nil, fmt.Errorf("all %d attempts failed: %w", p.cfg.MaxAttempts, lastErr)
}

// GenerateStream 流式生成（透传，不重试）
func (p *RetryProvider) GenerateStream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return p.inner.GenerateStream(ctx, messages)
}

// WithModel 创建使用指定模型的 RetryProvider 副本（透传给 inner）
func (p *RetryProvider) WithModel(model string) Provider {
	return &RetryProvider{
		inner: p.inner.WithModel(model),
		cfg:   p.cfg,
	}
}

// calcBackoff 计算指数退避延迟
func calcBackoff(attempt int, base, max time.Duration) time.Duration {
	// 2^(n-1) * base, capped at max
	delay := float64(base) * math.Pow(2, float64(attempt-1))
	if delay > float64(max) {
		delay = float64(max)
	}
	return time.Duration(delay)
}
