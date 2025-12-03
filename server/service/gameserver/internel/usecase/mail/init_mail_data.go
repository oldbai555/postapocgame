package mail

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/domain/repository"
)

// InitMailDataUseCase 初始化邮件数据用例
// 负责邮件数据的初始化（邮件列表结构）
type InitMailDataUseCase struct {
	playerRepo repository.PlayerRepository
}

// NewInitMailDataUseCase 创建初始化邮件数据用例
func NewInitMailDataUseCase(playerRepo repository.PlayerRepository) *InitMailDataUseCase {
	return &InitMailDataUseCase{
		playerRepo: playerRepo,
	}
}

// Execute 执行初始化邮件数据用例
func (uc *InitMailDataUseCase) Execute(ctx context.Context, roleID uint64) error {
	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 确保 MailData 已初始化
	if binaryData.MailData == nil {
		binaryData.MailData = &protocol.SiMailData{
			Mails: make([]*protocol.MailSt, 0),
		}
	}
	if binaryData.MailData.Mails == nil {
		binaryData.MailData.Mails = make([]*protocol.MailSt, 0)
	}

	return nil
}
