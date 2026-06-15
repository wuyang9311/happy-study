package auth

import "strings"

// 敏感词列表（基础版）
var sensitiveWords = []string{
	"fuck", "shit", "asshole", "bitch", "damn", "crap",
	"赌博", "色情", "毒品", "枪支", "暴力", "诈骗",
}

// 保留用户名列表（系统占用）
var reservedUsernames = []string{
	"admin", "root", "system", "administrator", "superuser",
	"管理员", "系统", "客服", "官方",
}

// IsReservedUsername 检查是否为保留用户名
func IsReservedUsername(username string) bool {
	lower := strings.ToLower(username)
	for _, reserved := range reservedUsernames {
		if lower == strings.ToLower(reserved) {
			return true
		}
	}
	return false
}

// ContainsSensitiveWord 检查是否包含敏感词
func ContainsSensitiveWord(text string) bool {
	lower := strings.ToLower(text)
	for _, word := range sensitiveWords {
		if strings.Contains(lower, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// SanitizeUsername 清洗用户名（去除特殊字符）
func SanitizeUsername(username string) string {
	// 只允许字母、数字、下划线、中划线、中文
	var result strings.Builder
	for _, r := range username {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-' ||
			(r >= 0x4E00 && r <= 0x9FFF) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ValidatePassword 密码强度检查
func ValidatePassword(password string) (bool, string) {
	if len(password) < 6 {
		return false, "密码长度不能少于6位"
	}
	if len(password) > 64 {
		return false, "密码长度不能超过64位"
	}

	hasLetter := false
	hasDigit := false

	for _, r := range password {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
		}
		if r >= '0' && r <= '9' {
			hasDigit = true
		}
	}

	if !hasLetter {
		return false, "密码必须包含至少一个字母"
	}
	if !hasDigit {
		return false, "密码必须包含至少一个数字"
	}

	return true, ""
}

// ValidateUsername 用户名合法性检查
func ValidateUsername(username string) (bool, string) {
	if len(username) < 3 {
		return false, "用户名长度不能少于3位"
	}
	if len(username) > 32 {
		return false, "用户名长度不能超过32位"
	}

	// 检查是否只包含合法字符
	cleaned := SanitizeUsername(username)
	if cleaned != username {
		return false, "用户名只能包含字母、数字、下划线、中划线和中文"
	}

	// 检查是否以字母或中文开头
	first := []rune(username)[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') ||
		(first >= 0x4E00 && first <= 0x9FFF)) {
		return false, "用户名必须以字母或中文开头"
	}

	return true, ""
}
