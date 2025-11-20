package entitysystem

import (
	"context"
	"encoding/json"
	"fmt"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
	"strconv"
	"strings"
)

// MailSys 邮件系统
type MailSys struct {
	*BaseSystem
	mailData  *protocol.SiMailData
	mailIdGen uint64 // 邮件ID生成器
}

// NewMailSys 创建邮件系统
func NewMailSys() *MailSys {
	return &MailSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysMail)),
		mailIdGen:  1,
	}
}

func GetMailSys(ctx context.Context) *MailSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysMail))
	if system == nil {
		return nil
	}
	mailSys, ok := system.(*MailSys)
	if !ok || !mailSys.IsOpened() {
		return nil
	}
	return mailSys
}

// OnInit 系统初始化
func (ms *MailSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("mail sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果mail_data不存在，则初始化
	if binaryData.MailData == nil {
		binaryData.MailData = &protocol.SiMailData{
			Mails: make([]*protocol.MailSt, 0),
		}
	}
	ms.mailData = binaryData.MailData

	// 如果Mails为空，初始化为空切片
	if ms.mailData.Mails == nil {
		ms.mailData.Mails = make([]*protocol.MailSt, 0)
	}

	// 初始化邮件ID生成器（找到最大ID+1）
	for _, mail := range ms.mailData.Mails {
		if mail != nil && mail.MailId >= ms.mailIdGen {
			ms.mailIdGen = mail.MailId + 1
		}
	}

	// 清理过期邮件
	ms.cleanExpiredMails()

	log.Infof("MailSys initialized: MailCount=%d", len(ms.mailData.Mails))
}

// GetMailData 获取邮件数据
func (ms *MailSys) GetMailData() *protocol.SiMailData {
	return ms.mailData
}

// GetMail 获取指定邮件
func (ms *MailSys) GetMail(mailId uint64) *protocol.MailSt {
	if ms.mailData == nil || ms.mailData.Mails == nil {
		return nil
	}
	for _, mail := range ms.mailData.Mails {
		if mail != nil && mail.MailId == mailId {
			return mail
		}
	}
	return nil
}

// SendMail 发送邮件（使用模板）
func (ms *MailSys) SendMail(ctx context.Context, templateId uint32, args map[string]interface{}) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	// 获取邮件模板配置
	templateConfig, ok := jsonconf.GetConfigManager().GetMailTemplateConfig(templateId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "mail template config not found: %d", templateId)
	}

	// 生成邮件ID
	mailId := ms.mailIdGen
	ms.mailIdGen++

	// 序列化参数
	argsJson := "{}"
	if args != nil {
		argsBytes, err := json.Marshal(args)
		if err == nil {
			argsJson = string(argsBytes)
		}
	}

	// 创建邮件
	mail := &protocol.MailSt{
		MailId:   mailId,
		ConfId:   templateId,
		Status:   0, // 未读
		CreateAt: uint32(servertime.Now().Unix()),
		Args:     argsJson,
		Title:    templateConfig.Title,
		Content:  templateConfig.Content,
	}

	// 处理附件（从模板配置中获取）
	if len(templateConfig.Rewards) > 0 {
		// 将奖励转换为附件字符串格式: id_count_bind|id_count_bind|...
		attachments := make([]string, 0, len(templateConfig.Rewards))
		for _, reward := range templateConfig.Rewards {
			attachments = append(attachments, fmt.Sprintf("%d_%d_%d", reward.ItemId, reward.Count, 1))
		}
		mail.Files = strings.Join(attachments, "|")
	}

	// 添加到邮件列表
	if ms.mailData.Mails == nil {
		ms.mailData.Mails = make([]*protocol.MailSt, 0)
	}
	ms.mailData.Mails = append(ms.mailData.Mails, mail)

	log.Infof("Mail sent: RoleID=%d, MailID=%d, TemplateID=%d", playerRole.GetPlayerRoleId(), mailId, templateId)
	return nil
}

// SendCustomMail 发送自定义邮件（不使用模板）
func (ms *MailSys) SendCustomMail(ctx context.Context, title, content string, rewards []*jsonconf.ItemSt) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	// 生成邮件ID
	mailId := ms.mailIdGen
	ms.mailIdGen++

	// 创建邮件
	mail := &protocol.MailSt{
		MailId:   mailId,
		ConfId:   0, // 自定义邮件使用0
		Status:   0, // 未读
		CreateAt: uint32(servertime.Now().Unix()),
		Args:     "{}",
		Title:    title,
		Content:  content,
	}

	// 处理附件
	if len(rewards) > 0 {
		attachments := make([]string, 0, len(rewards))
		for _, reward := range rewards {
			attachments = append(attachments, fmt.Sprintf("%d_%d_%d", reward.ItemId, reward.Count, 1))
		}
		mail.Files = strings.Join(attachments, "|")
	}

	// 添加到邮件列表
	if ms.mailData.Mails == nil {
		ms.mailData.Mails = make([]*protocol.MailSt, 0)
	}
	ms.mailData.Mails = append(ms.mailData.Mails, mail)

	log.Infof("Custom mail sent: RoleID=%d, MailID=%d", playerRole.GetPlayerRoleId(), mailId)
	return nil
}

// ReadMail 读取邮件（标记为已读）
func (ms *MailSys) ReadMail(ctx context.Context, mailId uint64) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	mail := ms.GetMail(mailId)
	if mail == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "mail not found: %d", mailId)
	}

	// 标记为已读
	if mail.Status == 0 {
		mail.Status = 1
		log.Infof("Mail read: RoleID=%d, MailID=%d", playerRole.GetPlayerRoleId(), mailId)
	}

	return nil
}

// ClaimMailAttachment 领取邮件附件
func (ms *MailSys) ClaimMailAttachment(ctx context.Context, mailId uint64) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	mail := ms.GetMail(mailId)
	if mail == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "mail not found: %d", mailId)
	}

	// 检查是否已领取
	if mail.Status == 2 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "mail attachment already claimed: %d", mailId)
	}

	// 解析附件
	if mail.Files == "" {
		// 没有附件，直接标记为已读已领取
		mail.Status = 2
		return nil
	}

	// 解析附件字符串: id_count_bind|id_count_bind|...
	attachments := strings.Split(mail.Files, "|")
	rewards := make([]*jsonconf.ItemAmount, 0, len(attachments))

	for _, attachment := range attachments {
		if attachment == "" {
			continue
		}
		parts := strings.Split(attachment, "_")
		if len(parts) != 3 {
			log.Warnf("Invalid attachment format: %s", attachment)
			continue
		}

		itemId, err1 := strconv.ParseUint(parts[0], 10, 32)
		count, err2 := strconv.ParseUint(parts[1], 10, 32)
		bind, err3 := strconv.ParseUint(parts[2], 10, 32)

		if err1 != nil || err2 != nil || err3 != nil {
			log.Warnf("Failed to parse attachment: %s", attachment)
			continue
		}

		rewards = append(rewards, &jsonconf.ItemAmount{
			ItemType: uint32(protocol.ItemType_ItemTypeMaterial), // 默认材料类型
			ItemId:   uint32(itemId),
			Count:    int64(count),
			Bind:     uint32(bind),
		})
	}

	// 发放奖励
	if len(rewards) > 0 {
		if err := playerRole.GrantRewards(ctx, rewards); err != nil {
			log.Errorf("GrantRewards failed: %v", err)
			return customerr.Wrap(err)
		}
	}

	// 标记为已读已领取
	mail.Status = 2

	// 如果邮件有Items字段，也更新它（用于客户端显示）
	if len(rewards) > 0 {
		// 只取第一个奖励作为Items（协议定义中Items是单个ItemSt）
		mail.Items = &protocol.ItemSt{
			ItemId: rewards[0].ItemId,
			Count:  uint32(rewards[0].Count),
			Bind:   rewards[0].Bind,
		}
	}

	log.Infof("Mail attachment claimed: RoleID=%d, MailID=%d", playerRole.GetPlayerRoleId(), mailId)
	return nil
}

// DeleteMail 删除邮件
func (ms *MailSys) DeleteMail(ctx context.Context, mailId uint64) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	if ms.mailData == nil || ms.mailData.Mails == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "mail not found: %d", mailId)
	}

	// 查找并删除邮件
	for i, mail := range ms.mailData.Mails {
		if mail != nil && mail.MailId == mailId {
			// 从切片中移除
			ms.mailData.Mails = append(ms.mailData.Mails[:i], ms.mailData.Mails[i+1:]...)
			log.Infof("Mail deleted: RoleID=%d, MailID=%d", playerRole.GetPlayerRoleId(), mailId)
			return nil
		}
	}

	return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "mail not found: %d", mailId)
}

// DeleteMails 批量删除邮件
func (ms *MailSys) DeleteMails(ctx context.Context, mailIds []uint64) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	if ms.mailData == nil || ms.mailData.Mails == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "no mails")
	}

	// 构建要删除的邮件ID集合
	deleteSet := make(map[uint64]bool)
	for _, mailId := range mailIds {
		deleteSet[mailId] = true
	}

	// 过滤出要保留的邮件
	validMails := make([]*protocol.MailSt, 0, len(ms.mailData.Mails))
	for _, mail := range ms.mailData.Mails {
		if mail != nil && !deleteSet[mail.MailId] {
			validMails = append(validMails, mail)
		}
	}

	ms.mailData.Mails = validMails
	log.Infof("Mails deleted: RoleID=%d, Count=%d", playerRole.GetPlayerRoleId(), len(mailIds))
	return nil
}

// ClaimMailAttachments 批量领取邮件附件
func (ms *MailSys) ClaimMailAttachments(ctx context.Context, mailIds []uint64) error {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	successCount := 0
	for _, mailId := range mailIds {
		if err := ms.ClaimMailAttachment(ctx, mailId); err != nil {
			log.Warnf("Claim mail attachment failed: MailID=%d, Error=%v", mailId, err)
		} else {
			successCount++
		}
	}

	log.Infof("Mail attachments claimed: RoleID=%d, SuccessCount=%d, TotalCount=%d", playerRole.GetPlayerRoleId(), successCount, len(mailIds))
	return nil
}

// GetMailsByCategory 根据分类获取邮件（系统邮件、玩家邮件、活动邮件）
func (ms *MailSys) GetMailsByCategory(category uint32) []*protocol.MailSt {
	if ms.mailData == nil || ms.mailData.Mails == nil {
		return nil
	}

	result := make([]*protocol.MailSt, 0)
	for _, mail := range ms.mailData.Mails {
		if mail == nil {
			continue
		}

		// 根据ConfId判断分类：0=自定义邮件(玩家邮件)，>0=模板邮件(系统/活动邮件)
		mailCategory := uint32(0)
		if mail.ConfId > 0 {
			// 系统邮件或活动邮件
			mailCategory = 1
		}

		if mailCategory == category {
			result = append(result, mail)
		}
	}

	return result
}

// SearchMails 搜索邮件（根据标题或内容）
func (ms *MailSys) SearchMails(keyword string) []*protocol.MailSt {
	if ms.mailData == nil || ms.mailData.Mails == nil {
		return nil
	}

	if keyword == "" {
		return ms.mailData.Mails
	}

	result := make([]*protocol.MailSt, 0)
	keywordLower := strings.ToLower(keyword)

	for _, mail := range ms.mailData.Mails {
		if mail == nil {
			continue
		}

		// 搜索标题和内容
		if strings.Contains(strings.ToLower(mail.Title), keywordLower) ||
			strings.Contains(strings.ToLower(mail.Content), keywordLower) {
			result = append(result, mail)
		}
	}

	return result
}

// cleanExpiredMails 清理过期邮件
func (ms *MailSys) cleanExpiredMails() {
	if ms.mailData == nil || ms.mailData.Mails == nil {
		return
	}

	now := servertime.Now().Unix()
	validMails := make([]*protocol.MailSt, 0, len(ms.mailData.Mails))

	for _, mail := range ms.mailData.Mails {
		if mail == nil {
			continue
		}

		// 获取邮件模板配置
		if mail.ConfId > 0 {
			templateConfig, ok := jsonconf.GetConfigManager().GetMailTemplateConfig(mail.ConfId)
			if ok && templateConfig.ExpireHours > 0 {
				expireTime := int64(mail.CreateAt) + int64(templateConfig.ExpireHours)*3600
				if now > expireTime {
					// 邮件已过期，跳过
					continue
				}
			}
		}

		validMails = append(validMails, mail)
	}

	ms.mailData.Mails = validMails
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysMail), func() iface.ISystem {
		return NewMailSys()
	})
}
