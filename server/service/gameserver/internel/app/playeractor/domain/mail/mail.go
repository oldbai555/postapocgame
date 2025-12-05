package mail

import (
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
)

// EnsureMailData 确保 MailData 初始化
func EnsureMailData(binaryData *protocol.PlayerRoleBinaryData) *protocol.SiMailData {
	if binaryData == nil {
		return nil
	}
	if binaryData.MailData == nil {
		binaryData.MailData = &protocol.SiMailData{
			Mails: make([]*protocol.MailSt, 0),
		}
	}
	if binaryData.MailData.Mails == nil {
		binaryData.MailData.Mails = make([]*protocol.MailSt, 0)
	}
	return binaryData.MailData
}

// NextMailID 在现有数据上生成下一个 MailId
func NextMailID(mailData *protocol.SiMailData) uint64 {
	var maxID uint64 = 0
	if mailData == nil {
		return 1
	}
	for _, m := range mailData.Mails {
		if m != nil && m.MailId > maxID {
			maxID = m.MailId
		}
	}
	return maxID + 1
}

// CleanExpiredMails 清理过期邮件（基于模板过期时间）
// 这里不依赖配置，调用方应在外侧根据配置判断是否过期
func CleanExpiredMails(mailData *protocol.SiMailData, isExpired func(mail *protocol.MailSt, now int64) bool) {
	if mailData == nil || mailData.Mails == nil || isExpired == nil {
		return
	}
	now := servertime.Now().Unix()
	valid := mailData.Mails[:0]
	for _, m := range mailData.Mails {
		if m == nil || isExpired(m, now) {
			continue
		}
		valid = append(valid, m)
	}
	mailData.Mails = valid
}
