package mail

import (
	"context"
	"fmt"
	"strings"

	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	maildomain "postapocgame/server/service/gameserver/internel/domain/mail"
	"postapocgame/server/service/gameserver/internel/domain/repository"
)

// SendCustomMailUseCase 发送自定义邮件
type SendCustomMailUseCase struct {
	playerRepo repository.PlayerRepository
}

func NewSendCustomMailUseCase(playerRepo repository.PlayerRepository) *SendCustomMailUseCase {
	return &SendCustomMailUseCase{playerRepo: playerRepo}
}

func (uc *SendCustomMailUseCase) Execute(ctx context.Context, roleID uint64, title, content string, rewards []*jsonconf.ItemSt) error {
	if roleID == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "未登录")
	}
	bd, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return customerr.Wrap(err)
	}
	mailData := maildomain.EnsureMailData(bd)
	if mailData == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "邮件数据异常")
	}

	mailID := maildomain.NextMailID(mailData)
	mail := &protocol.MailSt{
		MailId:   mailID,
		ConfId:   0,
		Status:   0,
		CreateAt: uint32(servertime.Now().Unix()),
		Args:     "{}",
		Title:    title,
		Content:  content,
	}

	if len(rewards) > 0 {
		attachments := make([]string, 0, len(rewards))
		for _, reward := range rewards {
			attachments = append(attachments, fmt.Sprintf("%d_%d_%d", reward.ItemId, reward.Count, 1))
		}
		mail.Files = strings.Join(attachments, "|")
	}

	mailData.Mails = append(mailData.Mails, mail)
	return nil
}
