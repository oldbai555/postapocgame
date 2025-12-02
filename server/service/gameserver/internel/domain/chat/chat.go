package chat

import (
	"strings"
	"unicode/utf8"
)

// ValidateContent 校验聊天内容
func ValidateContent(content string, maxRunes int) (bool, string) {
	length := utf8.RuneCountInString(content)
	if length == 0 {
		return false, "聊天内容不能为空"
	}
	if length > maxRunes {
		return false, "聊天内容过长"
	}
	return true, ""
}

// ContainsSensitive 判断是否包含敏感词
func ContainsSensitive(content string, words []string) bool {
	if len(words) == 0 {
		return false
	}
	lower := strings.ToLower(content)
	for _, word := range words {
		if word == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(word)) {
			return true
		}
	}
	return false
}
