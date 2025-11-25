package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
)

// GMSys 玩家级GM系统
type GMSys struct {
	*BaseSystem
	mgr *GMManager
}

func NewGMSys() *GMSys {
	return &GMSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysGM)),
		mgr:        NewGMManager(),
	}
}

func GetGMSys(ctx context.Context) *GMSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysGM))
	if system == nil {
		return nil
	}
	gmSys, ok := system.(*GMSys)
	if !ok || !gmSys.IsOpened() {
		return nil
	}
	return gmSys
}

func (gm *GMSys) OnInit(context.Context) {
	gm.mgr = NewGMManager()
}

func (gm *GMSys) ExecuteCommand(ctx context.Context, gmName string, args []string) (bool, string) {
	if gm.mgr == nil {
		return false, "GM系统未初始化"
	}
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return false, err.Error()
	}
	return gm.mgr.ExecuteCommand(ctx, playerRole, gmName, args)
}

// SendSystemNotification 发送系统通知给指定玩家
func SendSystemNotification(roleId uint64, title, content string, notificationType, priority uint32) error {
	playerRole := manager.GetPlayerRole(roleId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found: %d", roleId)
	}

	notification := &protocol.S2CSystemNotificationReq{
		Title:    title,
		Content:  content,
		Type:     notificationType,
		Priority: priority,
	}

	if err := playerRole.SendProtoMessage(uint16(protocol.S2CProtocol_S2CSystemNotification), notification); err != nil {
		log.Errorf("SendSystemNotification failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("System notification sent: RoleID=%d, Title=%s", roleId, title)
	return nil
}

// SendSystemNotificationToAll 发送系统通知给所有在线玩家
func SendSystemNotificationToAll(title, content string, notificationType, priority uint32) error {
	roleMgr := manager.GetPlayerRoleManager()
	allRoles := roleMgr.GetAll()

	successCount := 0
	for _, playerRole := range allRoles {
		if playerRole == nil {
			continue
		}

		notification := &protocol.S2CSystemNotificationReq{
			Title:    title,
			Content:  content,
			Type:     notificationType,
			Priority: priority,
		}

		if err := playerRole.SendProtoMessage(uint16(protocol.S2CProtocol_S2CSystemNotification), notification); err != nil {
			log.Errorf("SendSystemNotificationToAll failed for role %d: %v", playerRole.GetPlayerRoleId(), err)
			continue
		}
		successCount++
	}

	log.Infof("System notification sent to all: SuccessCount=%d/%d, Title=%s", successCount, len(allRoles), title)
	return nil
}

// SendSystemMail 发送系统邮件给指定玩家
func SendSystemMail(roleId uint64, title, content string, rewards []*jsonconf.ItemSt) error {
	playerRole := manager.GetPlayerRole(roleId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found: %d", roleId)
	}

	ctx := playerRole.WithContext(context.Background())
	mailSys := GetMailSys(ctx)
	if mailSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "mail system not ready")
	}

	if err := mailSys.SendCustomMail(ctx, title, content, rewards); err != nil {
		log.Errorf("SendSystemMail failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("System mail sent: RoleID=%d, Title=%s", roleId, title)
	return nil
}

// SendSystemMailToAll 发送系统邮件给所有在线玩家
func SendSystemMailToAll(title, content string, rewards []*jsonconf.ItemSt) error {
	roleMgr := manager.GetPlayerRoleManager()
	allRoles := roleMgr.GetAll()

	successCount := 0
	for _, playerRole := range allRoles {
		if playerRole == nil {
			continue
		}

		ctx := playerRole.WithContext(context.Background())
		mailSys := GetMailSys(ctx)
		if mailSys == nil {
			log.Errorf("SendSystemMailToAll: mail system not ready for role %d", playerRole.GetPlayerRoleId())
			continue
		}

		if err := mailSys.SendCustomMail(ctx, title, content, rewards); err != nil {
			log.Errorf("SendSystemMailToAll failed for role %d: %v", playerRole.GetPlayerRoleId(), err)
			continue
		}
		successCount++
	}

	log.Infof("System mail sent to all: SuccessCount=%d/%d, Title=%s", successCount, len(allRoles), title)
	return nil
}

// SendSystemMailByTemplate 使用模板发送系统邮件给指定玩家
func SendSystemMailByTemplate(roleId uint64, templateId uint32, args map[string]interface{}) error {
	playerRole := manager.GetPlayerRole(roleId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found: %d", roleId)
	}

	ctx := playerRole.WithContext(context.Background())
	mailSys := GetMailSys(ctx)
	if mailSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "mail system not ready")
	}

	if err := mailSys.SendMail(ctx, templateId, args); err != nil {
		log.Errorf("SendSystemMailByTemplate failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("System mail sent by template: RoleID=%d, TemplateID=%d", roleId, templateId)
	return nil
}

// SendSystemMailByTemplateToAll 使用模板发送系统邮件给所有在线玩家
func SendSystemMailByTemplateToAll(templateId uint32, args map[string]interface{}) error {
	roleMgr := manager.GetPlayerRoleManager()
	allRoles := roleMgr.GetAll()

	successCount := 0
	for _, playerRole := range allRoles {
		if playerRole == nil {
			continue
		}

		ctx := playerRole.WithContext(context.Background())
		mailSys := GetMailSys(ctx)
		if mailSys == nil {
			log.Errorf("SendSystemMailByTemplateToAll: mail system not ready for role %d", playerRole.GetPlayerRoleId())
			continue
		}

		if err := mailSys.SendMail(ctx, templateId, args); err != nil {
			log.Errorf("SendSystemMailByTemplateToAll failed for role %d: %v", playerRole.GetPlayerRoleId(), err)
			continue
		}
		successCount++
	}

	log.Infof("System mail sent by template to all: SuccessCount=%d/%d, TemplateID=%d", successCount, len(allRoles), templateId)
	return nil
}

// GrantRewardsByMail 通过邮件发放奖励
func GrantRewardsByMail(roleId uint64, title, content string, rewards []*jsonconf.ItemAmount) error {
	playerRole := manager.GetPlayerRole(roleId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found: %d", roleId)
	}

	ctx := playerRole.WithContext(context.Background())
	mailSys := GetMailSys(ctx)
	if mailSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "mail system not ready")
	}

	mailRewards := make([]*jsonconf.ItemSt, 0, len(rewards))
	for _, reward := range rewards {
		if reward == nil || reward.Count <= 0 {
			continue
		}
		mailRewards = append(mailRewards, &jsonconf.ItemSt{
			ItemId: reward.ItemId,
			Count:  uint32(reward.Count),
			Type:   reward.ItemType,
		})
	}

	if err := mailSys.SendCustomMail(ctx, title, content, mailRewards); err != nil {
		log.Errorf("GrantRewardsByMail failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("Rewards granted by mail: RoleID=%d, Title=%s", roleId, title)
	return nil
}

// handleGMCommand 处理GM命令
func handleGMCommand(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)

	// 解析GM命令请求
	var req protocol.C2SGMCommandReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		log.Errorf("unmarshal GM command request failed: %v", err)
		return err
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	roleCtx := playerRole.WithContext(ctx)

	// 执行GM命令
	gmSys := GetGMSys(roleCtx)
	if gmSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "gm system not ready")
	}
	success, message := gmSys.ExecuteCommand(roleCtx, req.GmName, req.Args)

	// 发送GM命令结果
	resp := &protocol.S2CGMCommandResultReq{
		Success: success,
		Message: message,
	}

	if err := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CGMCommandResult), resp); err != nil {
		log.Errorf("send GM command result failed: %v", err)
		return err
	}

	log.Infof("GM command executed: RoleID=%d, GMName=%s, Success=%v, Message=%s",
		playerRole.GetPlayerRoleId(), req.GmName, success, message)

	return nil
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysGM), func() iface.ISystem {
		return NewGMSys()
	})
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SGMCommand), handleGMCommand)
	})
}
