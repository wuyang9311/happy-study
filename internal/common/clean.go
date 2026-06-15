package common

import (
	"strings"
)

// CleanJSON 清理 LLM 输出中的 JSON 标记
// 处理 ```json ... ``` 包裹及前后空白
func CleanJSON(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	return content
}
