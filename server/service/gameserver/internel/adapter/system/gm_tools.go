package system

import (
	"context"
	"fmt"
	manager2 "postapocgame/server/service/gameserver/internel/app/manager"
	"postapocgame/server/service/gameserver/internel/core/gshare"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	mailusecase "postapocgame/server/service/gameserver/internel/usecase/mail"
)

// gmPlayerRepository 仅用于 GM 邮件用例的临时仓储实现
type gmPlayerRepository struct{}

func (r *gmPlayerRepository) GetBinaryData(ctx context.Context, roleID uint64) (*protocol.PlayerRoleBinaryData, error) {
	playerRole := manager2.GetPlayerRole(roleID)
	if playerRole == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found: %d", roleID)
	}
	return playerRole.GetBinaryData(), nil
}

func (r *gmPlayerRepository) SaveBinaryData(ctx context.Context, roleID uint64, data *protocol.PlayerRoleBinaryData) error {
	return database.SavePlayerBinaryData(uint(roleID), data)
}

// SendSystemNotification 发送系统通知给指定玩家
func SendSystemNotification(roleId uint64, title, content string, notificationType, priority uint32) error {
	playerRole := manager2.GetPlayerRole(roleId)
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
	roleMgr := manager2.GetPlayerRoleManager()
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
	playerRole := manager2.GetPlayerRole(roleId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found: %d", roleId)
	}

	ctx := playerRole.WithContext(context.Background())
	repo := &gmPlayerRepository{}
	uc := mailusecase.NewSendCustomMailUseCase(repo)
	if err := uc.Execute(ctx, roleId, title, content, rewards); err != nil {
		log.Errorf("SendSystemMail failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("System mail sent: RoleID=%d, Title=%s", roleId, title)
	return nil
}

// SendSystemMailToAll 发送系统邮件给所有在线玩家
func SendSystemMailToAll(title, content string, rewards []*jsonconf.ItemSt) error {
	roleMgr := manager2.GetPlayerRoleManager()
	allRoles := roleMgr.GetAll()

	successCount := 0
	for _, playerRole := range allRoles {
		if playerRole == nil {
			continue
		}

		ctx := playerRole.WithContext(context.Background())
		repo := &gmPlayerRepository{}
		uc := mailusecase.NewSendCustomMailUseCase(repo)
		if err := uc.Execute(ctx, playerRole.GetPlayerRoleId(), title, content, rewards); err != nil {
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
	playerRole := manager2.GetPlayerRole(roleId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found: %d", roleId)
	}

	ctx := playerRole.WithContext(context.Background())
	repo := &gmPlayerRepository{}
	uc := mailusecase.NewSendTemplateMailUseCase(repo, nil)
	if err := uc.Execute(ctx, roleId, templateId, args); err != nil {
		log.Errorf("SendSystemMailByTemplate failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("System mail sent by template: RoleID=%d, TemplateID=%d", roleId, templateId)
	return nil
}

// SendSystemMailByTemplateToAll 使用模板发送系统邮件给所有在线玩家
func SendSystemMailByTemplateToAll(templateId uint32, args map[string]interface{}) error {
	roleMgr := manager2.GetPlayerRoleManager()
	allRoles := roleMgr.GetAll()

	successCount := 0
	for _, playerRole := range allRoles {
		if playerRole == nil {
			continue
		}

		ctx := playerRole.WithContext(context.Background())
		repo := &gmPlayerRepository{}
		uc := mailusecase.NewSendTemplateMailUseCase(repo, nil)
		if err := uc.Execute(ctx, playerRole.GetPlayerRoleId(), templateId, args); err != nil {
			log.Errorf("SendSystemMailByTemplateToAll failed for role %d: %v", playerRole.GetPlayerRoleId(), err)
			continue
		}
		successCount++
	}

	log.Infof("System mail sent by template to all: SuccessCount=%d/%d, TemplateID=%d", successCount, len(allRoles), templateId)
	return nil
}

// handleGMCommand 处理GM命令（供控制层复用的低层解析逻辑）
func HandleGMCommand(ctx context.Context, msg *network.ClientMessage) (string, bool, error) {
	sessionId, ok := ctx.Value(gshare.ContextKeySession).(string)
	if !ok {
		return "", false, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found in context")
	}

	// 解析GM命令请求
	var req protocol.C2SGMCommandReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		log.Errorf("unmarshal GM command request failed: %v", err)
		return sessionId, false, err
	}

	// 获取玩家角色
	roleMgr := manager2.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		return sessionId, false, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	roleCtx := playerRole.WithContext(ctx)

	// 检查系统是否开启
	gmSys := GetGMSys(roleCtx)
	if gmSys == nil {
		return sessionId, false, customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "GM系统未开启")
	}
	success, message := gmSys.ExecuteCommand(roleCtx, req.GmName, req.Args)
	if success {
		// 成功时仅返回 success 标志与空 error，由调用方自行填充成功文案（若需要）
		return sessionId, true, nil
	}
	// 失败时将业务文案包装为 error，便于上层直接回传给客户端
	return sessionId, false, fmt.Errorf("%s", message)
}
