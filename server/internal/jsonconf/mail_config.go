/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package jsonconf

import "postapocgame/server/internal/protocol"

// MailTemplateConfig 邮件模板配置
type MailTemplateConfig struct {
	TemplateId  uint32          `json:"templateId"`  // 模板Id
	Title       string          `json:"title"`       // 邮件标题
	Content     string          `json:"content"`     // 邮件内容
	Sender      string          `json:"sender"`      // 发送者
	Rewards     []protocol.Item `json:"rewards"`     // 邮件奖励
	ExpireHours uint32          `json:"expireHours"` // 过期时间(小时)
}
