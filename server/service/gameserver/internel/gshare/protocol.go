/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package gshare

// ContextKey 类型用于定义 Context 的 key，避免使用字符串导致的冲突
type ContextKey string

const (
	// ContextKeyRole 用于在 Context 中存储玩家角色对象
	ContextKeyRole ContextKey = "playerRole"
	// ContextKeySession 用于在 Context 中存储 Session ID
	ContextKeySession ContextKey = "playerRoleSession"
)
