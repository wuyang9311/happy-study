package llm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
)

// mockProvider 模拟 Provider 用于测试
type mockProvider struct {
	callCount int
	failCount int // 第几次调用失败
	msg       *schema.Message
}

func (m *mockProvider) Generate(ctx context.Context, msgs []*schema.Message) (*schema.Message, error) {
	m.callCount++
	if m.callCount <= m.failCount {
		return nil, errors.New("mock error")
	}
	return m.msg, nil
}

func (m *mockProvider) GenerateStream(ctx context.Context, msgs []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return nil, errors.New("not implemented in mock")
}

func (m *mockProvider) WithModel(model string) Provider {
	return m
}

func TestRetryProvider_SuccessFirstTry(t *testing.T) {
	msg := schema.UserMessage("ok")
	mock := &mockProvider{msg: msg, failCount: 0}
	rp := NewRetryProvider(mock, &RetryConfig{MaxAttempts: 3, BaseDelay: 10 * time.Millisecond})

	resp, err := rp.Generate(context.Background(), []*schema.Message{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "ok" {
		t.Errorf("expected 'ok', got '%s'", resp.Content)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 call, got %d", mock.callCount)
	}
}

func TestRetryProvider_SuccessAfterRetry(t *testing.T) {
	msg := schema.UserMessage("recovered")
	mock := &mockProvider{msg: msg, failCount: 2}
	rp := NewRetryProvider(mock, &RetryConfig{MaxAttempts: 3, BaseDelay: 5 * time.Millisecond, MaxDelay: 20 * time.Millisecond})

	resp, err := rp.Generate(context.Background(), []*schema.Message{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "recovered" {
		t.Errorf("expected 'recovered', got '%s'", resp.Content)
	}
	if mock.callCount != 3 {
		t.Errorf("expected 3 calls, got %d", mock.callCount)
	}
}

func TestRetryProvider_AllFail(t *testing.T) {
	mock := &mockProvider{msg: nil, failCount: 99}
	rp := NewRetryProvider(mock, &RetryConfig{MaxAttempts: 3, BaseDelay: 5 * time.Millisecond})

	_, err := rp.Generate(context.Background(), []*schema.Message{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if mock.callCount != 3 {
		t.Errorf("expected 3 calls, got %d", mock.callCount)
	}
}

func TestRetryProvider_ContextCancelled(t *testing.T) {
	msg := schema.UserMessage("should not complete")
	mock := &mockProvider{msg: msg, failCount: 1}
	rp := NewRetryProvider(mock, &RetryConfig{MaxAttempts: 3, BaseDelay: 100 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	_, err := rp.Generate(ctx, []*schema.Message{})
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}
