package mail

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	"strconv"
	"strings"

	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

// ClaimAttachmentUseCase 领取单封邮件附件
type ClaimAttachmentUseCase struct {
	playerRepo repository.PlayerRepository
}

func NewClaimAttachmentUseCase(playerRepo repository.PlayerRepository) *ClaimAttachmentUseCase {
	return &ClaimAttachmentUseCase{playerRepo: playerRepo}
}

func (uc *ClaimAttachmentUseCase) Execute(ctx context.Context, roleID uint64, mailID uint64) (*protocol.MailSt, error) {
	if roleID == 0 {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "未登录")
	}
	mailData, err := uc.playerRepo.GetMailData(ctx)
	if err != nil {
		return nil, customerr.Wrap(err)
	}
	var mail *protocol.MailSt
	for _, m := range mailData.Mails {
		if m != nil && m.MailId == mailID {
			mail = m
			break
		}
	}
	if mail == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "邮件不存在")
	}
	if mail.Status == 2 {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "附件已领取")
	}

	// 解析附件字符串: id_count_bind|id_count_bind|...
	if mail.Files != "" {
		attachments := strings.Split(mail.Files, "|")
		rewards := make([]*jsonconf.ItemAmount, 0, len(attachments))
		for _, att := range attachments {
			if att == "" {
				continue
			}
			parts := strings.Split(att, "_")
			if len(parts) != 3 {
				continue
			}
			itemID, err1 := strconv.ParseUint(parts[0], 10, 32)
			count, err2 := strconv.ParseUint(parts[1], 10, 32)
			bind, err3 := strconv.ParseUint(parts[2], 10, 32)
			if err1 != nil || err2 != nil || err3 != nil {
				continue
			}
			rewards = append(rewards, &jsonconf.ItemAmount{
				ItemType: uint32(protocol.ItemType_ItemTypeMaterial),
				ItemId:   uint32(itemID),
				Count:    int64(count),
				Bind:     uint32(bind),
			})
		}
		_ = rewards // 具体发奖逻辑在更高层处理或复用现有 GrantRewardsByMail
	}

	mail.Status = 2
	return mail, nil
}
