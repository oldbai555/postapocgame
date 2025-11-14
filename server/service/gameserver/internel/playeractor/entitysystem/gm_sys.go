package entitysystem

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/manager"
	"sync"
)

// GMSys GM系统（用于手动模拟系统通知和邮件奖励）
type GMSys struct{}

// NewGMSys 创建GM系统（单例，不需要注册为系统）
func NewGMSys() *GMSys {
	return &GMSys{}
}

var (
	globalGMSys *GMSys
	gmSysOnce   sync.Once
)

// GetGMSys 获取全局GM系统
func GetGMSys() *GMSys {
	gmSysOnce.Do(func() {
		globalGMSys = NewGMSys()
	})
	return globalGMSys
}

// SendSystemNotification 发送系统通知给指定玩家
func (gm *GMSys) SendSystemNotification(roleId uint64, title, content string, notificationType, priority uint32) error {
	playerRole := manager.GetPlayerRole(roleId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found: %d", roleId)
	}

	// 发送系统通知
	notification := &protocol.S2CSystemNotificationReq{
		Title:    title,
		Content:  content,
		Type:     notificationType,
		Priority: priority,
	}

	if err := playerRole.SendJsonMessage(uint16(protocol.S2CProtocol_S2CSystemNotification), notification); err != nil {
		log.Errorf("SendSystemNotification failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("System notification sent: RoleID=%d, Title=%s", roleId, title)
	return nil
}

// SendSystemNotificationToAll 发送系统通知给所有在线玩家
func (gm *GMSys) SendSystemNotificationToAll(title, content string, notificationType, priority uint32) error {
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

		if err := playerRole.SendJsonMessage(uint16(protocol.S2CProtocol_S2CSystemNotification), notification); err != nil {
			log.Errorf("SendSystemNotificationToAll failed for role %d: %v", playerRole.GetPlayerRoleId(), err)
			continue
		}
		successCount++
	}

	log.Infof("System notification sent to all: SuccessCount=%d/%d, Title=%s", successCount, len(allRoles), title)
	return nil
}

// SendSystemMail 发送系统邮件给指定玩家
func (gm *GMSys) SendSystemMail(roleId uint64, title, content string, rewards []*jsonconf.ItemSt) error {
	playerRole := manager.GetPlayerRole(roleId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found: %d", roleId)
	}

	ctx := playerRole.WithContext(context.Background())
	mailSys := GetMailSys(ctx)
	if mailSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "mail system not ready")
	}

	// 发送自定义邮件
	if err := mailSys.SendCustomMail(ctx, title, content, rewards); err != nil {
		log.Errorf("SendSystemMail failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("System mail sent: RoleID=%d, Title=%s", roleId, title)
	return nil
}

// SendSystemMailToAll 发送系统邮件给所有在线玩家
func (gm *GMSys) SendSystemMailToAll(title, content string, rewards []*jsonconf.ItemSt) error {
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

		// 发送自定义邮件
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
func (gm *GMSys) SendSystemMailByTemplate(roleId uint64, templateId uint32, args map[string]interface{}) error {
	playerRole := manager.GetPlayerRole(roleId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found: %d", roleId)
	}

	ctx := playerRole.WithContext(context.Background())
	mailSys := GetMailSys(ctx)
	if mailSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "mail system not ready")
	}

	// 发送模板邮件
	if err := mailSys.SendMail(ctx, templateId, args); err != nil {
		log.Errorf("SendSystemMailByTemplate failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("System mail sent by template: RoleID=%d, TemplateID=%d", roleId, templateId)
	return nil
}

// SendSystemMailByTemplateToAll 使用模板发送系统邮件给所有在线玩家
func (gm *GMSys) SendSystemMailByTemplateToAll(templateId uint32, args map[string]interface{}) error {
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

		// 发送模板邮件
		if err := mailSys.SendMail(ctx, templateId, args); err != nil {
			log.Errorf("SendSystemMailByTemplateToAll failed for role %d: %v", playerRole.GetPlayerRoleId(), err)
			continue
		}
		successCount++
	}

	log.Infof("System mail sent by template to all: SuccessCount=%d/%d, TemplateID=%d", successCount, len(allRoles), templateId)
	return nil
}

// GrantRewardsByMail 通过邮件发放奖励（用于GM补发奖励等场景）
func (gm *GMSys) GrantRewardsByMail(roleId uint64, title, content string, rewards []*jsonconf.ItemAmount) error {
	playerRole := manager.GetPlayerRole(roleId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found: %d", roleId)
	}

	ctx := playerRole.WithContext(context.Background())
	mailSys := GetMailSys(ctx)
	if mailSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "mail system not ready")
	}

	// 转换奖励格式
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

	// 发送邮件
	if err := mailSys.SendCustomMail(ctx, title, content, mailRewards); err != nil {
		log.Errorf("GrantRewardsByMail failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("Rewards granted by mail: RoleID=%d, Title=%s", roleId, title)
	return nil
}
