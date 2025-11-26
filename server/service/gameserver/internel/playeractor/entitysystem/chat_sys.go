package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
	"strings"
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

func (cs *ChatSys) OnInit(_ context.Context) {
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

// handleChatWorld 处理世界聊天
func handleChatWorld(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleChatWorld: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	// 解析聊天请求
	var req protocol.C2SChatWorldReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleChatWorld: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	// 获取角色信息
	roleId := playerRole.GetPlayerRoleId()
	roleInfo := playerRole.GetRoleInfo()
	if roleInfo == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "角色信息不存在",
		})
	}

	// 内容验证
	content := req.Content
	if len(content) == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "聊天内容不能为空",
		})
	}

	// 内容长度限制（最大200字符）
	if len(content) > 200 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "聊天内容过长，最多200个字符",
		})
	}

	// 内容过滤
	filteredContent := filterChatContent(content)
	if filteredContent != content {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "聊天内容包含敏感词，请重新输入",
		})
	}

	// 发送到 PublicActor 进行广播
	chatMsg := &protocol.ChatWorldMsg{
		SenderId:   roleId,
		SenderName: roleInfo.RoleName,
		Content:    filteredContent,
	}
	msgData, err := proto.Marshal(chatMsg)
	if err != nil {
		log.Errorf("handleChatWorld: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdChatWorld), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleChatWorld: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleChatPrivate 处理私聊
func handleChatPrivate(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleChatPrivate: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	// 解析聊天请求
	var req protocol.C2SChatPrivateReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleChatPrivate: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	// 获取角色信息
	roleId := playerRole.GetPlayerRoleId()
	roleInfo := playerRole.GetRoleInfo()
	if roleInfo == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "角色信息不存在",
		})
	}

	// 验证目标角色
	if req.TargetId == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "目标角色ID无效",
		})
	}

	if req.TargetId == roleId {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "不能给自己发私聊",
		})
	}

	// 内容验证
	content := req.Content
	if len(content) == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "聊天内容不能为空",
		})
	}

	// 内容长度限制（最大200字符）
	if len(content) > 200 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "聊天内容过长，最多200个字符",
		})
	}

	// 内容过滤
	filteredContent := filterChatContent(content)
	if filteredContent != content {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "聊天内容包含敏感词，请重新输入",
		})
	}

	// 发送到 PublicActor 进行转发
	chatMsg := &protocol.ChatPrivateMsg{
		SenderId:   roleId,
		TargetId:   req.TargetId,
		SenderName: roleInfo.RoleName,
		Content:    filteredContent,
	}
	msgData, err := proto.Marshal(chatMsg)
	if err != nil {
		log.Errorf("handleChatPrivate: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdChatPrivate), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleChatPrivate: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// filterChatContent 过滤聊天内容（使用配置化的敏感词库）
func filterChatContent(content string) string {
	// 从配置管理器获取敏感词配置
	configMgr := jsonconf.GetConfigManager()
	if configMgr == nil {
		return content
	}

	sensitiveWordConfig := configMgr.GetSensitiveWordConfig()
	if sensitiveWordConfig == nil || len(sensitiveWordConfig.Words) == 0 {
		return content
	}

	contentLower := strings.ToLower(content)
	for _, word := range sensitiveWordConfig.Words {
		// 简单的字符串包含检查
		if strings.Contains(contentLower, strings.ToLower(word)) {
			// 包含敏感词，返回空字符串表示需要过滤
			return ""
		}
	}

	return content
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysChat), NewChatSys)
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SChatWorld), handleChatWorld)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SChatPrivate), handleChatPrivate)
	})
}
