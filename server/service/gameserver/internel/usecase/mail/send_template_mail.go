package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	maildomain "postapocgame/server/service/gameserver/internel/domain/mail"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// SendTemplateMailUseCase 使用模板发送邮件
type SendTemplateMailUseCase struct {
	playerRepo    repository.PlayerRepository
	configManager interfaces.ConfigManager
}

func NewSendTemplateMailUseCase(playerRepo repository.PlayerRepository, cfg interfaces.ConfigManager) *SendTemplateMailUseCase {
	return &SendTemplateMailUseCase{
		playerRepo:    playerRepo,
		configManager: cfg,
	}
}

// Execute 向单个角色发送模板邮件
func (uc *SendTemplateMailUseCase) Execute(ctx context.Context, roleID uint64, templateID uint32, args map[string]interface{}) error {
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

	rawCfg, ok := uc.configManager.GetMailTemplateConfig(templateID)
	if !ok || rawCfg == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "mail template config not found: %d", templateID)
	}
	cfg, ok := rawCfg.(*jsonconf.MailTemplateConfig)
	if !ok || cfg == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "invalid mail template config type: %T", rawCfg)
	}

	mailID := maildomain.NextMailID(mailData)

	argsJSON := "{}"
	if len(args) > 0 {
		if b, e := json.Marshal(args); e == nil {
			argsJSON = string(b)
		}
	}

	mail := &protocol.MailSt{
		MailId:   mailID,
		ConfId:   templateID,
		Status:   0,
		CreateAt: uint32(servertime.Now().Unix()),
		Args:     argsJSON,
		Title:    cfg.Title,
		Content:  cfg.Content,
	}

	if len(cfg.Rewards) > 0 {
		attachments := make([]string, 0, len(cfg.Rewards))
		for _, reward := range cfg.Rewards {
			attachments = append(attachments, fmt.Sprintf("%d_%d_%d", reward.ItemId, reward.Count, 1))
		}
		mail.Files = strings.Join(attachments, "|")
	}

	mailData.Mails = append(mailData.Mails, mail)
	return nil
}
