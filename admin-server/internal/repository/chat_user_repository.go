package repository

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type ChatUserRepository interface {
	FindByChatID(ctx context.Context, chatID uint64) ([]model.ChatUser, error)
	FindByUserID(ctx context.Context, userID uint64) ([]model.ChatUser, error)
	Create(ctx context.Context, chatUser *model.ChatUser) error
	DeleteByChatIDAndUserID(ctx context.Context, chatID, userID uint64) error
}

type chatUserRepository struct {
	model model.ChatUserModel
	conn  sqlx.SqlConn
}

func NewChatUserRepository(repo *Repository) ChatUserRepository {
	return &chatUserRepository{model: repo.ChatUserModel, conn: repo.DB}
}

func (r *chatUserRepository) FindByChatID(ctx context.Context, chatID uint64) ([]model.ChatUser, error) {
	query := `SELECT * FROM chat_user WHERE chat_id = ? ORDER BY joined_at ASC`
	var list []model.ChatUser
	err := r.conn.QueryRowsCtx(ctx, &list, query, chatID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *chatUserRepository) FindByUserID(ctx context.Context, userID uint64) ([]model.ChatUser, error) {
	query := `SELECT * FROM chat_user WHERE user_id = ? ORDER BY joined_at ASC`
	var list []model.ChatUser
	err := r.conn.QueryRowsCtx(ctx, &list, query, userID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *chatUserRepository) Create(ctx context.Context, chatUser *model.ChatUser) error {
	_, err := r.model.Insert(ctx, chatUser)
	return err
}

func (r *chatUserRepository) DeleteByChatIDAndUserID(ctx context.Context, chatID, userID uint64) error {
	query := `DELETE FROM chat_user WHERE chat_id = ? AND user_id = ?`
	_, err := r.conn.ExecCtx(ctx, query, chatID, userID)
	return err
}
