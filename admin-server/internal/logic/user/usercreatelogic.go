// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"time"

	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/errs"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type UserCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserCreateLogic {
	return &UserCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserCreateLogic) UserCreate(req *types.UserCreateReq) error {
	if req == nil || req.Username == "" || req.Password == "" {
		return errs.New(errs.CodeBadRequest, "用户名和密码不能为空")
	}

	userRepo := repository.NewUserRepository(l.svcCtx.Repository)
	// 检查用户名是否已存在
	_, err := userRepo.FindByUsername(l.ctx, req.Username)
	if err == nil {
		return errs.New(errs.CodeBadRequest, "用户名已存在")
	}

	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "密码加密失败", err)
	}

	user := model.AdminUser{
		Username:     req.Username,
		Nickname:     req.Nickname,
		PasswordHash: string(hash),
		Avatar:       req.Avatar,
		Signature:    req.Signature,
		DepartmentId: req.DepartmentId,
		Status:       req.Status,
	}

	if err := userRepo.Create(l.ctx, &user); err != nil {
		return errs.Wrap(errs.CodeInternalError, "创建用户失败", err)
	}

	// 创建用户成功后，自动初始化聊天相关数据
	if err := l.initChatForNewUser(user.Id); err != nil {
		// 记录错误但不影响用户创建
		logx.Errorf("初始化新用户聊天数据失败: %v", err)
	}

	return nil
}

// initChatForNewUser 为新用户初始化聊天数据
// 1. 将新用户加入默认企业群组（chat_id=1）
// 2. 为系统中所有其他用户创建与该新用户的私聊
func (l *UserCreateLogic) initChatForNewUser(newUserID uint64) error {
	chatRepo := repository.NewChatRepository(l.svcCtx.Repository)
	chatUserRepo := repository.NewChatUserRepository(l.svcCtx.Repository)
	userRepo := repository.NewUserRepository(l.svcCtx.Repository)

	now := time.Now().Unix()

	// 1. 将新用户加入默认企业群组（chat_id=1）
	defaultGroupChatID := uint64(1)
	groupChat, err := chatRepo.FindByID(l.ctx, defaultGroupChatID)
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询默认企业群组失败", err)
	}
	if groupChat.DeletedAt != 0 {
		logx.Infof("默认企业群组不存在或已删除，跳过加入群组操作")
	} else {
		// 检查用户是否已在群组中
		chatUsers, _ := chatUserRepo.FindByChatID(l.ctx, defaultGroupChatID)
		alreadyInGroup := false
		for _, cu := range chatUsers {
			if cu.UserId == newUserID {
				alreadyInGroup = true
				break
			}
		}
		if !alreadyInGroup {
			chatUser := &model.ChatUser{
				ChatId:    defaultGroupChatID,
				UserId:    newUserID,
				JoinedAt:  now,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := chatUserRepo.Create(l.ctx, chatUser); err != nil {
				logx.Errorf("将新用户加入默认企业群组失败: %v", err)
				// 继续执行，不中断流程
			}
		}
	}

	// 2. 为系统中所有其他用户创建与该新用户的私聊
	// 查询所有启用的用户（除了新用户自己）
	allUsers, _, err := userRepo.FindPage(l.ctx, 1, 10000, "")
	if err != nil {
		return errs.Wrap(errs.CodeInternalError, "查询用户列表失败", err)
	}

	for _, existingUser := range allUsers {
		// 跳过新用户自己
		if existingUser.Id == newUserID {
			continue
		}
		// 只处理启用的用户
		if existingUser.DeletedAt != 0 || existingUser.Status != 1 {
			continue
		}

		// 检查是否已存在私聊
		existingChat, err := chatRepo.FindPrivateChatByUserIDs(l.ctx, newUserID, existingUser.Id)
		if err == nil && existingChat != nil {
			// 私聊已存在，跳过
			continue
		}

		// 创建新的私聊
		privateChat := &model.Chat{
			Name:        "", // 私聊名称为空
			Type:        1,  // 类型：1 私聊
			Avatar:      "",
			Description: "",
			CreatedBy:   0, // 私聊创建人为0
			CreatedAt:   now,
			UpdatedAt:   now,
			DeletedAt:   0,
		}
		if err := chatRepo.Create(l.ctx, privateChat); err != nil {
			logx.Errorf("创建私聊失败 (新用户=%d, 现有用户=%d): %v", newUserID, existingUser.Id, err)
			continue
		}

		// 将两个用户都加入私聊
		chatUser1 := &model.ChatUser{
			ChatId:    privateChat.Id,
			UserId:    newUserID,
			JoinedAt:  now,
			CreatedAt: now,
			UpdatedAt: now,
		}
		chatUser2 := &model.ChatUser{
			ChatId:    privateChat.Id,
			UserId:    existingUser.Id,
			JoinedAt:  now,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := chatUserRepo.Create(l.ctx, chatUser1); err != nil {
			logx.Errorf("将新用户加入私聊失败: %v", err)
		}
		if err := chatUserRepo.Create(l.ctx, chatUser2); err != nil {
			logx.Errorf("将现有用户加入私聊失败: %v", err)
		}
	}

	return nil
}
