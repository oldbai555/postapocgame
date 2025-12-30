package repository

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"postapocgame/admin-server/internal/model"
)

type ChatRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.Chat, error)
	FindByUserID(ctx context.Context, userID uint64) ([]model.Chat, error)
	FindUsersByChatID(ctx context.Context, chatID uint64) ([]model.ChatUser, error)
	Create(ctx context.Context, chat *model.Chat) error
	Update(ctx context.Context, chat *model.Chat) error
	DeleteByID(ctx context.Context, id uint64) error
	FindPrivateChatByUserIDs(ctx context.Context, userID1, userID2 uint64) (*model.Chat, error)
}

type chatRepository struct {
	model model.ChatModel
	conn  sqlx.SqlConn
}

func NewChatRepository(repo *Repository) ChatRepository {
	return &chatRepository{model: repo.ChatModel, conn: repo.DB}
}

func (r *chatRepository) FindByID(ctx context.Context, id uint64) (*model.Chat, error) {
	return r.model.FindOne(ctx, id)
}

func (r *chatRepository) FindByUserID(ctx context.Context, userID uint64) ([]model.Chat, error) {
	// 通过chat_user关联表查询用户参与的所有聊天
	query := `
		SELECT c.* 
		FROM chat c
		INNER JOIN chat_user cu ON c.id = cu.chat_id
		WHERE cu.user_id = ? AND c.deleted_at = 0
		ORDER BY c.created_at DESC
	`
	var list []model.Chat
	err := r.conn.QueryRowsCtx(ctx, &list, query, userID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *chatRepository) FindUsersByChatID(ctx context.Context, chatID uint64) ([]model.ChatUser, error) {
	// 查询聊天中的所有用户
	query := `
		SELECT * 
		FROM chat_user
		WHERE chat_id = ?
		ORDER BY joined_at ASC
	`
	var list []model.ChatUser
	err := r.conn.QueryRowsCtx(ctx, &list, query, chatID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *chatRepository) Create(ctx context.Context, chat *model.Chat) error {
	_, err := r.model.Insert(ctx, chat)
	return err
}

func (r *chatRepository) Update(ctx context.Context, chat *model.Chat) error {
	return r.model.Update(ctx, chat)
}

func (r *chatRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.model.Delete(ctx, id)
}

func (r *chatRepository) FindPrivateChatByUserIDs(ctx context.Context, userID1, userID2 uint64) (*model.Chat, error) {
	// 查找两个用户之间的私聊（type=1）
	// 私聊必须包含且仅包含这两个用户
	query := `
		SELECT c.* 
		FROM chat c
		INNER JOIN chat_user cu1 ON c.id = cu1.chat_id AND cu1.user_id = ?
		INNER JOIN chat_user cu2 ON c.id = cu2.chat_id AND cu2.user_id = ?
		WHERE c.type = 1 AND c.deleted_at = 0
		GROUP BY c.id
		HAVING COUNT(DISTINCT cu.user_id) = 2
		LIMIT 1
	`
	// 修复：使用子查询来正确计算用户数
	query = `
		SELECT c.* 
		FROM chat c
		WHERE c.type = 1 AND c.deleted_at = 0
		AND EXISTS (SELECT 1 FROM chat_user cu1 WHERE cu1.chat_id = c.id AND cu1.user_id = ?)
		AND EXISTS (SELECT 1 FROM chat_user cu2 WHERE cu2.chat_id = c.id AND cu2.user_id = ?)
		AND (SELECT COUNT(*) FROM chat_user cu WHERE cu.chat_id = c.id) = 2
		LIMIT 1
	`
	var chat model.Chat
	err := r.conn.QueryRowCtx(ctx, &chat, query, userID1, userID2)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}
