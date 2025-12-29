package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type ChatOnlineUserRepository interface {
	FindByConnectionID(ctx context.Context, connectionID string) (*model.ChatOnlineUser, error)
	FindByUserID(ctx context.Context, userID uint64) ([]model.ChatOnlineUser, error)
	FindAll(ctx context.Context) ([]model.ChatOnlineUser, error)
	Create(ctx context.Context, user *model.ChatOnlineUser) error
	Update(ctx context.Context, user *model.ChatOnlineUser) error
	DeleteByConnectionID(ctx context.Context, connectionID string) error
	DeleteByUserID(ctx context.Context, userID uint64) error
}

type chatOnlineUserRepository struct {
	model model.ChatOnlineUserModel
	conn  sqlx.SqlConn
}

func NewChatOnlineUserRepository(repo *Repository) ChatOnlineUserRepository {
	return &chatOnlineUserRepository{model: repo.ChatOnlineUserModel, conn: repo.DB}
}

func (r *chatOnlineUserRepository) FindByConnectionID(ctx context.Context, connectionID string) (*model.ChatOnlineUser, error) {
	return r.model.FindOneByConnectionId(ctx, connectionID)
}

func (r *chatOnlineUserRepository) FindByUserID(ctx context.Context, userID uint64) ([]model.ChatOnlineUser, error) {
	// 需要扩展 Model 方法，暂时使用查询
	var list []model.ChatOnlineUser
	query := "SELECT * FROM `chat_online_user` WHERE user_id = ? ORDER BY created_at DESC"
	err := r.conn.QueryRowsCtx(ctx, &list, query, userID)
	return list, err
}

func (r *chatOnlineUserRepository) FindAll(ctx context.Context) ([]model.ChatOnlineUser, error) {
	var list []model.ChatOnlineUser
	query := "SELECT * FROM `chat_online_user` ORDER BY created_at DESC"
	err := r.conn.QueryRowsCtx(ctx, &list, query)
	return list, err
}

func (r *chatOnlineUserRepository) Create(ctx context.Context, user *model.ChatOnlineUser) error {
	_, err := r.model.Insert(ctx, user)
	return err
}

func (r *chatOnlineUserRepository) Update(ctx context.Context, user *model.ChatOnlineUser) error {
	return r.model.Update(ctx, user)
}

func (r *chatOnlineUserRepository) DeleteByConnectionID(ctx context.Context, connectionID string) error {
	user, err := r.model.FindOneByConnectionId(ctx, connectionID)
	if err != nil {
		return err
	}
	if user != nil {
		return r.model.Delete(ctx, user.Id)
	}
	return nil
}

func (r *chatOnlineUserRepository) DeleteByUserID(ctx context.Context, userID uint64) error {
	list, err := r.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, user := range list {
		if err := r.model.Delete(ctx, user.Id); err != nil {
			return err
		}
	}
	return nil
}
