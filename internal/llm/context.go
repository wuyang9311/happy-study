package llm

import "context"

// contextKey 用于在 context 中传递模型名
type contextKey string

const modelKey contextKey = "user_model"

// WithUserModel 将用户选择的模型名注入 context
func WithUserModel(ctx context.Context, model string) context.Context {
	return context.WithValue(ctx, modelKey, model)
}

// UserModelFromContext 从 context 中读取用户选择的模型名
func UserModelFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(modelKey).(string); ok {
		return v
	}
	return ""
}
