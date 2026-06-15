package common

import (
	"strings"
)

// SanitizeInput 清洗用户输入，防止 prompt 注入
// 移除控制字符，限制长度
func SanitizeInput(input string) string {
	// 移除控制字符（保留换行和制表符）
	safe := strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, input)

	// 限制最大长度（避免 token 爆炸）
	const maxLen = 2000
	if len(safe) > maxLen {
		safe = safe[:maxLen]
	}

	return safe
}
