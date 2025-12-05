package mail

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

// ReadMailUseCase 标记邮件为已读
type ReadMailUseCase struct {
	playerRepo repository.PlayerRepository
}

func NewReadMailUseCase(playerRepo repository.PlayerRepository) *ReadMailUseCase {
	return &ReadMailUseCase{playerRepo: playerRepo}
}

func (uc *ReadMailUseCase) Execute(ctx context.Context, roleID uint64, mailID uint64) (*protocol.MailSt, error) {
	if roleID == 0 {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "未登录")
	}
	mailData, err := uc.playerRepo.GetMailData(ctx)
	if err != nil {
		return nil, customerr.Wrap(err)
	}

	var target *protocol.MailSt
	for _, m := range mailData.Mails {
		if m != nil && m.MailId == mailID {
			target = m
			break
		}
	}
	if target == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "邮件不存在")
	}
	if target.Status == 0 {
		target.Status = 1
	}
	return target, nil
}

// DeleteMailUseCase 删除单封邮件
type DeleteMailUseCase struct {
	playerRepo repository.PlayerRepository
}

func NewDeleteMailUseCase(playerRepo repository.PlayerRepository) *DeleteMailUseCase {
	return &DeleteMailUseCase{playerRepo: playerRepo}
}

func (uc *DeleteMailUseCase) Execute(ctx context.Context, roleID uint64, mailID uint64) error {
	if roleID == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "未登录")
	}
	mailData, err := uc.playerRepo.GetMailData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	out := mailData.Mails[:0]
	found := false
	for _, m := range mailData.Mails {
		if m != nil && m.MailId == mailID {
			found = true
			continue
		}
		out = append(out, m)
	}
	if !found {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "邮件不存在")
	}
	mailData.Mails = out
	return nil
}
