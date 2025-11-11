package entitysystem

import (
	"fmt"
	"postapocgame/server/internal/custom_id"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
	"sync/atomic"
	"time"
)

var (
	ErrMailNotFound              = fmt.Errorf("mail not found")
	ErrMailNoReward              = fmt.Errorf("mail has no reward")
	ErrMailRewardAlreadyReceived = fmt.Errorf("mail reward already received")
	ErrMailHasUnreceivedReward   = fmt.Errorf("mail has unreceived reward")
	ErrMailRewardTooMany         = fmt.Errorf("mail reward too many (max 10)")
)

var nextMailID uint64 = 1

// MailSys 邮件系统
type MailSys struct {
	*BaseSystem
	mails map[uint64]*protocol.Mail // mailID -> Mail
}

// NewMailSys 创建邮件系统
func NewMailSys(role iface.IPlayerRole) *MailSys {
	sys := &MailSys{
		BaseSystem: NewBaseSystem(custom_id.SysMail, role),
		mails:      make(map[uint64]*protocol.Mail),
	}
	return sys
}

// OnRoleLogin 角色登录时下发邮件列表
func (s *MailSys) OnRoleLogin() {
	return
}

// SendMailList 下发邮件列表
func (s *MailSys) SendMailList() error {
	mails := make([]protocol.Mail, 0, len(s.mails))
	for _, mail := range s.mails {
		mails = append(mails, *mail)
	}

	data := &protocol.MailListResponse{
		Mails: mails,
	}
	jsonData, _ := tool.JsonMarshal(data)
	return s.role.SendMessage(protocol.S2C_MailList, jsonData)
}

// SendMailDetail 下发邮件详情
func (s *MailSys) SendMailDetail(mailID uint64) error {
	mail, ok := s.mails[mailID]
	if !ok {
		return ErrMailNotFound
	}

	data := &protocol.MailDetailResponse{
		Mail: *mail,
	}
	jsonData, _ := tool.JsonMarshal(data)
	return s.role.SendMessage(protocol.S2C_MailDetail, jsonData)
}

// AddMail 添加邮件
func (s *MailSys) AddMail(title, content, sender string, rewards []protocol.Item) uint64 {
	mailId := atomic.AddUint64(&nextMailID, 1)

	mail := &protocol.Mail{
		MailId:     mailId,
		Title:      title,
		Content:    content,
		Sender:     sender,
		SendTime:   time.Now().Unix(),
		HasReward:  len(rewards) > 0,
		IsRead:     false,
		IsReceived: false,
		Rewards:    rewards,
	}

	s.mails[mailId] = mail
	s.SendMailList()

	return mailId
}

// ReadMail 读取邮件
func (s *MailSys) ReadMail(mailID uint64) error {
	mail, ok := s.mails[mailID]
	if !ok {
		return ErrMailNotFound
	}

	if !mail.IsRead {
		mail.IsRead = true
		s.SendMailDetail(mailID)
	}

	return nil
}

// DeleteMail 删除邮件
func (s *MailSys) DeleteMail(mailID uint64) error {
	mail, ok := s.mails[mailID]
	if !ok {
		return ErrMailNotFound
	}

	// 如果有未领取的奖励，不允许删除
	if mail.HasReward && !mail.IsReceived {
		return ErrMailHasUnreceivedReward
	}

	delete(s.mails, mailID)
	s.SendMailList()

	return nil
}

// ReceiveMailReward 领取邮件奖励
func (s *MailSys) ReceiveMailReward(mailID uint64) error {
	mail, ok := s.mails[mailID]
	if !ok {
		return ErrMailNotFound
	}

	if !mail.HasReward {
		return ErrMailNoReward
	}

	if mail.IsReceived {
		return ErrMailRewardAlreadyReceived
	}

	// 检查奖励数量（最多10个）
	if len(mail.Rewards) > 10 {
		return ErrMailRewardTooMany
	}

	// 发放奖励
	if err := s.role.GiveAwards(mail.Rewards); err != nil {
		return fmt.Errorf("give rewards failed: %w", err)
	}

	// 标记为已领取
	mail.IsReceived = true
	s.SendMailDetail(mailID)

	return nil
}

// GetMail 获取邮件
func (s *MailSys) GetMail(mailID uint64) (*protocol.Mail, bool) {
	mail, ok := s.mails[mailID]
	return mail, ok
}

// GetMailCount 获取邮件数量
func (s *MailSys) GetMailCount() int {
	return len(s.mails)
}

// GetUnreadMailCount 获取未读邮件数量
func (s *MailSys) GetUnreadMailCount() int {
	count := 0
	for _, mail := range s.mails {
		if !mail.IsRead {
			count++
		}
	}
	return count
}

// DeleteAllReadMails 删除所有已读邮件
func (s *MailSys) DeleteAllReadMails() error {
	var toDelete []uint64
	for mailID, mail := range s.mails {
		if mail.IsRead && (!mail.HasReward || mail.IsReceived) {
			toDelete = append(toDelete, mailID)
		}
	}

	for _, mailID := range toDelete {
		delete(s.mails, mailID)
	}

	if len(toDelete) > 0 {
		s.SendMailList()
	}

	return nil
}

func handleReadMail(playerRole iface.IPlayerRole, msg *network.ClientMessage) error {
	// 解析请求
	var req struct {
		MailID uint64 `json:"mailId"`
	}
	if err := tool.JsonUnmarshal(msg.Data, &req); err != nil {
		log.Errorf("unmarshal read mail request failed: %v", err)
		return customerr.Wrap(err)
	}

	// 获取邮件系统
	mailSys := playerRole.GetSystem(custom_id.SysMail)
	if mailSys == nil {
		return customerr.NewCustomErr("邮件系统未找到")
	}

	ms, ok := mailSys.(*MailSys)
	if !ok {
		return customerr.NewCustomErr("邮件系统类型错误")
	}

	// 读取邮件
	if err := ms.ReadMail(req.MailID); err != nil {
		return customerr.Wrap(err)
	}

	return nil
}
func handleDeleteMail(playerRole iface.IPlayerRole, msg *network.ClientMessage) error {
	// 解析请求
	var req struct {
		MailID uint64 `json:"mailId"`
	}
	if err := tool.JsonUnmarshal(msg.Data, &req); err != nil {
		log.Errorf("unmarshal delete mail request failed: %v", err)
		return customerr.Wrap(err)
	}

	// 获取邮件系统
	mailSys := playerRole.GetSystem(custom_id.SysMail)
	if mailSys == nil {
		return customerr.NewCustomErr("邮件系统未找到")
	}

	ms, ok := mailSys.(*MailSys)
	if !ok {
		return customerr.NewCustomErr("邮件系统类型错误")
	}

	// 删除邮件
	if err := ms.DeleteMail(req.MailID); err != nil {
		return customerr.Wrap(err)
	}

	return nil
}
func handleReceiveMailReward(playerRole iface.IPlayerRole, msg *network.ClientMessage) error {
	// 解析请求
	var req struct {
		MailID uint64 `json:"mailId"`
	}
	if err := tool.JsonUnmarshal(msg.Data, &req); err != nil {
		log.Errorf("unmarshal receive mail reward request failed: %v", err)
		return customerr.Wrap(err)
	}

	// 获取邮件系统
	mailSys := playerRole.GetSystem(custom_id.SysMail)
	if mailSys == nil {
		return customerr.NewCustomErr("邮件系统未找到")
	}

	ms, ok := mailSys.(*MailSys)
	if !ok {
		return customerr.NewCustomErr("邮件系统类型错误")
	}

	// 领取邮件奖励
	if err := ms.ReceiveMailReward(req.MailID); err != nil {
		return customerr.Wrap(err)
	}

	return nil
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(custom_id.SysMail, func(role iface.IPlayerRole) iface.ISystem {
		return NewMailSys(role)
	})
	clientprotocol.Register(1, 6, handleReadMail)
	clientprotocol.Register(1, 7, handleDeleteMail)
	clientprotocol.Register(1, 8, handleReceiveMailReward)
}
