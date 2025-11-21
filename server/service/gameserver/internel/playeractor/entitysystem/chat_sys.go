package entitysystem

import (
	"context"
	"time"
	"unicode/utf8"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/service/gameserver/internel/iface"
)

var (
	chatCooldown        = 10 * time.Second            // 每10秒允许发一次
	chatMaxRuneLength   = 26                          // 最多26个中文字符
	chatSensitiveFilter = []string{"测试", "禁用", "不和谐"} // 敏感词列表
)

// ChatSys 聊天系统
// 挂载到玩家实体
// 支持世界、私聊两种消息入口、频率限制、敏感词过滤
// 冷却与最近发送时间不区分频道
type ChatSys struct {
	*BaseSystem
	lastChatTime time.Time
}

func NewChatSys() iface.ISystem {
	return &ChatSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysChat)),
	}
}

func (cs *ChatSys) OnInit(ctx context.Context) {
	// do nothing
}

// CheckChatMessage 校验消息内容（返回错误字符串为原因）
func (cs *ChatSys) CheckChatMessage(content string) (ok bool, reason string) {
	length := utf8.RuneCountInString(content)
	if length == 0 {
		return false, "消息不能为空"
	}
	if length > chatMaxRuneLength {
		return false, "消息过长（最大26汉字）"
	}
	for _, word := range chatSensitiveFilter {
		if len(word) > 0 && contains(content, word) {
			return false, "消息包含敏感词: " + word
		}
	}
	return true, ""
}

// CheckCooldown 检查聊天冷却
func (cs *ChatSys) CheckCooldown() bool {
	if servertime.Since(cs.lastChatTime) < chatCooldown {
		return false
	}
	return true
}

// SetCooldown 更新聊天冷却时间戳
func (cs *ChatSys) SetCooldown() {
	cs.lastChatTime = servertime.Now()
}

// contains 判断字符串包含（支持中文敏感词简单遍历）
func contains(content, keyword string) bool {
	if len(keyword) == 0 || len(content) == 0 {
		return false
	}
	if utf8.RuneCountInString(keyword) == 0 {
		return false
	}
	contentRunes := []rune(content)
	keywordRunes := []rune(keyword)
	if len(contentRunes) < len(keywordRunes) {
		return false
	}
	return findSubstr(contentRunes, keywordRunes)
}

func findSubstr(haystack, needle []rune) bool {
outer:
	for i := 0; i <= len(haystack)-len(needle); i++ {
		for j, v := range needle {
			if haystack[i+j] != v {
				continue outer
			}
		}
		return true
	}
	return false
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysChat), NewChatSys)
}
